// Package firewall implements the firewall and Deep Packet Inspection (DPI)
// layer for the proxy engine.
//
// The firewall evaluates each request against a set of ordered rules. The
// DPI inspector analyses application-layer protocol characteristics to detect
// and optionally block tunnelled protocols (e.g. BitTorrent over HTTP).
package firewall

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
)

// Action specifies what the firewall should do when a rule matches.
type Action int

const (
	// Allow passes the request through.
	Allow Action = iota
	// Block drops the request with a 403 response.
	Block
	// Log allows the request but writes an entry to the audit log.
	Log
)

func (a Action) String() string {
	switch a {
	case Allow:
		return "ALLOW"
	case Block:
		return "BLOCK"
	case Log:
		return "LOG"
	default:
		return "UNKNOWN"
	}
}

// Rule is a single firewall rule that matches against an incoming request.
type Rule struct {
	// Name is a human-readable label used in audit logs.
	Name string
	// Action is what happens when the rule matches.
	Action Action
	// match is the internal predicate; it returns true when the rule applies.
	match func(r *http.Request) bool
}

// Firewall evaluates ordered Rules against each request.
type Firewall struct {
	mu          sync.RWMutex
	rules       []Rule
	defaultRule Action
}

// New creates a Firewall with the given default action (applied when no
// explicit rule matches).
func New(defaultAction Action) *Firewall {
	return &Firewall{defaultRule: defaultAction}
}

// AddRule appends a rule to the end of the evaluation chain.
func (fw *Firewall) AddRule(r Rule) {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	fw.rules = append(fw.rules, r)
}

// BlockDomain returns a Rule that blocks requests to a specific hostname.
func BlockDomain(domain string) Rule {
	domain = strings.ToLower(domain)
	return Rule{
		Name:   "block-domain:" + domain,
		Action: Block,
		match: func(r *http.Request) bool {
			host := strings.ToLower(hostWithoutPort(r.Host))
			return host == domain || strings.HasSuffix(host, "."+domain)
		},
	}
}

// BlockCIDR returns a Rule that blocks requests whose resolved target IP falls
// within the given CIDR range.
func BlockCIDR(cidr string) (Rule, error) {
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return Rule{}, fmt.Errorf("firewall: invalid CIDR %q: %w", cidr, err)
	}
	return Rule{
		Name:   "block-cidr:" + cidr,
		Action: Block,
		match: func(r *http.Request) bool {
			host := hostWithoutPort(r.Host)
			ips, err := net.LookupHost(host)
			if err != nil {
				return false
			}
			for _, rawIP := range ips {
				ip := net.ParseIP(rawIP)
				if ip != nil && network.Contains(ip) {
					return true
				}
			}
			return false
		},
	}, nil
}

// AllowDomain returns a Rule that explicitly allows requests to a hostname.
func AllowDomain(domain string) Rule {
	domain = strings.ToLower(domain)
	return Rule{
		Name:   "allow-domain:" + domain,
		Action: Allow,
		match: func(r *http.Request) bool {
			host := strings.ToLower(hostWithoutPort(r.Host))
			return host == domain || strings.HasSuffix(host, "."+domain)
		},
	}
}

// BlockContentType returns a Rule that blocks responses (or requests) with a
// specific Content-Type prefix (e.g. "application/x-bittorrent").
func BlockContentType(contentType string) Rule {
	contentType = strings.ToLower(contentType)
	return Rule{
		Name:   "block-content-type:" + contentType,
		Action: Block,
		match: func(r *http.Request) bool {
			ct := strings.ToLower(r.Header.Get("Content-Type"))
			return strings.HasPrefix(ct, contentType)
		},
	}
}

// Evaluate runs the firewall rule chain against r and returns the matching
// Action. Rules are evaluated in order; the first match wins.
func (fw *Firewall) Evaluate(r *http.Request) (Action, string) {
	fw.mu.RLock()
	defer fw.mu.RUnlock()

	for _, rule := range fw.rules {
		if rule.match(r) {
			return rule.Action, rule.Name
		}
	}
	return fw.defaultRule, "default"
}

// Middleware returns a core.RequestMiddleware that blocks requests denied by
// the firewall.
func (fw *Firewall) Middleware() func(r *http.Request) error {
	return func(r *http.Request) error {
		action, ruleName := fw.Evaluate(r)
		if action == Block {
			return fmt.Errorf("blocked by rule %q", ruleName)
		}
		return nil
	}
}

// --- DPI Inspector ---

// Protocol is an application-layer protocol detected by the DPI inspector.
type Protocol string

const (
	ProtocolHTTP       Protocol = "HTTP"
	ProtocolHTTPS      Protocol = "HTTPS"
	ProtocolBitTorrent Protocol = "BitTorrent"
	ProtocolP2P        Protocol = "P2P"
	ProtocolUnknown    Protocol = "Unknown"
)

// Inspector performs Deep Packet Inspection to identify application-layer
// protocols based on HTTP request headers and URI patterns.
type Inspector struct {
	blocked map[Protocol]bool
}

// NewInspector creates a DPI Inspector. Protocols listed in blocked will
// cause Inspect to return an error.
func NewInspector(blocked ...Protocol) *Inspector {
	m := make(map[Protocol]bool, len(blocked))
	for _, p := range blocked {
		m[p] = true
	}
	return &Inspector{blocked: m}
}

// Inspect detects the application-layer protocol of the request and returns
// an error if that protocol is in the block list.
func (ins *Inspector) Inspect(r *http.Request) (Protocol, error) {
	proto := ins.detect(r)
	if ins.blocked[proto] {
		return proto, fmt.Errorf("DPI: blocked protocol %q", proto)
	}
	return proto, nil
}

// detect uses header heuristics to classify the protocol.
func (ins *Inspector) detect(r *http.Request) Protocol {
	ua := strings.ToLower(r.UserAgent())
	ct := strings.ToLower(r.Header.Get("Content-Type"))
	uri := strings.ToLower(r.URL.Path)

	switch {
	case strings.Contains(ua, "bittorrent") || strings.Contains(ua, "libtorrent"):
		return ProtocolBitTorrent
	case strings.Contains(ct, "application/x-bittorrent"):
		return ProtocolBitTorrent
	case strings.HasSuffix(uri, ".torrent"):
		return ProtocolBitTorrent
	case strings.Contains(ua, "p2p") || strings.Contains(ua, "emule"):
		return ProtocolP2P
	case r.Method == http.MethodConnect:
		return ProtocolHTTPS
	default:
		return ProtocolHTTP
	}
}

// Middleware returns a RequestMiddleware that applies DPI inspection.
func (ins *Inspector) Middleware() func(r *http.Request) error {
	return func(r *http.Request) error {
		_, err := ins.Inspect(r)
		return err
	}
}

// hostWithoutPort strips the port from a host string.
func hostWithoutPort(host string) string {
	if h, _, err := net.SplitHostPort(host); err == nil {
		return h
	}
	return host
}
