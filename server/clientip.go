package main

import (
	"fmt"
	"net/http"
	"net/netip"
	"strings"
)

// The request path is: visitor -> a layer-4 load balancer (PROXY protocol) -> a reverse proxy, which learns the
// visitor's address from the PROXY header and appends it to X-Forwarded-For -> this server, whose peer is the proxy.
// So the rightmost X-Forwarded-For entry that is not one of the trusted proxies (TRUSTED_PROXIES) is the visitor;
// everything left of it was sent by the client and is not believed.

// parseTrusted reads a comma-separated list of CIDRs or single addresses.
func parseTrusted(s string) ([]netip.Prefix, error) {
	var out []netip.Prefix
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if p, err := netip.ParsePrefix(part); err == nil {
			out = append(out, p.Masked())
			continue
		}
		a, err := netip.ParseAddr(part)
		if err != nil {
			return nil, fmt.Errorf("TRUSTED_PROXIES: %q is not an address or CIDR", part)
		}
		out = append(out, netip.PrefixFrom(a.Unmap(), a.Unmap().BitLen()))
	}
	return out, nil
}

func isTrusted(a netip.Addr, trusted []netip.Prefix) bool {
	for _, p := range trusted {
		if p.Contains(a) {
			return true
		}
	}
	return false
}

// parseHop reads one address: a bare IP, or one with a port ("1.2.3.4:5", "[2001:db8::1]:5").
func parseHop(s string) (netip.Addr, bool) {
	s = strings.TrimSpace(s)
	if a, err := netip.ParseAddr(s); err == nil {
		return a.Unmap(), true
	}
	if ap, err := netip.ParseAddrPort(s); err == nil {
		return ap.Addr().Unmap(), true
	}
	return netip.Addr{}, false
}

// clientIP returns the visitor's address. ok is false when it cannot be known (an unreadable entry in the trusted
// part of the chain); the page view is then not sent.
func clientIP(r *http.Request, trusted []netip.Prefix) (netip.Addr, bool) {
	addr, ok := parseHop(r.RemoteAddr)
	if !ok {
		return netip.Addr{}, false
	}
	if !isTrusted(addr, trusted) {
		return addr, true
	}
	var hops []string
	for _, v := range r.Header.Values("X-Forwarded-For") {
		hops = append(hops, strings.Split(v, ",")...)
	}
	for i := len(hops) - 1; i >= 0; i-- {
		a, ok := parseHop(hops[i])
		if !ok {
			return netip.Addr{}, false
		}
		if !isTrusted(a, trusted) {
			return a, true
		}
		addr = a
	}
	// Every hop is a trusted proxy (a request from inside): the first one.
	return addr, true
}
