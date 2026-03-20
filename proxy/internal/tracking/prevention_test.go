package tracking

import (
	"net/http"
	"testing"
)

func TestIsTracker(t *testing.T) {
	e := New()

	cases := []struct {
		host    string
		tracker bool
	}{
		{"doubleclick.net", true},
		{"sub.doubleclick.net", true},
		{"google.com", false},
		{"example.com", false},
		{"hotjar.com:443", true},
	}

	for _, c := range cases {
		got := e.IsTracker(c.host)
		if got != c.tracker {
			t.Errorf("IsTracker(%q) = %v, want %v", c.host, got, c.tracker)
		}
	}
}

func TestAddTracker(t *testing.T) {
	e := New()
	e.AddTracker("custom-tracker.example")

	if !e.IsTracker("custom-tracker.example") {
		t.Error("custom tracker should be detected after AddTracker")
	}
	if !e.IsTracker("api.custom-tracker.example") {
		t.Error("subdomain of custom tracker should be detected")
	}
}

func TestProcessBlocksTracker(t *testing.T) {
	e := New()

	req, _ := http.NewRequest(http.MethodGet, "https://googletagmanager.com/gtm.js", nil)
	req.Host = "googletagmanager.com"

	if err := e.Process(req); err == nil {
		t.Error("expected error when requesting tracker domain")
	}
}

func TestProcessStripsUTMParams(t *testing.T) {
	e := New()

	req, _ := http.NewRequest(http.MethodGet,
		"https://example.com/page?utm_source=newsletter&utm_campaign=spring&q=search", nil)

	if err := e.Process(req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	q := req.URL.Query()
	for _, p := range []string{"utm_source", "utm_campaign"} {
		if q.Has(p) {
			t.Errorf("tracking param %q should have been removed", p)
		}
	}
	if q.Get("q") != "search" {
		t.Error("non-tracking query param 'q' should be preserved")
	}
}

func TestProcessStripsFingerprintHeaders(t *testing.T) {
	e := New()

	req, _ := http.NewRequest(http.MethodGet, "https://example.com/", nil)
	req.Header.Set("Sec-CH-UA", `"Chromium";v="112"`)
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Accept", "text/html")

	if err := e.Process(req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Header.Get("Sec-CH-UA") != "" {
		t.Error("Sec-CH-UA should have been stripped")
	}
	if req.Header.Get("Sec-Fetch-Mode") != "" {
		t.Error("Sec-Fetch-Mode should have been stripped")
	}
	if req.Header.Get("Accept") != "text/html" {
		t.Error("Accept header should be preserved")
	}
}

func TestMiddleware(t *testing.T) {
	e := New()
	mw := e.Middleware()

	req, _ := http.NewRequest(http.MethodGet,
		"https://example.com/?fbclid=abc123&page=1", nil)

	if err := mw(req); err != nil {
		t.Fatalf("unexpected middleware error: %v", err)
	}
	if req.URL.Query().Has("fbclid") {
		t.Error("fbclid should have been stripped by middleware")
	}
}

func TestCleanURL(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{
			"https://example.com/?utm_source=test&q=hello",
			"https://example.com/?q=hello",
		},
		{
			"https://example.com/page",
			"https://example.com/page",
		},
		{
			"https://example.com/?gclid=CjwK&fbclid=abc",
			"https://example.com/",
		},
	}

	for _, c := range cases {
		got, err := CleanURL(c.in)
		if err != nil {
			t.Fatalf("CleanURL(%q) error: %v", c.in, err)
		}
		if got != c.want {
			t.Errorf("CleanURL(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestCleanURLInvalid(t *testing.T) {
	_, err := CleanURL("://bad url")
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}
