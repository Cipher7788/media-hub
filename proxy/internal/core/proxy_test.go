package core

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.ListenAddr != ":8080" {
		t.Errorf("expected ListenAddr :8080, got %s", cfg.ListenAddr)
	}
	if cfg.ReadTimeout != 30*time.Second {
		t.Errorf("expected ReadTimeout 30s, got %v", cfg.ReadTimeout)
	}
}

func TestVersion(t *testing.T) {
	v := Version()
	if v == "" {
		t.Error("Version() returned empty string")
	}
}

func TestRemoveHopHeaders(t *testing.T) {
	h := http.Header{}
	h.Set("Connection", "Keep-Alive, X-Custom")
	h.Set("Keep-Alive", "timeout=5")
	h.Set("X-Custom", "value")
	h.Set("Content-Type", "text/plain")

	removeHopHeaders(h)

	if h.Get("Keep-Alive") != "" {
		t.Error("Keep-Alive should have been removed")
	}
	if h.Get("X-Custom") != "" {
		t.Error("X-Custom (listed in Connection) should have been removed")
	}
	if h.Get("Content-Type") != "text/plain" {
		t.Error("Content-Type should not have been removed")
	}
}

func TestProxyHTTP(t *testing.T) {
	// Upstream server that just echoes a 200 OK.
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	proxy := New(DefaultConfig(), nil)

	req, err := http.NewRequest(http.MethodGet, upstream.URL+"/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.RequestURI = ""

	rr := httptest.NewRecorder()
	proxy.handleHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestProxyMiddlewareBlock(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	proxy := New(DefaultConfig(), nil)
	proxy.Use(func(r *http.Request) error {
		return http.ErrNoCookie // any non-nil error blocks
	})

	req, err := http.NewRequest(http.MethodGet, upstream.URL+"/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	proxy.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

func TestSplitComma(t *testing.T) {
	cases := []struct {
		in  string
		out []string
	}{
		{"a, b, c", []string{"a", "b", "c"}},
		{"single", []string{"single"}},
		{"", []string{""}},
	}
	for _, c := range cases {
		got := splitComma(c.in)
		if len(got) != len(c.out) {
			t.Errorf("splitComma(%q) len = %d, want %d", c.in, len(got), len(c.out))
			continue
		}
		for i := range got {
			if got[i] != c.out[i] {
				t.Errorf("splitComma(%q)[%d] = %q, want %q", c.in, i, got[i], c.out[i])
			}
		}
	}
}
