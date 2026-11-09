package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"
)

// Page views go to an ingest endpoint (ANALYTICS_URL), which does all the processing: bot filter,
// country, browser, the daily salted hash. This side only picks real page views (countable, in handler.go) and
// forwards them: after the response, in a goroutine, once, with a 2 s timeout. No retry, no queue, nothing on disk.
// When all slots are busy the page view is dropped.
//
// Nothing about a page view is ever logged. A failure logs one line with a fixed reason (timeout, network, status
// code); the error itself is not printed, so neither the request URL nor anything from the visitor can reach the log.

const (
	sendTimeout = 2 * time.Second
	maxInFlight = 32
)

// hit is the ingest API's request body.
type hit struct {
	Site      string `json:"site"`
	URL       string `json:"url"`
	Referrer  string `json:"referrer"`
	IP        string `json:"ip"`
	UserAgent string `json:"user_agent"`
}

type sender struct {
	url, token, site string
	client           *http.Client
	timeout          time.Duration
	slots            chan struct{}
	wg               sync.WaitGroup
	lastBusyLog      atomic.Int64 // unix seconds
	logf             func(format string, args ...any)
}

func newSender(url, token, site string) *sender {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil // straight to the endpoint, no proxy
	transport.MaxIdleConnsPerHost = maxInFlight
	return &sender{
		url:   url,
		token: token,
		site:  site,
		client: &http.Client{
			Transport: transport,
			// A redirect is an error on the endpoint's side; following it could carry the body somewhere else.
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		},
		timeout: sendTimeout,
		slots:   make(chan struct{}, maxInFlight),
		logf:    log.Printf,
	}
}

// send forwards a page view in the background, or drops it when maxInFlight sends are already running.
func (s *sender) send(h hit) {
	select {
	case s.slots <- struct{}{}:
	default:
		now := time.Now().Unix()
		if last := s.lastBusyLog.Load(); now-last >= 60 && s.lastBusyLog.CompareAndSwap(last, now) {
			s.logf("analytics: send skipped (busy)")
		}
		return
	}
	s.wg.Add(1)
	go func() {
		defer func() {
			<-s.slots
			s.wg.Done()
		}()
		s.post(h)
	}()
}

func (s *sender) post(h hit) {
	h.URL = truncate(h.URL, 2048)
	h.Referrer = truncate(h.Referrer, 2048)
	h.UserAgent = truncate(h.UserAgent, 1024)
	body, err := json.Marshal(h)
	if err != nil {
		s.logf("analytics: send failed (encode)")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.url, bytes.NewReader(body))
	if err != nil {
		s.logf("analytics: send failed (request)")
		return
	}
	req.Header.Set("Authorization", "Bearer "+s.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		var ne net.Error
		if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &ne) && ne.Timeout()) {
			s.logf("analytics: send failed (timeout)")
		} else {
			s.logf("analytics: send failed (network)")
		}
		return
	}
	io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))
	resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		s.logf("analytics: send failed (status %d)", resp.StatusCode)
	}
}

// wait blocks until the running sends are done, or the context ends.
func (s *sender) wait(ctx context.Context) {
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-ctx.Done():
	}
}

// truncate cuts s to at most n bytes without splitting a UTF-8 character.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	// s[n] is the first byte left out; back up while it continues the character before it.
	for n > 0 && !utf8.RuneStart(s[n]) {
		n--
	}
	s = s[:n]
	return s
}
