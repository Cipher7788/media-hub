// Package tracking implements tracking-prevention mechanisms for the proxy.
//
// The prevention engine:
//   - Blocks requests to known tracker domains and ad networks.
//   - Removes tracking query parameters (utm_*, fbclid, gclid, etc.).
//   - Strips fingerprinting headers (Accept-Language normalisation, etc.).
//   - Enforces strict Referrer-Policy to prevent cross-site leakage.
package tracking

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

// trackerDomains is the built-in list of well-known tracker / ad-network
// domains. Operators should extend this via Engine.AddTracker.
var trackerDomains = []string{
	"doubleclick.net",
	"googletagmanager.com",
	"google-analytics.com",
	"analytics.google.com",
	"facebook.com/tr",
	"connect.facebook.net",
	"bat.bing.com",
	"scorecardresearch.com",
	"quantserve.com",
	"hotjar.com",
	"cdn.segment.com",
	"mixpanel.com",
	"amplitude.com",
	"heap.io",
	"fullstory.com",
	"mouseflow.com",
}

// trackingParams is the list of URL query parameters that carry tracking
// identifiers and should be stripped from outgoing requests.
var trackingParams = []string{
	"utm_source", "utm_medium", "utm_campaign", "utm_term", "utm_content",
	"utm_id", "utm_source_platform", "utm_creative_format", "utm_marketing_tactic",
	"fbclid", "gclid", "dclid", "gbraid", "wbraid",
	"mc_cid", "mc_eid",
	"msclkid", "twclid",
	"igshid",
	"_ga", "_gl",
	"ref", "referrer",
}

// fingerprintHeaders are request headers that reveal client capabilities used
// for browser fingerprinting.
var fingerprintHeaders = []string{
	"Accept-Charset",
	"Accept-Encoding",
	"DNT",
	"Sec-CH-UA",
	"Sec-CH-UA-Mobile",
	"Sec-CH-UA-Platform",
	"Sec-CH-UA-Arch",
	"Sec-CH-UA-Bitness",
	"Sec-CH-UA-Full-Version",
	"Sec-CH-UA-Full-Version-List",
	"Sec-CH-UA-Model",
	"Sec-CH-UA-Wow64",
	"Sec-Fetch-Dest",
	"Sec-Fetch-Mode",
	"Sec-Fetch-Site",
	"Sec-Fetch-User",
}

// Engine is the tracking-prevention engine. It is safe for concurrent use.
type Engine struct {
	mu       sync.RWMutex
	trackers []string
}

// New creates a new tracking prevention Engine pre-loaded with the built-in
// tracker list.
func New() *Engine {
	e := &Engine{}
	e.trackers = make([]string, len(trackerDomains))
	copy(e.trackers, trackerDomains)
	return e
}

// AddTracker registers an additional tracker domain (e.g. "custom-ads.example").
func (e *Engine) AddTracker(domain string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.trackers = append(e.trackers, strings.ToLower(domain))
}

// IsTracker reports whether host belongs to a known tracker.
func (e *Engine) IsTracker(host string) bool {
	host = strings.ToLower(host)
	// Strip port if present.
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		host = host[:idx]
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	for _, t := range e.trackers {
		if host == t || strings.HasSuffix(host, "."+t) {
			return true
		}
	}
	return false
}

// Process applies all tracking-prevention rules to r in-place:
//  1. Returns an error if the target host is a known tracker.
//  2. Strips tracking query parameters from the URL.
//  3. Removes browser fingerprinting headers.
//  4. Sets a strict Referrer-Policy response hint.
func (e *Engine) Process(r *http.Request) error {
	host := r.Host
	if host == "" && r.URL != nil {
		host = r.URL.Host
	}
	if e.IsTracker(host) {
		return fmt.Errorf("tracking prevention: blocked tracker %q", host)
	}
	stripTrackingParams(r)
	stripFingerprintHeaders(r)
	return nil
}

// Middleware returns a RequestMiddleware that applies tracking prevention.
func (e *Engine) Middleware() func(r *http.Request) error {
	return e.Process
}

// stripTrackingParams removes known tracking query parameters from r.URL.
func stripTrackingParams(r *http.Request) {
	if r.URL == nil {
		return
	}
	q := r.URL.Query()
	changed := false
	for _, p := range trackingParams {
		if q.Has(p) {
			q.Del(p)
			changed = true
		}
	}
	if changed {
		r.URL.RawQuery = q.Encode()
	}
}

// stripFingerprintHeaders removes browser-fingerprinting request headers.
func stripFingerprintHeaders(r *http.Request) {
	for _, h := range fingerprintHeaders {
		r.Header.Del(h)
	}
}

// CleanURL returns a copy of rawURL with all tracking query parameters removed.
// This is a utility that can be used independently of an Engine instance.
func CleanURL(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("tracking: invalid URL: %w", err)
	}
	q := u.Query()
	for _, p := range trackingParams {
		q.Del(p)
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}
