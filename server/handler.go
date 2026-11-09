package main

import (
	"bytes"
	"log"
	"net/http"
	"net/netip"
	"net/url"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

// handler serves the site. It writes no access log: nothing about a request is logged, ever.
type handler struct {
	site      *site
	trusted   []netip.Prefix // proxies whose X-Forwarded-For entry is believed (TRUSTED_PROXIES)
	analytics *sender        // nil when analytics is off
}

// recorder notes the status a response went out with.
type recorder struct {
	http.ResponseWriter
	status int
}

func (r *recorder) WriteHeader(status int) {
	if r.status == 0 {
		r.status = status
	}
	r.ResponseWriter.WriteHeader(status)
}

func (r *recorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	return r.ResponseWriter.Write(b)
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rec := &recorder{ResponseWriter: w}
	defer func() {
		if v := recover(); v != nil {
			if v == http.ErrAbortHandler {
				panic(v)
			}
			// The stack only: the panic value or net/http's own panic log could hold request data.
			log.Printf("server: panic while serving a request\n%s", debug.Stack())
			if rec.status == 0 {
				http.Error(rec, "Internal Server Error", http.StatusInternalServerError)
			}
		}
	}()

	page := h.serve(rec, r)

	// The response has been written; the page view goes to the analytics endpoint in the background.
	if h.analytics != nil && page && countable(r, rec.status, rec.Header().Get("Content-Type")) {
		if ip, ok := clientIP(r, h.trusted); ok {
			h.analytics.send(hit{
				Site:      h.analytics.site,
				URL:       r.URL.RequestURI(),
				Referrer:  r.Header.Get("Referer"),
				IP:        ip.String(),
				UserAgent: r.Header.Get("User-Agent"),
			})
		}
	}
}

// countable is the sender rule of the analytics contract: a GET that ended in 200 with an HTML body, and not a
// prefetch or prerender.
func countable(r *http.Request, status int, contentType string) bool {
	if r.Method != http.MethodGet || status != http.StatusOK {
		return false
	}
	if !strings.HasPrefix(strings.ToLower(contentType), "text/html") {
		return false
	}
	for _, name := range []string{"Sec-Purpose", "Purpose"} {
		for _, v := range r.Header.Values(name) {
			v = strings.ToLower(v)
			if strings.Contains(v, "prefetch") || strings.Contains(v, "prerender") {
				return false
			}
		}
	}
	return true
}

// serve answers the request and reports whether it served a page (an HTML file of the site, not a redirect, a
// meta-refresh page, the 404 page or the health check).
func (h *handler) serve(w http.ResponseWriter, r *http.Request) bool {
	if r.URL.Path == "/healthz" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(http.StatusOK)
		if r.Method != http.MethodHead {
			w.Write([]byte("ok\n"))
		}
		return false
	}
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return false
	}

	// _redirects first: on Cloudflare Pages a redirect wins over a file and over _headers.
	if loc, status, ok := findRedirect(h.site.redirects, r.URL.EscapedPath(), r.URL.RawQuery); ok {
		w.Header().Set("Location", loc)
		w.WriteHeader(status)
		return false
	}

	p := r.URL.Path
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	f, canonical := h.site.resolve(p)
	if f == nil {
		applyHeaders(w.Header(), h.site.headers, p)
		writeFile(w, r, h.site.notFound, http.StatusNotFound)
		return false
	}
	if canonical != p {
		w.Header().Set("Location", withQuery((&url.URL{Path: canonical}).EscapedPath(), r.URL.RawQuery))
		w.WriteHeader(http.StatusPermanentRedirect)
		return false
	}
	applyHeaders(w.Header(), h.site.headers, p)
	writeFile(w, r, f, http.StatusOK)
	return f.page
}

// writeFile sends a file, gzip-compressed when the client takes it. With status 200, conditional (If-None-Match) and
// range requests are answered by http.ServeContent.
func writeFile(w http.ResponseWriter, r *http.Request, f *file, status int) {
	body, etag := f.body, f.etag
	if f.gz != nil {
		w.Header().Add("Vary", "Accept-Encoding")
		if acceptsGzip(r.Header) {
			body, etag = f.gz, strings.TrimSuffix(f.etag, `"`)+`-gz"`
			w.Header().Set("Content-Encoding", "gzip")
		}
	}
	w.Header().Set("Content-Type", f.ctype)
	if status == http.StatusOK {
		w.Header().Set("ETag", etag)
		http.ServeContent(w, r, "", time.Time{}, bytes.NewReader(body))
		return
	}
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.WriteHeader(status)
	if r.Method != http.MethodHead {
		w.Write(body)
	}
}
