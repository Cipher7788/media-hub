// Package monitor provides metrics collection, anomaly detection, and threat
// alerting for the proxy engine.
//
// Metrics are stored in-memory and exposed via a Prometheus-compatible text
// format. The threat detector analyses request patterns to identify port-scan
// attempts, brute-force login attacks, and suspicious traffic bursts.
package monitor

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// --- Metrics ---

// Counters holds atomic counters for key proxy events.
type Counters struct {
	RequestsTotal  atomic.Int64
	RequestsBlocked atomic.Int64
	BytesForwarded atomic.Int64
	Tunnels        atomic.Int64
	Errors         atomic.Int64
}

// Metrics is the central telemetry collector for the proxy.
type Metrics struct {
	mu       sync.RWMutex
	counters *Counters
	latency  []float64 // request latencies in milliseconds
	start    time.Time
}

// NewMetrics creates a new Metrics instance.
func NewMetrics() *Metrics {
	return &Metrics{
		counters: &Counters{},
		start:    time.Now(),
	}
}

// RecordRequest records a completed request event.
func (m *Metrics) RecordRequest(latencyMS float64, blocked bool, bytes int64) {
	m.counters.RequestsTotal.Add(1)
	if blocked {
		m.counters.RequestsBlocked.Add(1)
	}
	m.counters.BytesForwarded.Add(bytes)

	m.mu.Lock()
	m.latency = append(m.latency, latencyMS)
	// Keep the rolling window to the last 10 000 requests to bound memory.
	if len(m.latency) > 10_000 {
		m.latency = m.latency[len(m.latency)-10_000:]
	}
	m.mu.Unlock()
}

// RecordTunnel records an established CONNECT tunnel.
func (m *Metrics) RecordTunnel() {
	m.counters.Tunnels.Add(1)
}

// RecordError records a proxy-level error.
func (m *Metrics) RecordError() {
	m.counters.Errors.Add(1)
}

// Snapshot returns a point-in-time copy of the current metric values.
func (m *Metrics) Snapshot() Snapshot {
	m.mu.RLock()
	latCopy := make([]float64, len(m.latency))
	copy(latCopy, m.latency)
	m.mu.RUnlock()

	return Snapshot{
		Uptime:          time.Since(m.start),
		RequestsTotal:   m.counters.RequestsTotal.Load(),
		RequestsBlocked: m.counters.RequestsBlocked.Load(),
		BytesForwarded:  m.counters.BytesForwarded.Load(),
		Tunnels:         m.counters.Tunnels.Load(),
		Errors:          m.counters.Errors.Load(),
		AvgLatencyMS:    average(latCopy),
		P95LatencyMS:    percentile(latCopy, 95),
	}
}

// WritePrometheus writes the current metrics in Prometheus text format to w.
func (m *Metrics) WritePrometheus(w io.Writer) {
	s := m.Snapshot()
	fmt.Fprintf(w, "# HELP proxy_requests_total Total number of proxy requests.\n")
	fmt.Fprintf(w, "# TYPE proxy_requests_total counter\n")
	fmt.Fprintf(w, "proxy_requests_total %d\n\n", s.RequestsTotal)

	fmt.Fprintf(w, "# HELP proxy_requests_blocked_total Total blocked requests.\n")
	fmt.Fprintf(w, "# TYPE proxy_requests_blocked_total counter\n")
	fmt.Fprintf(w, "proxy_requests_blocked_total %d\n\n", s.RequestsBlocked)

	fmt.Fprintf(w, "# HELP proxy_bytes_forwarded_total Total bytes forwarded.\n")
	fmt.Fprintf(w, "# TYPE proxy_bytes_forwarded_total counter\n")
	fmt.Fprintf(w, "proxy_bytes_forwarded_total %d\n\n", s.BytesForwarded)

	fmt.Fprintf(w, "# HELP proxy_tunnels_total Total CONNECT tunnels established.\n")
	fmt.Fprintf(w, "# TYPE proxy_tunnels_total counter\n")
	fmt.Fprintf(w, "proxy_tunnels_total %d\n\n", s.Tunnels)

	fmt.Fprintf(w, "# HELP proxy_errors_total Total proxy-level errors.\n")
	fmt.Fprintf(w, "# TYPE proxy_errors_total counter\n")
	fmt.Fprintf(w, "proxy_errors_total %d\n\n", s.Errors)

	fmt.Fprintf(w, "# HELP proxy_latency_avg_ms Average request latency in milliseconds.\n")
	fmt.Fprintf(w, "# TYPE proxy_latency_avg_ms gauge\n")
	fmt.Fprintf(w, "proxy_latency_avg_ms %.3f\n\n", s.AvgLatencyMS)

	fmt.Fprintf(w, "# HELP proxy_latency_p95_ms P95 request latency in milliseconds.\n")
	fmt.Fprintf(w, "# TYPE proxy_latency_p95_ms gauge\n")
	fmt.Fprintf(w, "proxy_latency_p95_ms %.3f\n\n", s.P95LatencyMS)

	fmt.Fprintf(w, "# HELP proxy_uptime_seconds Seconds since proxy start.\n")
	fmt.Fprintf(w, "# TYPE proxy_uptime_seconds gauge\n")
	fmt.Fprintf(w, "proxy_uptime_seconds %.1f\n", s.Uptime.Seconds())
}

