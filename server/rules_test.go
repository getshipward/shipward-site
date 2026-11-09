package main

import (
	"net/http"
	"os"
	"strings"
	"testing"
)

// countRuleLines counts the lines of a rules file that start a rule: not blank, not a comment and, for _headers,
// not indented.
func countRuleLines(t *testing.T, name string, indentedToo bool) int {
	t.Helper()
	b, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, line := range strings.Split(string(b), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if !indentedToo && (line[0] == ' ' || line[0] == '\t') {
			continue
		}
		n++
	}
	return n
}

// The real files: every rule in them must parse, so a rule in a form this server does not know fails CI instead of
// being skipped in production.
func TestRealHeadersFileParses(t *testing.T) {
	f, err := os.Open("../static/_headers")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	rules, err := parseHeaders(f)
	if err != nil {
		t.Fatal(err)
	}
	if want := countRuleLines(t, "../static/_headers", false); len(rules) != want || want == 0 {
		t.Fatalf("parsed %d header rules, the file has %d", len(rules), want)
	}
	h := http.Header{}
	applyHeaders(h, rules, "/convert/html/")
	if !strings.Contains(h.Get("Content-Security-Policy"), "default-src 'none'") {
		t.Errorf("CSP on a page: %q", h.Get("Content-Security-Policy"))
	}
	h = http.Header{}
	applyHeaders(h, rules, "/css/site.123.css")
	if h.Get("Cache-Control") != "public, max-age=31536000, immutable" {
		t.Errorf("Cache-Control on a stylesheet: %q", h.Get("Cache-Control"))
	}
}

func TestRealRedirectsFileParses(t *testing.T) {
	f, err := os.Open("../static/_redirects")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	rules, err := parseRedirects(f)
	if err != nil {
		t.Fatal(err)
	}
	if want := countRuleLines(t, "../static/_redirects", true); len(rules) != want {
		t.Fatalf("parsed %d redirects, the file has %d", len(rules), want)
	}
}

func TestParseHeadersRejects(t *testing.T) {
	for name, in := range map[string]string{
		"placeholder":     "/:slug\n  X-A: b\n",
		"detach":          "/*\n  ! X-A\n",
		"host":            "https://example.com/*\n  X-A: b\n",
		"header first":    "  X-A: b\n",
		"no colon":        "/*\n  X-A b\n",
		"empty value":     "/*\n  X-A:\n",
		"rule no headers": "/a\n/b\n  X-A: b\n",
		"last no headers": "/a\n  X-A: b\n/b\n",
	} {
		if _, err := parseHeaders(strings.NewReader(in)); err == nil {
			t.Errorf("%s: parsed, want an error", name)
		}
	}
}

func TestParseRedirectsRejects(t *testing.T) {
	for name, in := range map[string]string{
		"rewrite":           "/a /b 200\n",
		"404":               "/a /b 404\n",
		"placeholder":       "/a/:id /b/:id 301\n",
		"star in middle":    "/a/*/b /c 301\n",
		"splat without *":   "/a /:splat 301\n",
		"host source":       "https://x.example/a /b 301\n",
		"relative dest":     "/a b 301\n",
		"protocol-relative": "/a //evil.example 301\n",
		"too many fields":   "/a /b 301 force\n",
		"one field":         "/a\n",
		"bad status":        "/a /b moved\n",
	} {
		if _, err := parseRedirects(strings.NewReader(in)); err == nil {
			t.Errorf("%s: parsed, want an error", name)
		}
	}
}

func TestFindRedirect(t *testing.T) {
	rules, err := parseRedirects(strings.NewReader(`
/docs / 301
/docs/endpoints /#routes 301
/docs/* /:splat 301
/old /new/
/away https://example.com/x 308
`))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		path, query, loc string
		status           int
	}{
		{"/docs", "", "/", 301},
		{"/docs/endpoints", "", "/#routes", 301},
		{"/docs/endpoints", "a=1", "/?a=1#routes", 301},
		{"/docs/convert/html/", "utm_source=x", "/convert/html/?utm_source=x", 301},
		{"/docs/", "", "/", 301},
		{"/docs//evil.example", "", "/evil.example", 301},
		{`/docs/\evil.example`, "", "/evil.example", 301},
		{"/old", "", "/new/", 302},
		{"/away", "q=1", "https://example.com/x?q=1", 308},
	} {
		loc, status, ok := findRedirect(rules, tc.path, tc.query)
		if !ok || loc != tc.loc || status != tc.status {
			t.Errorf("%s?%s: got %q %d %v, want %q %d", tc.path, tc.query, loc, status, ok, tc.loc, tc.status)
		}
	}
	for _, p := range []string{"/", "/docsx", "/old/", "/convert/html/"} {
		if _, _, ok := findRedirect(rules, p, ""); ok {
			t.Errorf("%s: redirected, want no match", p)
		}
	}
}

func TestApplyHeadersJoinsAndDefaults(t *testing.T) {
	rules, err := parseHeaders(strings.NewReader("/*\n  X-A: one\n/a/*\n  X-A: two\n  X-B: b\n"))
	if err != nil {
		t.Fatal(err)
	}
	h := http.Header{}
	applyHeaders(h, rules, "/a/x")
	if h.Get("X-A") != "one, two" || h.Get("X-B") != "b" {
		t.Errorf("got %v", h)
	}
	if h.Get("Cache-Control") != "public, max-age=0, must-revalidate" {
		t.Errorf("default Cache-Control: %q", h.Get("Cache-Control"))
	}
	h = http.Header{}
	applyHeaders(h, rules, "/b")
	if h.Get("X-A") != "one" || h.Get("X-B") != "" {
		t.Errorf("got %v", h)
	}
}
