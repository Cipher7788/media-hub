// Package core implements the HTTP/HTTPS proxy engine.
//
// The proxy supports:
//   - Plain HTTP proxying via request forwarding
//   - HTTPS tunnelling via the CONNECT method
//   - Per-request middleware hooks for firewall, anonymisation, and tracking
//     prevention layers
package core

import (
	"crypto/tls"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"time"
)

// Config holds runtime configuration for the proxy server.
type Config struct {
	// ListenAddr is the TCP address the proxy listens on (e.g. ":8080").
	ListenAddr string
	// ReadTimeout is the maximum duration for reading a complete request.
	ReadTimeout time.Duration
	// WriteTimeout is the maximum duration before timing out a response write.
	WriteTimeout time.Duration
	// IdleTimeout is the maximum amount of time to wait for the next request.
	IdleTimeout time.Duration
	// TLSConfig, when non-nil, enables TLS on the listening socket.
	TLSConfig *tls.Config
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		ListenAddr:   ":8080",
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  90 * time.Second,
	}
}

// RequestMiddleware is a function that can inspect or modify an incoming
// request before it is forwarded. Returning a non-nil error aborts the
// request with a 403 response.
type RequestMiddleware func(r *http.Request) error

// Proxy is the core HTTP/HTTPS proxy engine.
type Proxy struct {
	cfg         Config
	transport   *http.Transport
	logger      *slog.Logger
	middlewares []RequestMiddleware
}

// New creates a new Proxy with the supplied configuration.
func New(cfg Config, logger *slog.Logger) *Proxy {
	if logger == nil {
		logger = slog.Default()
	}
	return &Proxy{
		cfg:    cfg,
		logger: logger,
		transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 20 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			MaxIdleConns:          200,
			MaxIdleConnsPerHost:   20,
			IdleConnTimeout:       90 * time.Second,
			// Disable HTTP/2 so we can inspect plain-text headers in the DPI layer.
			ForceAttemptHTTP2: false,
		},
	}
}

// Use appends a RequestMiddleware to the proxy's middleware chain.
func (p *Proxy) Use(m RequestMiddleware) {
	p.middlewares = append(p.middlewares, m)
}

// ListenAndServe starts the proxy and blocks until the server returns.
func (p *Proxy) ListenAndServe() error {
	srv := &http.Server{
		Addr:         p.cfg.ListenAddr,
		Handler:      p,
		ReadTimeout:  p.cfg.ReadTimeout,
		WriteTimeout: p.cfg.WriteTimeout,
		IdleTimeout:  p.cfg.IdleTimeout,
		TLSConfig:    p.cfg.TLSConfig,
	}
	p.logger.Info("proxy listening", "addr", p.cfg.ListenAddr)
	return srv.ListenAndServe()
}

// ServeHTTP is the main dispatcher. CONNECT requests are handled as HTTPS
// tunnels; all other requests are forwarded as plain HTTP.
func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Run middleware chain.
	for _, m := range p.middlewares {
		if err := m(r); err != nil {
			p.logger.Warn("request blocked by middleware",
				"method", r.Method,
				"host", r.Host,
				"reason", err,
			)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
	}

	if r.Method == http.MethodConnect {
		p.handleTunnel(w, r)
	} else {
		p.handleHTTP(w, r)
	}

	p.logger.Info("request",
		"method", r.Method,
		"host", r.Host,
		"path", r.URL.Path,
		"duration", time.Since(start),
	)
}

// handleHTTP forwards a plain HTTP request to the upstream server.
func (p *Proxy) handleHTTP(w http.ResponseWriter, r *http.Request) {
	// Remove hop-by-hop headers before forwarding.
	removeHopHeaders(r.Header)
	r.RequestURI = ""

	resp, err := p.transport.RoundTrip(r)
	if err != nil {
		p.logger.Error("upstream request failed", "err", err)
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	removeHopHeaders(resp.Header)
	copyHeaders(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(w, resp.Body); err != nil {
		p.logger.Warn("response copy error", "err", err)
	}
}

// handleTunnel establishes a raw TCP tunnel for HTTPS (CONNECT) requests.
func (p *Proxy) handleTunnel(w http.ResponseWriter, r *http.Request) {
	dest, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		p.logger.Error("tunnel dial failed", "host", r.Host, "err", err)
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	defer dest.Close()

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	client, _, err := hijacker.Hijack()
	if err != nil {
		p.logger.Error("hijack failed", "err", err)
		return
	}
	defer client.Close()

	errCh := make(chan error, 2)
	pipe := func(dst, src net.Conn) {
		_, err := io.Copy(dst, src)
		errCh <- err
	}

	go pipe(dest, client)
	go pipe(client, dest)

	// Wait for one side to close.
	if err := <-errCh; err != nil {
		p.logger.Debug("tunnel closed", "host", r.Host, "reason", err)
	}
}

// hopHeaders is the set of headers that must not be forwarded by a proxy.
var hopHeaders = []string{
	"Connection", "Proxy-Connection", "Keep-Alive", "Proxy-Authenticate",
	"Proxy-Authorization", "Te", "Trailers", "Transfer-Encoding", "Upgrade",
}

func removeHopHeaders(h http.Header) {
	// Read Connection header BEFORE deleting it, so we can also remove the
	// headers it names (per RFC 7230 §6.1).
	connHeaders := splitComma(h.Get("Connection"))

	for _, k := range hopHeaders {
		h.Del(k)
	}
	for _, f := range connHeaders {
		if f != "" {
			h.Del(f)
		}
	}
}

func copyHeaders(dst, src http.Header) {
	for k, vs := range src {
		for _, v := range vs {
			dst.Add(k, v)
		}
	}
}

func splitComma(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			out = append(out, trimSpace(s[start:i]))
			start = i + 1
		}
	}
	out = append(out, trimSpace(s[start:]))
	return out
}

func trimSpace(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t') {
		s = s[:len(s)-1]
	}
	return s
}

// Version returns the semantic version of the proxy engine.
func Version() string {
	return fmt.Sprintf("proxy-engine/1.0.0 go/%s", goVersion())
}

func goVersion() string {
	return "1.22"
}
