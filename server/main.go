// Command server serves the built site (Hugo's public/) at shipward.dev: the headers and redirects from static/_headers
// and static/_redirects, directory URLs with a trailing slash, 404.html for unknown paths. It can forward page views to
// an analytics endpoint (analytics.go).
//
// Settings, from the environment:
//
//	LISTEN_ADDR      where to listen (default :8080); plain HTTP, TLS ends at the gateway
//	SITE_DIR         the built site (default public)
//	TRUSTED_PROXIES  comma-separated CIDRs whose X-Forwarded-For entry is believed (the gateway's pod network)
//	ANALYTICS_URL    the full ingest endpoint (.../internal/analytics/v1/hits), posted to as is; empty: off
//	ANALYTICS_TOKEN  the bearer token for that endpoint
//	ANALYTICS_SITE   this site's host, as the endpoint knows it
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	log.SetFlags(0)
	if err := run(); err != nil {
		log.Fatalf("server: %v", err)
	}
}

func env(name, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(name)); v != "" {
		return v
	}
	return fallback
}

func run() error {
	addr := env("LISTEN_ADDR", ":8080")
	dir := env("SITE_DIR", "public")

	s, err := loadSite(dir)
	if err != nil {
		return err
	}
	trusted, err := parseTrusted(os.Getenv("TRUSTED_PROXIES"))
	if err != nil {
		return err
	}
	h := &handler{site: s, trusted: trusted}

	if u := env("ANALYTICS_URL", ""); u != "" {
		parsed, err := url.Parse(u)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			return errors.New("ANALYTICS_URL is not an http(s) URL")
		}
		token, site := env("ANALYTICS_TOKEN", ""), strings.ToLower(env("ANALYTICS_SITE", ""))
		if token == "" || site == "" {
			return errors.New("ANALYTICS_URL is set: ANALYTICS_TOKEN and ANALYTICS_SITE are needed too")
		}
		h.analytics = newSender(u, token, site)
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    64 << 10,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errc := make(chan error, 1)
	go func() { errc <- srv.ListenAndServe() }()
	state := "off"
	if h.analytics != nil {
		state = "on for " + h.analytics.site
	}
	log.Printf("server: %d files from %s on %s, %d header rules, %d redirects, analytics %s",
		len(s.files), dir, addr, len(s.headers), len(s.redirects), state)

	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
	}
	log.Printf("server: shutting down")
	shutdown, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdown); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	if h.analytics != nil {
		h.analytics.wait(shutdown)
	}
	return nil
}
