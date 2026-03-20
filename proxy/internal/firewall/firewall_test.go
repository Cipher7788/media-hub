package firewall

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func makeRequest(t *testing.T, method, url string) *http.Request {
	t.Helper()
	r, err := http.NewRequest(method, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestActionString(t *testing.T) {
	cases := map[Action]string{
		Allow: "ALLOW",
		Block: "BLOCK",
		Log:   "LOG",
	}
	for action, want := range cases {
		if got := action.String(); got != want {
			t.Errorf("Action(%d).String() = %q, want %q", action, got, want)
		}
	}
}

func TestFirewallBlockDomain(t *testing.T) {
	fw := New(Allow)
	fw.AddRule(BlockDomain("evil.example.com"))

	tests := []struct {
		host   string
		wantOK bool
	}{
		{"http://evil.example.com/path", false},
		{"http://sub.evil.example.com/path", false},
		{"http://good.example.com/path", true},
	}

	for _, tt := range tests {
		req := makeRequest(t, http.MethodGet, tt.host)
		action, _ := fw.Evaluate(req)
		gotOK := action != Block
		if gotOK != tt.wantOK {
			t.Errorf("host %s: allowed=%v, want %v", tt.host, gotOK, tt.wantOK)
		}
	}
}

func TestFirewallAllowOverridesBlock(t *testing.T) {
	fw := New(Block) // default blocks everything
	fw.AddRule(AllowDomain("safe.example.com"))

	req := makeRequest(t, http.MethodGet, "http://safe.example.com/")
	action, _ := fw.Evaluate(req)
	if action != Allow {
		t.Errorf("expected Allow, got %v", action)
	}

	req2 := makeRequest(t, http.MethodGet, "http://other.example.com/")
	action2, _ := fw.Evaluate(req2)
	if action2 != Block {
		t.Errorf("expected Block for unknown domain, got %v", action2)
	}
}

func TestFirewallBlockContentType(t *testing.T) {
	fw := New(Allow)
	fw.AddRule(BlockContentType("application/x-bittorrent"))

	req := makeRequest(t, http.MethodPost, "http://example.com/upload")
	req.Header.Set("Content-Type", "application/x-bittorrent; charset=utf-8")

	action, _ := fw.Evaluate(req)
	if action != Block {
		t.Errorf("expected Block for BitTorrent content-type, got %v", action)
	}
}

func TestFirewallMiddleware(t *testing.T) {
	fw := New(Allow)
	fw.AddRule(BlockDomain("blocked.example.com"))
	mw := fw.Middleware()

	req := makeRequest(t, http.MethodGet, "http://blocked.example.com/")
	if err := mw(req); err == nil {
		t.Error("expected middleware to return error for blocked domain")
	}

	req2 := makeRequest(t, http.MethodGet, "http://allowed.example.com/")
	if err := mw(req2); err != nil {
		t.Errorf("unexpected error for allowed domain: %v", err)
	}
}

func TestInvalidCIDR(t *testing.T) {
	_, err := BlockCIDR("not-a-cidr")
	if err == nil {
		t.Error("expected error for invalid CIDR")
	}
}

func TestDPIInspectorHTTP(t *testing.T) {
	ins := NewInspector()
	req := makeRequest(t, http.MethodGet, "http://example.com/page")
	proto, err := ins.Inspect(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if proto != ProtocolHTTP {
		t.Errorf("expected HTTP, got %v", proto)
	}
}

func TestDPIInspectorBitTorrentBlocked(t *testing.T) {
	ins := NewInspector(ProtocolBitTorrent)
	req := makeRequest(t, http.MethodGet, "http://tracker.example.com/file.torrent")
	_, err := ins.Inspect(req)
	if err == nil {
		t.Error("expected DPI to block BitTorrent")
	}
}

func TestDPIInspectorConnect(t *testing.T) {
	ins := NewInspector()
	req := makeRequest(t, http.MethodConnect, "https://secure.example.com:443")
	proto, err := ins.Inspect(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if proto != ProtocolHTTPS {
		t.Errorf("expected HTTPS for CONNECT, got %v", proto)
	}
}

func TestDPIMiddleware(t *testing.T) {
	ins := NewInspector(ProtocolBitTorrent)
	mw := ins.Middleware()

	req := makeRequest(t, http.MethodGet, "http://example.com/file.torrent")
	if err := mw(req); err == nil {
		t.Error("expected DPI middleware to block torrent request")
	}
}

func TestHostWithoutPort(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"example.com:8080", "example.com"},
		{"example.com", "example.com"},
		{"[::1]:443", "::1"},
	}
	for _, c := range cases {
		got := hostWithoutPort(c.in)
		if got != c.want {
			t.Errorf("hostWithoutPort(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// Ensure the recorder test helper works as expected.
func TestRecorder(t *testing.T) {
	_ = httptest.NewRecorder()
}
