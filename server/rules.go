package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

// The site's headers and redirects live in static/_headers and static/_redirects, in Cloudflare Pages' formats
// (https://developers.cloudflare.com/pages/configuration/headers/ and .../redirects/). Hugo copies them into the build,
// and this server reads them from there, so those two files stay the only place the rules are written.
//
// Only the subset the files use is accepted. Anything else (placeholders such as :slug, "!" to remove a header,
// absolute URLs as the source, rewrites with status 200, more than one *) fails at startup and in the tests, instead
// of being quietly ignored.

// headerRule sets headers on every response whose request path matches the pattern.
type headerRule struct {
	pattern string
	match   *regexp.Regexp
	headers [][2]string // name, value, in file order
}

// redirectRule answers requests for its source with a redirect. A source ending in * matches everything that starts
// with the part before it; :splat in the destination is what the * matched.
type redirectRule struct {
	source string
	prefix bool // source ended in *; source holds the part before it
	dest   string
	status int
}

var headerName = regexp.MustCompile("^[!#$%&'*+.^_`|~0-9A-Za-z-]+$")

func parseHeaders(r io.Reader) ([]headerRule, error) {
	var rules []headerRule
	sc := bufio.NewScanner(r)
	for n := 1; sc.Scan(); n++ {
		raw := sc.Text()
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if raw[0] != ' ' && raw[0] != '\t' {
			if len(rules) > 0 && len(rules[len(rules)-1].headers) == 0 {
				return nil, fmt.Errorf("_headers line %d: the rule before it sets no header", n)
			}
			re, err := pathPattern(line)
			if err != nil {
				return nil, fmt.Errorf("_headers line %d: %w", n, err)
			}
			rules = append(rules, headerRule{pattern: line, match: re})
			continue
		}
		if len(rules) == 0 {
			return nil, fmt.Errorf("_headers line %d: a header before any path", n)
		}
		if strings.HasPrefix(line, "!") {
			return nil, fmt.Errorf("_headers line %d: removing headers (!) is not supported", n)
		}
		name, value, ok := strings.Cut(line, ":")
		name, value = strings.TrimSpace(name), strings.TrimSpace(value)
		if !ok || !headerName.MatchString(name) || value == "" {
			return nil, fmt.Errorf("_headers line %d: not a \"Name: value\" header", n)
		}
		last := &rules[len(rules)-1]
		last.headers = append(last.headers, [2]string{http.CanonicalHeaderKey(name), value})
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if len(rules) > 0 && len(rules[len(rules)-1].headers) == 0 {
		return nil, fmt.Errorf("_headers: the last rule sets no header")
	}
	return rules, nil
}

// pathPattern turns a path with * wildcards (each matches anything, / included) into an anchored expression.
func pathPattern(p string) (*regexp.Regexp, error) {
	if !strings.HasPrefix(p, "/") {
		return nil, fmt.Errorf("%q: only paths starting with / are supported (no host names)", p)
	}
	if strings.ContainsAny(p, " \t") {
		return nil, fmt.Errorf("%q: a path has no spaces", p)
	}
	for _, seg := range strings.Split(p, "/") {
		if strings.HasPrefix(seg, ":") {
			return nil, fmt.Errorf("%q: placeholders (:name) are not supported", p)
		}
	}
	parts := strings.Split(p, "*")
	for i := range parts {
		parts[i] = regexp.QuoteMeta(parts[i])
	}
	return regexp.Compile("^" + strings.Join(parts, ".*") + "$")
}

func parseRedirects(r io.Reader) ([]redirectRule, error) {
	var rules []redirectRule
	sc := bufio.NewScanner(r)
	for n := 1; sc.Scan(); n++ {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		f := strings.Fields(line)
		if len(f) != 2 && len(f) != 3 {
			return nil, fmt.Errorf("_redirects line %d: want \"source destination [status]\"", n)
		}
		rule := redirectRule{source: f[0], dest: f[1], status: http.StatusFound}
		if len(f) == 3 {
			status, err := strconv.Atoi(f[2])
			if err != nil {
				return nil, fmt.Errorf("_redirects line %d: status %q is not a number", n, f[2])
			}
			switch status {
			case 301, 302, 303, 307, 308:
				rule.status = status
			default:
				return nil, fmt.Errorf("_redirects line %d: status %d is not supported (only redirects)", n, status)
			}
		}
		src := rule.source
		if !strings.HasPrefix(src, "/") {
			return nil, fmt.Errorf("_redirects line %d: the source must be a path starting with /", n)
		}
		for _, seg := range strings.Split(src, "/") {
			if strings.HasPrefix(seg, ":") {
				return nil, fmt.Errorf("_redirects line %d: placeholders (:name) are not supported", n)
			}
		}
		if i := strings.Index(src, "*"); i >= 0 {
			if i != len(src)-1 {
				return nil, fmt.Errorf("_redirects line %d: * is supported only at the end of the source", n)
			}
			rule.source, rule.prefix = src[:i], true
		}
		if strings.Contains(rule.dest, ":splat") && !rule.prefix {
			return nil, fmt.Errorf("_redirects line %d: :splat without * in the source", n)
		}
		if !strings.HasPrefix(rule.dest, "/") && !strings.HasPrefix(rule.dest, "https://") &&
			!strings.HasPrefix(rule.dest, "http://") {
			return nil, fmt.Errorf("_redirects line %d: the destination must be a path or an http(s) URL", n)
		}
		if strings.HasPrefix(rule.dest, "//") {
			return nil, fmt.Errorf("_redirects line %d: the destination must not start with //", n)
		}
		rules = append(rules, rule)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return rules, nil
}

// findRedirect returns the Location and status for an escaped request path and raw query, or ok=false. The first
// matching rule wins. The query goes along to the destination unless the destination has one of its own.
func findRedirect(rules []redirectRule, escapedPath, rawQuery string) (location string, status int, ok bool) {
	for _, rule := range rules {
		var dest string
		switch {
		case rule.prefix && strings.HasPrefix(escapedPath, rule.source):
			dest = strings.ReplaceAll(rule.dest, ":splat", escapedPath[len(rule.source):])
		case !rule.prefix && escapedPath == rule.source:
			dest = rule.dest
		default:
			continue
		}
		// A splat such as "/evil.example" after "/docs/" would make "//evil.example", which browsers read as another
		// host. Keep a path a path.
		if strings.HasPrefix(dest, "/") {
			dest = "/" + strings.TrimLeft(dest, "/\\")
		}
		return withQuery(dest, rawQuery), rule.status, true
	}
	return "", 0, false
}

// withQuery appends a query to a URL that has none, before its #fragment.
func withQuery(dest, rawQuery string) string {
	if rawQuery == "" || strings.Contains(dest, "?") {
		return dest
	}
	base, frag, hasFrag := strings.Cut(dest, "#")
	if hasFrag {
		return base + "?" + rawQuery + "#" + frag
	}
	return base + "?" + rawQuery
}

// applyHeaders sets the headers of every rule matching the request path. A header set by more than one matching rule
// gets their values joined with ", ", as Cloudflare Pages does.
func applyHeaders(h http.Header, rules []headerRule, urlPath string) {
	seen := map[string]bool{}
	for _, rule := range rules {
		if !rule.match.MatchString(urlPath) {
			continue
		}
		for _, kv := range rule.headers {
			if seen[kv[0]] {
				h.Set(kv[0], h.Get(kv[0])+", "+kv[1])
			} else {
				h.Set(kv[0], kv[1])
				seen[kv[0]] = true
			}
		}
	}
	// Cloudflare Pages' default when no rule says otherwise: browsers may keep a copy but must ask again each time.
	if h.Get("Cache-Control") == "" {
		h.Set("Cache-Control", "public, max-age=0, must-revalidate")
	}
}
