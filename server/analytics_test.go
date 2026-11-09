package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"sync"
	"testing"
	"time"
)

// fakeIngest records what the sender posts.
type fakeIngest struct {
	mu     sync.Mutex
	hits   []hit
	auth   []string
	ctypes []string
	status int
	delay  time.Duration
}

func (c *fakeIngest) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if c.delay > 0 {
		time.Sleep(c.delay)
	}
	var h hit
	body, _ := io.ReadAll(r.Body)
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&h); err != nil || r.Method != http.MethodPost || r.URL.Path != "/internal/analytics/v1/hits" {
		w.WriteHeader(422)
		return
	}
	c.mu.Lock()
	c.hits = append(c.hits, h)
	c.auth = append(c.auth, r.Header.Get("Authorization"))
	c.ctypes = append(c.ctypes, r.Header.Get("Content-Type"))
	status := c.status
	c.mu.Unlock()
	if status == 0 {
		status = http.StatusNoContent
	}
	w.WriteHeader(status)
}

func (c *fakeIngest) take() []hit {
	c.mu.Lock()
	defer c.mu.Unlock()
	h := c.hits
	c.hits = nil
	return h
}

// logLines collects the sender's log lines, and anything written to the standard logger meanwhile.
type logLines struct {
	mu    sync.Mutex
	lines []string
}

func (l *logLines) logf(format string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.lines = append(l.lines, fmt.Sprintf(format, args...))
}

func (l *logLines) Write(p []byte) (int, error) {
	l.logf("%s", strings.TrimRight(string(p), "\n"))
	return len(p), nil
}

func (l *logLines) all() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]string(nil), l.lines...)
}

func withAnalytics(t *testing.T, ingest http.Handler) (*handler, *logLines) {
	t.Helper()
	h := testHandler(t)
	srv := httptest.NewServer(ingest)
	t.Cleanup(srv.Close)
	h.analytics = newSender(srv.URL+"/internal/analytics/v1/hits", "tok_123", "shipward.dev")
	logs := &logLines{}
	h.analytics.logf = logs.logf
	prev := log.Writer()
	log.SetOutput(logs)
	t.Cleanup(func() { log.SetOutput(prev) })
	return h, logs
}

func settle(h *handler) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	h.analytics.wait(ctx)
}

func TestAnalyticsDecisionTable(t *testing.T) {
	ingest := &fakeIngest{}
	h, logs := withAnalytics(t, ingest)
	page := testHandler(t) // the same site without analytics, for comparing response headers

	cases := []struct {
		name, method, target string
		header               map[string]string
		sent                 bool
	}{
		{"page", "GET", "/convert/html/?utm_source=news&a=1", nil, true},
		{"home", "GET", "/", nil, true},
		{"HEAD", "HEAD", "/convert/html/", nil, false},
		{"POST", "POST", "/convert/html/", nil, false},
		{"Sec-Purpose prefetch", "GET", "/convert/html/", map[string]string{"Sec-Purpose": "prefetch"}, false},
		{"Sec-Purpose prerender", "GET", "/convert/html/", map[string]string{"Sec-Purpose": "prefetch;prerender"}, false},
		{"Purpose prefetch", "GET", "/convert/html/", map[string]string{"Purpose": "prefetch"}, false},
		{"stylesheet", "GET", "/css/site.abc123.css", nil, false},
		{"search index", "GET", "/index.json", nil, false},
		{"icon", "GET", "/favicon.svg", nil, false},
		{"404", "GET", "/nope/", nil, false},
		{"404 page itself", "GET", "/404.html", nil, false},
		{"_redirects rule", "GET", "/chromium/html/", nil, false},
		{"trailing slash redirect", "GET", "/convert/html", nil, false},
		{"meta-refresh alias", "GET", "/keys/", nil, false},
		{"meta-refresh short URL", "GET", "/errors/", nil, false},
		{"health check", "GET", "/healthz", nil, false},
		{"range", "GET", "/convert/html/", map[string]string{"Range": "bytes=0-10"}, false},
	}
	for _, tc := range cases {
		rec := do(h, tc.method, tc.target, tc.header)
		settle(h)
		got := ingest.take()
		if tc.sent != (len(got) == 1) || len(got) > 1 {
			t.Errorf("%s: %d hits sent, want sent=%v", tc.name, len(got), tc.sent)
		}
		// Analytics adds no header and no cookie: the response is the same as without it.
		plain := do(page, tc.method, tc.target, tc.header)
		if fmt.Sprint(rec.Header()) != fmt.Sprint(plain.Header()) || rec.Code != plain.Code {
			t.Errorf("%s: response differs with analytics on:\n%v\n%v", tc.name, rec.Header(), plain.Header())
		}
	}

	// 304: a revalidation is not a new view.
	etag := do(page, "GET", "/convert/html/", nil).Header().Get("ETag")
	do(h, "GET", "/convert/html/", map[string]string{"If-None-Match": etag})
	settle(h)
	if got := ingest.take(); len(got) != 0 {
		t.Errorf("304: %d hits sent", len(got))
	}

	// The payload of a page view.
	do(h, "GET", "/convert/html/?utm_source=news&a=1", nil)
	settle(h)
	got := ingest.take()
	if len(got) != 1 {
		t.Fatalf("%d hits", len(got))
	}
	want := hit{
		Site:      "shipward.dev",
		URL:       "/convert/html/?utm_source=news&a=1",
		Referrer:  "https://www.example.org/some/page?q=secret",
		IP:        "203.0.113.9",
		UserAgent: "Mozilla/5.0 (X11; Linux x86_64) Firefox/140.0",
	}
	if got[0] != want {
		t.Errorf("payload %+v, want %+v", got[0], want)
	}
	if ingest.auth[0] != "Bearer tok_123" || ingest.ctypes[0] != "application/json" {
		t.Errorf("headers: %q %q", ingest.auth[0], ingest.ctypes[0])
	}
	if lines := logs.all(); len(lines) != 0 {
		t.Errorf("logged on success: %q", lines)
	}
}

