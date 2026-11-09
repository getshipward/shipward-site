package main

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const pageHTML = `<!doctype html><html><head><title>HTML to PDF</title></head><body><h1>HTML to PDF</h1>` +
	`<p>Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore ` +
	`magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo ` +
	`consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla ` +
	`pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim.</p>` +
	`</body></html>`

const testRedirects = `/docs /  301
/docs/endpoints /#routes 301
/docs/* /:splat 301
/chromium/html /convert/html/ 301
/chromium/html/ /convert/html/ 301
`

// testSite writes a small built site with the real _headers and test redirects.
func testSite(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		"index.html":              "<!doctype html><title>Home</title><p>Home</p>",
		"convert/html/index.html": pageHTML,
		"convert/html/index.md":   "# HTML to PDF\n\nSend an index.html.\n",
		"llms.txt":                "# Shipward\n",
		"keys/index.html":         `<!doctype html><html><head><meta name=robots content="noindex"><meta http-equiv=refresh content="0; url=/authentication/"></head></html>`,
		"errors/index.html":       `<!doctype html><meta http-equiv="refresh" content="0; url=/limits/#errors">`,
		"404.html":                "<!doctype html><title>Page not found</title><p>Not here.</p>",
		"css/site.abc123.css":     "body{color:#000}",
		"index.json":              `[{"title":"Home"}]`,
		"favicon.svg":             `<svg xmlns="http://www.w3.org/2000/svg"></svg>`,
	}
	for name, body := range files {
		p := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// The real _headers, plus a rule the Markdown test needs. The real _redirects has no rules yet, so the redirect
	// tests use their own (the kinds of rule a page move adds).
	headers, err := os.ReadFile(filepath.Join("..", "static", "_headers"))
	if err != nil {
		t.Fatal(err)
	}
	headers = append(headers, []byte("\n/*/index.md\n  X-Robots-Tag: noindex\n")...)
	if err := os.WriteFile(filepath.Join(dir, "_headers"), headers, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "_redirects"), []byte(testRedirects), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func testHandler(t *testing.T) *handler {
	t.Helper()
	s, err := loadSite(testSite(t))
	if err != nil {
		t.Fatal(err)
	}
	trusted, err := parseTrusted("10.42.0.0/16")
	if err != nil {
		t.Fatal(err)
	}
	return &handler{site: s, trusted: trusted}
}

func do(h http.Handler, method, target string, header map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, nil)
	req.RemoteAddr = "10.42.3.7:41234"
	req.Header.Set("X-Forwarded-For", "198.51.100.1, 203.0.113.9")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) Firefox/140.0")
	req.Header.Set("Referer", "https://www.example.org/some/page?q=secret")
	for k, v := range header {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestServesPagesWithHeaders(t *testing.T) {
	h := testHandler(t)
	rec := do(h, "GET", "/convert/html/", nil)
	if rec.Code != 200 || rec.Body.String() != pageHTML {
		t.Fatalf("got %d %q", rec.Code, rec.Body.String())
	}
	for name, want := range map[string]string{
		"Content-Type":              "text/html; charset=utf-8",
		"X-Frame-Options":           "DENY",
		"Strict-Transport-Security": "max-age=31536000",
		"Cache-Control":             "public, max-age=0, must-revalidate",
	} {
		if got := rec.Header().Get(name); got != want {
			t.Errorf("%s: %q, want %q", name, got, want)
		}
	}
	if !strings.HasPrefix(rec.Header().Get("Content-Security-Policy"), "default-src 'none'") {
		t.Errorf("CSP: %q", rec.Header().Get("Content-Security-Policy"))
	}

	rec = do(h, "GET", "/css/site.abc123.css", nil)
	if rec.Code != 200 || rec.Header().Get("Cache-Control") != "public, max-age=31536000, immutable" ||
		rec.Header().Get("Content-Type") != "text/css; charset=utf-8" {
		t.Errorf("stylesheet: %d %v", rec.Code, rec.Header())
	}
	rec = do(h, "GET", "/", nil)
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "Home") {
		t.Errorf("home: %d", rec.Code)
	}
}

func TestServesMarkdownForAI(t *testing.T) {
	h := testHandler(t)
	for path, want := range map[string]string{
		"/convert/html/index.md": "text/markdown; charset=utf-8",
		"/llms.txt":              "text/plain; charset=utf-8",
	} {
		rec := do(h, "GET", path, nil)
		if rec.Code != 200 || rec.Header().Get("Content-Type") != want {
			t.Errorf("%s: %d %q, want %q", path, rec.Code, rec.Header().Get("Content-Type"), want)
		}
		if got := rec.Header().Get("X-Robots-Tag"); path != "/llms.txt" && got != "noindex" {
			t.Errorf("%s: X-Robots-Tag %q, want noindex (the HTML page is the one to index)", path, got)
		}
	}
}

func TestRedirects(t *testing.T) {
	h := testHandler(t)
	for target, want := range map[string]struct {
		status int
		loc    string
	}{
		"/chromium/html":             {301, "/convert/html/"},
		"/chromium/html/":            {301, "/convert/html/"},
		"/docs":                      {301, "/"},
		"/docs/endpoints":            {301, "/#routes"},
		"/docs/convert/html/?a=1":    {301, "/convert/html/?a=1"},
		"/docs//evil.example":        {301, "/evil.example"},
		"/convert/html":              {308, "/convert/html/"},
		"/convert/html?utm_source=x": {308, "/convert/html/?utm_source=x"},
		"/convert/html/index.html":   {308, "/convert/html/"},
		"/index.html":                {308, "/"},
		"/convert//html/":            {308, "/convert/html/"},
	} {
		rec := do(h, "GET", target, nil)
		if rec.Code != want.status || rec.Header().Get("Location") != want.loc {
			t.Errorf("%s: %d %q, want %d %q", target, rec.Code, rec.Header().Get("Location"), want.status, want.loc)
		}
	}
}

