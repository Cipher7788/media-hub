package anonymize

import (
	"net/http"
	"testing"
)

func TestDefaultSanitiser(t *testing.T) {
	s := DefaultSanitiser()
	if !s.ReplaceUserAgent {
		t.Error("expected ReplaceUserAgent=true")
	}
	if !s.StripReferer {
		t.Error("expected StripReferer=true")
	}
}

func TestSanitiserRemovesIdentifyingHeaders(t *testing.T) {
	s := DefaultSanitiser()

	req, _ := http.NewRequest(http.MethodGet, "http://example.com/", nil)
	req.Header.Set("X-Forwarded-For", "1.2.3.4")
	req.Header.Set("Via", "1.1 proxy.example.com")
	req.Header.Set("Forwarded", "for=1.2.3.4")
	req.Header.Set("User-Agent", "RealBrowser/99.0")
	req.Header.Set("Referer", "https://private.example.com/")

	s.Sanitise(req)

	for _, h := range []string{"X-Forwarded-For", "Via", "Forwarded"} {
		if req.Header.Get(h) != "" {
			t.Errorf("header %q should have been removed", h)
		}
	}
	if req.Header.Get("User-Agent") != genericUserAgent {
		t.Errorf("User-Agent should be replaced with generic value, got %q",
			req.Header.Get("User-Agent"))
	}
	if req.Header.Get("Referer") != "" {
		t.Error("Referer should have been removed")
	}
}

func TestSanitiserCookieStripping(t *testing.T) {
	s := DefaultSanitiser()
	s.StripCookies = true

	req, _ := http.NewRequest(http.MethodGet, "http://example.com/", nil)
	req.Header.Set("Cookie", "session=abc123")

	s.Sanitise(req)

	if req.Header.Get("Cookie") != "" {
		t.Error("Cookie should have been stripped")
	}
}

func TestSanitiserMiddleware(t *testing.T) {
	s := DefaultSanitiser()
	mw := s.Middleware()

	req, _ := http.NewRequest(http.MethodGet, "http://example.com/", nil)
	req.Header.Set("X-Real-IP", "10.0.0.1")

	if err := mw(req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Header.Get("X-Real-IP") != "" {
		t.Error("X-Real-IP should have been removed by middleware")
	}
}

func TestCircuitBuilderNotEnoughNodes(t *testing.T) {
	cb := NewCircuitBuilder(3)
	cb.RegisterNode(Node{Addr: "relay1:4001", Weight: 1})
	cb.RegisterNode(Node{Addr: "relay2:4001", Weight: 1})

	_, err := cb.Build()
	if err == nil {
		t.Error("expected error when fewer nodes than hops")
	}
}

func TestCircuitBuilderSuccess(t *testing.T) {
	cb := NewCircuitBuilder(2)
	nodes := []Node{
		{Addr: "relay1:4001", Weight: 1},
		{Addr: "relay2:4001", Weight: 2},
		{Addr: "relay3:4001", Weight: 1},
	}
	for _, n := range nodes {
		cb.RegisterNode(n)
	}

	c, err := cb.Build()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Len() != 2 {
		t.Errorf("expected circuit length 2, got %d", c.Len())
	}

	// Verify nodes are distinct.
	h0, _ := c.Hop(0)
	h1, _ := c.Hop(1)
	if h0.Addr == h1.Addr {
		t.Error("circuit hops should be distinct")
	}
}

func TestCircuitHopOutOfRange(t *testing.T) {
	cb := NewCircuitBuilder(1)
	cb.RegisterNode(Node{Addr: "relay1:4001"})

	c, err := cb.Build()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Hop(5); err == nil {
		t.Error("expected out-of-range error")
	}
}

func TestFormatCircuit(t *testing.T) {
	cb := NewCircuitBuilder(2)
	cb.RegisterNode(Node{Addr: "relay1:4001"})
	cb.RegisterNode(Node{Addr: "relay2:4001"})

	c, _ := cb.Build()
	desc := FormatCircuit(c)
	if desc == "" {
		t.Error("FormatCircuit returned empty string")
	}

	if FormatCircuit(nil) == "" {
		t.Error("FormatCircuit(nil) should return a non-empty placeholder")
	}
}

func TestWeightedSampleDistribution(t *testing.T) {
	nodes := []Node{
		{Addr: "heavy", Weight: 100},
		{Addr: "light", Weight: 1},
	}
	counts := map[string]int{}
	for i := 0; i < 200; i++ {
		result := weightedSample(nodes, 1)
		if len(result) > 0 {
			counts[result[0].Addr]++
		}
	}
	// With weight 100 vs 1, heavy should be selected significantly more often.
	if counts["heavy"] <= counts["light"] {
		t.Errorf("weighted sampling broken: heavy=%d, light=%d",
			counts["heavy"], counts["light"])
	}
}
