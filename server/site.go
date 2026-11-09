package main

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

// file is one file of the built site, held in memory (the whole site is under a megabyte).
type file struct {
	body  []byte
	gz    []byte // gzip-compressed body, or nil when compressing does not pay
	etag  string
	ctype string
	page  bool // an HTML page a visitor reads: counted by analytics. Not the meta-refresh pages (redirects).
}

type site struct {
	files     map[string]*file // by URL path: "/index.html", "/convert/html/index.html", "/css/site.123.css"
	notFound  *file            // 404.html
	headers   []headerRule
	redirects []redirectRule
}

var contentTypes = map[string]string{
	".html":        "text/html; charset=utf-8",
	".css":         "text/css; charset=utf-8",
	".js":          "text/javascript; charset=utf-8",
	".json":        "application/json",
	".xml":         "application/xml",
	".txt":         "text/plain; charset=utf-8",
	".md":          "text/markdown; charset=utf-8", // every page as Markdown, for AI assistants (hugo.toml)
	".svg":         "image/svg+xml",
	".ico":         "image/x-icon",
	".png":         "image/png",
	".jpg":         "image/jpeg",
	".jpeg":        "image/jpeg",
	".gif":         "image/gif",
	".webp":        "image/webp",
	".avif":        "image/avif",
	".woff2":       "font/woff2",
	".webmanifest": "application/manifest+json",
	".pdf":         "application/pdf",
}

// Hugo's alias pages (layouts/alias.html) and stable short URLs (layouts/_default/redirect.html) only send the reader
// on with a meta refresh. They are redirects, not page views.
var metaRefresh = regexp.MustCompile(`(?i)<meta[^>]+http-equiv=["']?refresh`)

// loadSite reads the built site (Hugo's public/) and its _headers and _redirects.
func loadSite(dir string) (*site, error) {
	s := &site{files: map[string]*file{}}
	for name, parse := range map[string]func(*os.File) error{
		"_headers": func(f *os.File) (err error) {
			s.headers, err = parseHeaders(f)
			return err
		},
		"_redirects": func(f *os.File) (err error) {
			s.redirects, err = parseRedirects(f)
			return err
		},
	} {
		f, err := os.Open(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		err = parse(f)
		f.Close()
		if err != nil {
			return nil, err
		}
	}

	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			return err
		}
		urlPath := "/" + filepath.ToSlash(rel)
		if urlPath == "/_headers" || urlPath == "/_redirects" || strings.HasPrefix(path.Base(urlPath), ".") {
			return nil
		}
		body, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		s.files[urlPath] = newFile(urlPath, body)
		return nil
	})
	if err != nil {
		return nil, err
	}
	if s.files["/index.html"] == nil {
		return nil, fmt.Errorf("%s has no index.html: build the site first (scripts/build.sh)", dir)
	}
	s.notFound = s.files["/404.html"]
	if s.notFound == nil {
		return nil, fmt.Errorf("%s has no 404.html", dir)
	}
	s.notFound.page = false
	// Only served as the answer to unknown paths, always with status 404.
	delete(s.files, "/404.html")
	return s, nil
}

func newFile(urlPath string, body []byte) *file {
	ext := strings.ToLower(path.Ext(urlPath))
	ctype := contentTypes[ext]
	if ctype == "" {
		ctype = mime.TypeByExtension(ext)
	}
	if ctype == "" {
		ctype = http.DetectContentType(body)
	}
	sum := sha256.Sum256(body)
	f := &file{
		body:  body,
		etag:  `"` + hex.EncodeToString(sum[:12]) + `"`,
		ctype: ctype,
		page:  ext == ".html" && !metaRefresh.Match(body),
	}
	if compressible(ctype) && len(body) >= 512 {
		var buf bytes.Buffer
		zw, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression)
		zw.Write(body)
		zw.Close()
		if buf.Len() < len(body)*9/10 {
			f.gz = buf.Bytes()
		}
	}
	return f
}

func compressible(ctype string) bool {
	return strings.HasPrefix(ctype, "text/") || strings.Contains(ctype, "json") || strings.Contains(ctype, "xml") ||
		strings.Contains(ctype, "javascript") || strings.HasPrefix(ctype, "image/svg") ||
		strings.HasPrefix(ctype, "image/x-icon")
}

// resolve finds the file for a request path, as Cloudflare Pages did for a Hugo site: "/dir/" serves
// dir/index.html, "/dir" and "/dir/index.html" redirect to "/dir/" (canonical differs from the path asked for), and
// unknown paths give nil.
func (s *site) resolve(p string) (f *file, canonical string) {
	clean := path.Clean("/" + p)
	if clean == "/" {
		return s.files["/index.html"], "/"
	}
	if strings.HasSuffix(clean, "/index.html") && s.files[clean] != nil {
		return s.files[clean], strings.TrimSuffix(clean, "index.html")
	}
	if !strings.HasSuffix(p, "/") {
		if f := s.files[clean]; f != nil {
			return f, clean
		}
	}
	if f := s.files[clean+"/index.html"]; f != nil {
		return f, clean + "/"
	}
	return nil, ""
}

// acceptsGzip reports whether the Accept-Encoding header allows gzip (and does not refuse it with q=0).
func acceptsGzip(h http.Header) bool {
	ok := false
	for _, v := range h.Values("Accept-Encoding") {
		for _, part := range strings.Split(v, ",") {
			name, params, _ := strings.Cut(strings.TrimSpace(part), ";")
			name = strings.ToLower(strings.TrimSpace(name))
			if name != "gzip" && name != "*" {
				continue
			}
			refused := false
			for _, param := range strings.Split(params, ";") {
				k, v, _ := strings.Cut(strings.TrimSpace(param), "=")
				if strings.EqualFold(strings.TrimSpace(k), "q") && strings.Trim(strings.TrimSpace(v), "0.") == "" {
					refused = true
				}
			}
			if name == "gzip" {
				return !refused
			}
			ok = !refused
		}
	}
	return ok
}