func TestNotFound(t *testing.T) {
	h := testHandler(t)
	for _, target := range []string{"/nope", "/nope/", "/404.html", "/css/site.abc123.css/", "/_headers", "/_redirects"} {
		rec := do(h, "GET", target, nil)
		if rec.Code != 404 || !strings.Contains(rec.Body.String(), "Not here.") {
			t.Errorf("%s: %d %q", target, rec.Code, rec.Body.String())
		}
		if rec.Header().Get("X-Frame-Options") != "DENY" || rec.Header().Get("Content-Type") != "text/html; charset=utf-8" {
			t.Errorf("%s: headers %v", target, rec.Header())
		}
	}
	rec := do(h, "HEAD", "/nope", nil)
	if rec.Code != 404 || rec.Body.Len() != 0 {
		t.Errorf("HEAD 404: %d, %d bytes", rec.Code, rec.Body.Len())
	}
}

func TestHealthzMethodsAndConditional(t *testing.T) {
	h := testHandler(t)
	if rec := do(h, "GET", "/healthz", nil); rec.Code != 200 || rec.Body.String() != "ok\n" {
		t.Errorf("healthz: %d %q", rec.Code, rec.Body.String())
	}
	if rec := do(h, "POST", "/convert/html/", nil); rec.Code != 405 || rec.Header().Get("Allow") != "GET, HEAD" {
		t.Errorf("POST: %d", rec.Code)
	}
	rec := do(h, "GET", "/convert/html/", nil)
	etag := rec.Header().Get("ETag")
	if etag == "" {
		t.Fatal("no ETag")
	}
	if rec := do(h, "GET", "/convert/html/", map[string]string{"If-None-Match": etag}); rec.Code != 304 {
		t.Errorf("If-None-Match: %d", rec.Code)
	}
	if rec := do(h, "HEAD", "/convert/html/", nil); rec.Code != 200 || rec.Body.Len() != 0 {
		t.Errorf("HEAD: %d, %d bytes", rec.Code, rec.Body.Len())
	}
}

func TestGzip(t *testing.T) {
	h := testHandler(t)
	rec := do(h, "GET", "/convert/html/", map[string]string{"Accept-Encoding": "br, gzip"})
	if rec.Header().Get("Content-Encoding") != "gzip" || rec.Header().Get("Vary") != "Accept-Encoding" {
		t.Fatalf("headers: %v", rec.Header())
	}
	zr, err := gzip.NewReader(bytes.NewReader(rec.Body.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(zr)
	if string(body) != pageHTML {
		t.Errorf("body after gunzip differs")
	}
	rec = do(h, "GET", "/convert/html/", map[string]string{"Accept-Encoding": "gzip;q=0, *"})
	if rec.Header().Get("Content-Encoding") != "" || rec.Body.String() != pageHTML {
		t.Errorf("gzip;q=0 still compressed")
	}
}

func TestAcceptsGzip(t *testing.T) {
	for v, want := range map[string]bool{
		"":                 false,
		"gzip":             true,
		"GZIP":             true,
		"br, gzip;q=0.5":   true,
		"gzip;q=0":         false,
		"gzip; q=0.000":    false,
		"*":                true,
		"*;q=0":            false,
		"deflate, br":      false,
		"gzip;q=0, *;q=1":  false,
		"identity, *;q=.1": true,
	} {
		h := http.Header{}
		if v != "" {
			h.Set("Accept-Encoding", v)
		}
		if got := acceptsGzip(h); got != want {
			t.Errorf("%q: %v, want %v", v, got, want)
		}
	}
}

// The server never sets a cookie, on any kind of response.
func TestNoSetCookieEver(t *testing.T) {
	h := testHandler(t)
	h.analytics = newSender("http://127.0.0.1:1/hits", "t", "shipward.dev")
	h.analytics.logf = func(string, ...any) {}
	for _, target := range []string{"/", "/convert/html/", "/convert/html", "/chromium/html", "/nope", "/healthz",
		"/css/site.abc123.css", "/keys/"} {
		for _, method := range []string{"GET", "HEAD", "POST"} {
			if rec := do(h, method, target, nil); len(rec.Header().Values("Set-Cookie")) > 0 {
				t.Errorf("%s %s set a cookie", method, target)
			}
		}
	}
}

// The real build, when there is one (scripts/build.sh): it loads, every page is served, and the aliases are not
// counted as pages.
func TestBuiltSite(t *testing.T) {
	if _, err := os.Stat("../public/index.html"); err != nil {
		t.Skip("no ../public: run scripts/build.sh to include the real site")
	}
	s, err := loadSite("../public")
	if err != nil {
		t.Fatal(err)
	}
	h := &handler{site: s}
	pages := 0
	for p, f := range s.files {
		if !strings.HasSuffix(p, "/index.html") {
			continue
		}
		dir := strings.TrimSuffix(p, "index.html")
		rec := do(h, "GET", dir, nil)
		if rec.Code != 200 {
			t.Errorf("%s: %d", dir, rec.Code)
		}
		if f.page {
			pages++
		}
	}
	if pages == 0 {
		t.Error("no pages")
	}
	if rec := do(h, "GET", "/does-not-exist/", nil); rec.Code != 404 || !strings.Contains(rec.Body.String(), "Page not found") {
		t.Errorf("404: %d", rec.Code)
	}
}