// Failures log one fixed line, with nothing from the visitor in it.
func TestAnalyticsFailureLogs(t *testing.T) {
	secret := []string{"203.0.113.9", "198.51.100.1", "Firefox", "Mozilla", "example.org", "q=secret", "utm_source",
		"convert", "tok_123"}
	check := func(name string, logs *logLines, want string) {
		t.Helper()
		lines := logs.all()
		if len(lines) != 1 || lines[0] != want {
			t.Errorf("%s: logged %q, want [%q]", name, lines, want)
		}
		for _, line := range lines {
			for _, s := range secret {
				if strings.Contains(line, s) {
					t.Errorf("%s: log line %q contains %q", name, line, s)
				}
			}
		}
	}

	h, logs := withAnalytics(t, &fakeIngest{status: 500})
	do(h, "GET", "/convert/html/?utm_source=news", nil)
	settle(h)
	check("status", logs, "analytics: send failed (status 500)")

	h, logs = withAnalytics(t, &fakeIngest{delay: 300 * time.Millisecond})
	h.analytics.timeout = 50 * time.Millisecond
	do(h, "GET", "/convert/html/?utm_source=news", nil)
	settle(h)
	check("timeout", logs, "analytics: send failed (timeout)")

	h, logs = withAnalytics(t, &fakeIngest{})
	closed := httptest.NewServer(http.NotFoundHandler())
	closed.Close()
	h.analytics.url = closed.URL + "/internal/analytics/v1/hits?ip=203.0.113.9"
	do(h, "GET", "/convert/html/?utm_source=news", nil)
	settle(h)
	check("network", logs, "analytics: send failed (network)")
}

// With every slot busy a page view is dropped at once, and the response does not wait.
func TestAnalyticsDropsWhenBusy(t *testing.T) {
	release := make(chan struct{})
	var mu sync.Mutex
	calls := 0
	h, logs := withAnalytics(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		calls++
		mu.Unlock()
		<-release
		w.WriteHeader(204)
	}))
	h.analytics.slots = make(chan struct{}, 2)
	start := time.Now()
	for range 5 {
		if rec := do(h, "GET", "/convert/html/", nil); rec.Code != 200 {
			t.Fatalf("status %d", rec.Code)
		}
	}
	if time.Since(start) > time.Second {
		t.Errorf("responses waited for the sender")
	}
	close(release)
	settle(h)
	if calls != 2 {
		t.Errorf("%d sends, want 2", calls)
	}
	if lines := logs.all(); len(lines) != 1 || lines[0] != "analytics: send skipped (busy)" {
		t.Errorf("logged %q", lines)
	}
}

func TestClientIP(t *testing.T) {
	trusted, err := parseTrusted("10.42.0.0/16, 10.0.0.2")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		remote, xff, want string
	}{
		// The gateway appends the visitor; a forged entry on the left is ignored.
		{"10.42.3.7:5000", "198.51.100.1, 203.0.113.9", "203.0.113.9"},
		{"10.42.3.7:5000", "203.0.113.9", "203.0.113.9"},
		{"10.42.3.7:5000", "2001:db8::1", "2001:db8::1"},
		{"10.42.3.7:5000", "::ffff:203.0.113.9", "203.0.113.9"},
		{"10.42.3.7:5000", "203.0.113.9, 10.42.0.5", "203.0.113.9"},
		{"10.42.3.7:5000", "203.0.113.9:4711", "203.0.113.9"},
		// Not from a trusted proxy: the header is not believed at all.
		{"192.0.2.10:5000", "203.0.113.9", "192.0.2.10"},
		{"[2001:db8::2]:5000", "203.0.113.9", "2001:db8::2"},
		// From a trusted proxy, with no header.
		{"10.42.3.7:5000", "", "10.42.3.7"},
		// Garbage in the trusted part of the chain: unknown.
		{"10.42.3.7:5000", "203.0.113.9, nonsense", ""},
	} {
		r := httptest.NewRequest("GET", "/", nil)
		r.RemoteAddr = tc.remote
		if tc.xff != "" {
			r.Header.Set("X-Forwarded-For", tc.xff)
		}
		got, ok := clientIP(r, trusted)
		if tc.want == "" {
			if ok {
				t.Errorf("%s %q: %v, want unknown", tc.remote, tc.xff, got)
			}
			continue
		}
		if !ok || got != netip.MustParseAddr(tc.want) {
			t.Errorf("%s %q: %v %v, want %s", tc.remote, tc.xff, got, ok, tc.want)
		}
	}
	// Without TRUSTED_PROXIES nothing in the header is believed.
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "10.42.3.7:5000"
	r.Header.Set("X-Forwarded-For", "203.0.113.9")
	if got, _ := clientIP(r, nil); got.String() != "10.42.3.7" {
		t.Errorf("no trusted proxies: %v", got)
	}
	if _, err := parseTrusted("10.42.0.0/16,nonsense"); err == nil {
		t.Error("parseTrusted accepted nonsense")
	}
}

func TestTruncate(t *testing.T) {
	if got := truncate("abcdef", 4); got != "abcd" {
		t.Errorf("%q", got)
	}
	// "š" is two bytes: cutting inside it drops it whole.
	if got := truncate("abšc", 3); got != "ab" {
		t.Errorf("%q", got)
	}
	if got := truncate("abšc", 4); got != "abš" {
		t.Errorf("%q", got)
	}
}