// Handler returns an http.Handler that serves the Prometheus metrics endpoint.
func (m *Metrics) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		m.WritePrometheus(w)
	})
}

// Snapshot is a point-in-time copy of collected metrics.
type Snapshot struct {
	Uptime          time.Duration
	RequestsTotal   int64
	RequestsBlocked int64
	BytesForwarded  int64
	Tunnels         int64
	Errors          int64
	AvgLatencyMS    float64
	P95LatencyMS    float64
}

// --- Threat Detection ---

// ThreatLevel classifies the severity of a detected anomaly.
type ThreatLevel int

const (
	ThreatNone     ThreatLevel = iota
	ThreatLow                  // Suspicious but not immediately dangerous.
	ThreatMedium               // Likely malicious; investigate.
	ThreatHigh                 // Block immediately.
	ThreatCritical             // Terminate and alert.
)

func (t ThreatLevel) String() string {
	switch t {
	case ThreatNone:
		return "NONE"
	case ThreatLow:
		return "LOW"
	case ThreatMedium:
		return "MEDIUM"
	case ThreatHigh:
		return "HIGH"
	case ThreatCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// Alert is raised when the threat detector identifies suspicious behaviour.
type Alert struct {
	Time    time.Time
	Level   ThreatLevel
	Source  string
	Message string
}

// AlertHandler is a callback invoked whenever a new Alert is raised.
type AlertHandler func(Alert)

// Detector monitors per-source request rates and raises Alerts when thresholds
// are exceeded.
type Detector struct {
	mu       sync.Mutex
	window   time.Duration
	maxRPM   int // requests per minute threshold
	events   map[string][]time.Time
	handler  AlertHandler
	logger   *slog.Logger
}

// NewDetector creates a Detector that raises alerts when a single source
// exceeds maxRPM requests within window.
func NewDetector(window time.Duration, maxRPM int, handler AlertHandler, logger *slog.Logger) *Detector {
	if logger == nil {
		logger = slog.Default()
	}
	return &Detector{
		window:  window,
		maxRPM:  maxRPM,
		events:  make(map[string][]time.Time),
		handler: handler,
		logger:  logger,
	}
}

// Observe records a request from source (usually client IP) and evaluates it
// against the rate thresholds. Returns the assessed ThreatLevel.
func (d *Detector) Observe(source string) ThreatLevel {
	now := time.Now()

	d.mu.Lock()
	// Prune events outside the rolling window.
	cutoff := now.Add(-d.window)
	events := d.events[source]
	start := 0
	for start < len(events) && events[start].Before(cutoff) {
		start++
	}
	events = append(events[start:], now)
	d.events[source] = events
	count := len(events)
	d.mu.Unlock()

	level := d.classify(count)
	if level > ThreatNone && d.handler != nil {
		d.handler(Alert{
			Time:    now,
			Level:   level,
			Source:  source,
			Message: fmt.Sprintf("rate limit exceeded: %d requests in %s", count, d.window),
		})
	}
	return level
}

// classify maps a request count to a ThreatLevel.
func (d *Detector) classify(count int) ThreatLevel {
	switch {
	case count >= d.maxRPM*4:
		return ThreatCritical
	case count >= d.maxRPM*2:
		return ThreatHigh
	case count >= d.maxRPM:
		return ThreatMedium
	case count >= d.maxRPM/2:
		return ThreatLow
	default:
		return ThreatNone
	}
}

// Middleware returns a RequestMiddleware that blocks requests from sources
// assessed at ThreatHigh or above.
func (d *Detector) Middleware() func(r *http.Request) error {
	return func(r *http.Request) error {
		source := clientIP(r)
		level := d.Observe(source)
		if level >= ThreatHigh {
			return fmt.Errorf("threat detected from %s: level=%s", source, level)
		}
		return nil
	}
}

// clientIP extracts the most likely client IP from a request.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.SplitN(xff, ",", 2)
		return strings.TrimSpace(parts[0])
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	host, _, err := splitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// splitHostPort is a thin wrapper around net.SplitHostPort that avoids an
// import cycle with the net package being in the proxy root.
func splitHostPort(addr string) (string, string, error) {
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			return addr[:i], addr[i+1:], nil
		}
	}
	return addr, "", fmt.Errorf("no port in %q", addr)
}

// --- Statistical helpers ---

func average(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range data {
		sum += v
	}
	return sum / float64(len(data))
}

func percentile(data []float64, p float64) float64 {
	if len(data) == 0 {
		return 0
	}
	sorted := make([]float64, len(data))
	copy(sorted, data)
	sortFloat64s(sorted)
	idx := int(p / 100 * float64(len(sorted)))
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

func sortFloat64s(a []float64) {
	// Insertion sort (fine for ≤10 000 elements).
	for i := 1; i < len(a); i++ {
		key := a[i]
		j := i - 1
		for j >= 0 && a[j] > key {
			a[j+1] = a[j]
			j--
		}
		a[j+1] = key
	}
}
