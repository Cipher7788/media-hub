package monitor

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestThreatLevelString(t *testing.T) {
	cases := map[ThreatLevel]string{
		ThreatNone:     "NONE",
		ThreatLow:      "LOW",
		ThreatMedium:   "MEDIUM",
		ThreatHigh:     "HIGH",
		ThreatCritical: "CRITICAL",
	}
	for level, want := range cases {
		if got := level.String(); got != want {
			t.Errorf("ThreatLevel(%d).String() = %q, want %q", level, got, want)
		}
	}
}

func TestMetricsRecordAndSnapshot(t *testing.T) {
	m := NewMetrics()
	m.RecordRequest(10.5, false, 1024)
	m.RecordRequest(20.0, true, 2048)
	m.RecordTunnel()
	m.RecordError()

	s := m.Snapshot()
	if s.RequestsTotal != 2 {
		t.Errorf("RequestsTotal = %d, want 2", s.RequestsTotal)
	}
	if s.RequestsBlocked != 1 {
		t.Errorf("RequestsBlocked = %d, want 1", s.RequestsBlocked)
	}
	if s.BytesForwarded != 3072 {
		t.Errorf("BytesForwarded = %d, want 3072", s.BytesForwarded)
	}
	if s.Tunnels != 1 {
		t.Errorf("Tunnels = %d, want 1", s.Tunnels)
	}
	if s.Errors != 1 {
		t.Errorf("Errors = %d, want 1", s.Errors)
	}
	if s.AvgLatencyMS == 0 {
		t.Error("AvgLatencyMS should be non-zero")
	}
}

func TestMetricsPrometheusOutput(t *testing.T) {
	m := NewMetrics()
	m.RecordRequest(5.0, false, 512)

	var sb strings.Builder
	m.WritePrometheus(&sb)
	out := sb.String()

	for _, expected := range []string{
		"proxy_requests_total",
		"proxy_bytes_forwarded_total",
		"proxy_latency_avg_ms",
	} {
		if !strings.Contains(out, expected) {
			t.Errorf("Prometheus output missing %q", expected)
		}
	}
}

func TestMetricsHandler(t *testing.T) {
	m := NewMetrics()
	srv := httptest.NewServer(m.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "text/plain") {
		t.Errorf("unexpected Content-Type: %s", ct)
	}
}

func TestDetectorRateLimit(t *testing.T) {
	var alerts []Alert
	d := NewDetector(1*time.Minute, 10, func(a Alert) {
		alerts = append(alerts, a)
	}, nil)

	source := "192.168.1.1"

	// Send 10 requests — should hit ThreatMedium threshold.
	for i := 0; i < 10; i++ {
		d.Observe(source)
	}

	if len(alerts) == 0 {
		t.Error("expected at least one alert after exceeding threshold")
	}
	if alerts[len(alerts)-1].Level < ThreatMedium {
		t.Errorf("expected at least ThreatMedium, got %s", alerts[len(alerts)-1].Level)
	}
}

func TestDetectorCritical(t *testing.T) {
	d := NewDetector(1*time.Minute, 5, nil, nil)

	source := "10.0.0.1"
	var lastLevel ThreatLevel
	for i := 0; i < 30; i++ {
		lastLevel = d.Observe(source)
	}

	if lastLevel != ThreatCritical {
		t.Errorf("expected ThreatCritical after 30 requests, got %s", lastLevel)
	}
}

func TestDetectorMiddlewareBlocks(t *testing.T) {
	d := NewDetector(1*time.Minute, 2, nil, nil)

	req, _ := http.NewRequest(http.MethodGet, "http://example.com/", nil)
	req.RemoteAddr = "192.168.1.100:12345"

	mw := d.Middleware()
	// First 4 requests (2x threshold = ThreatHigh).
	var lastErr error
	for i := 0; i < 5; i++ {
		lastErr = mw(req)
	}

	if lastErr == nil {
		t.Error("expected middleware to block after rate limit exceeded")
	}
}

func TestAverage(t *testing.T) {
	if average(nil) != 0 {
		t.Error("average(nil) should return 0")
	}
	if got := average([]float64{10, 20, 30}); got != 20 {
		t.Errorf("average = %v, want 20", got)
	}
}

func TestPercentile(t *testing.T) {
	if percentile(nil, 95) != 0 {
		t.Error("percentile(nil, 95) should return 0")
	}
	data := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	p95 := percentile(data, 95)
	if p95 < 9 {
		t.Errorf("p95 = %v, expected >= 9", p95)
	}
}

func TestClientIP(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "http://example.com/", nil)
	req.RemoteAddr = "10.0.0.1:12345"

	if ip := clientIP(req); ip != "10.0.0.1" {
		t.Errorf("clientIP = %q, want %q", ip, "10.0.0.1")
	}

	req.Header.Set("X-Forwarded-For", "203.0.113.5, 192.168.0.1")
	if ip := clientIP(req); ip != "203.0.113.5" {
		t.Errorf("clientIP with XFF = %q, want %q", ip, "203.0.113.5")
	}
}
