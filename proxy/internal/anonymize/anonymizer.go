// Package anonymize implements the anonymisation engine for the proxy.
//
// The engine provides two complementary capabilities:
//
//  1. Header Sanitisation – strips or replaces HTTP headers that can reveal
//     the client's identity, software, or network position.
//
//  2. Circuit Routing – forwards requests through a chain of relay nodes so
//     that no single node observes both the origin and the destination.
package anonymize

import (
	"fmt"
	"math/rand/v2"
	"net/http"
	"strings"
	"sync"
)

// identifyingHeaders is the set of request headers stripped during
// sanitisation because they can reveal client identity or capabilities.
var identifyingHeaders = []string{
	"X-Forwarded-For",
	"X-Real-IP",
	"X-Client-IP",
	"X-Originating-IP",
	"Via",
	"Forwarded",
	"True-Client-IP",
	"CF-Connecting-IP",
	"Fastly-Client-IP",
	"X-ProxyUser-IP",
}

// genericUserAgent is substituted for the real User-Agent to prevent browser
// fingerprinting.
const genericUserAgent = "Mozilla/5.0 (compatible; AnonymousProxy/1.0)"

// Sanitiser removes identifying HTTP headers from outgoing requests.
type Sanitiser struct {
	// ReplaceUserAgent, when true, overwrites the User-Agent header with a
	// generic value.
	ReplaceUserAgent bool
	// StripReferer, when true, removes the Referer header.
	StripReferer bool
	// StripCookies, when true, removes Cookie headers (useful in explicit
	// privacy mode; not default because it breaks most sites).
	StripCookies bool
}

// DefaultSanitiser returns a Sanitiser with safe defaults.
func DefaultSanitiser() *Sanitiser {
	return &Sanitiser{
		ReplaceUserAgent: true,
		StripReferer:     true,
		StripCookies:     false,
	}
}

// Sanitise modifies r in-place, removing identifying headers.
func (s *Sanitiser) Sanitise(r *http.Request) {
	for _, h := range identifyingHeaders {
		r.Header.Del(h)
	}
	if s.ReplaceUserAgent {
		r.Header.Set("User-Agent", genericUserAgent)
	}
	if s.StripReferer {
		r.Header.Del("Referer")
	}
	if s.StripCookies {
		r.Header.Del("Cookie")
	}
}

// Middleware returns a RequestMiddleware that sanitises each request.
func (s *Sanitiser) Middleware() func(r *http.Request) error {
	return func(r *http.Request) error {
		s.Sanitise(r)
		return nil
	}
}

// --- Circuit Routing ---

// Node represents a relay node in an anonymisation circuit.
type Node struct {
	// Addr is the TCP address of the relay (host:port).
	Addr string
	// Weight influences how often this node is selected; higher is more likely.
	Weight int
}

// Circuit is an ordered sequence of relay nodes. Traffic passes through each
// node in turn before reaching the destination.
type Circuit struct {
	nodes []Node
}

// Hop returns the n-th relay in the circuit (0-indexed).
func (c *Circuit) Hop(n int) (Node, error) {
	if n < 0 || n >= len(c.nodes) {
		return Node{}, fmt.Errorf("anonymize: hop %d out of range (circuit len %d)", n, len(c.nodes))
	}
	return c.nodes[n], nil
}

// Len returns the number of hops in the circuit.
func (c *Circuit) Len() int { return len(c.nodes) }

// CircuitBuilder constructs anonymisation circuits from a pool of relay nodes.
type CircuitBuilder struct {
	mu    sync.RWMutex
	nodes []Node
	hops  int
}

// NewCircuitBuilder creates a builder that will select hops relays per circuit
// from the registered node pool.
func NewCircuitBuilder(hops int) *CircuitBuilder {
	if hops < 1 {
		hops = 3
	}
	return &CircuitBuilder{hops: hops}
}

// RegisterNode adds a relay node to the pool.
func (cb *CircuitBuilder) RegisterNode(n Node) {
	if n.Weight < 1 {
		n.Weight = 1
	}
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.nodes = append(cb.nodes, n)
}

// Build selects hops distinct relay nodes at random (weighted) and returns a
// Circuit. Returns an error if the pool has fewer nodes than required hops.
func (cb *CircuitBuilder) Build() (*Circuit, error) {
	cb.mu.RLock()
	pool := make([]Node, len(cb.nodes))
	copy(pool, cb.nodes)
	cb.mu.RUnlock()

	if len(pool) < cb.hops {
		return nil, fmt.Errorf("anonymize: circuit requires %d hops but only %d nodes available",
			cb.hops, len(pool))
	}

	selected := weightedSample(pool, cb.hops)
	return &Circuit{nodes: selected}, nil
}

// weightedSample returns n distinct nodes chosen proportionally to their weight.
func weightedSample(nodes []Node, n int) []Node {
	remaining := make([]Node, len(nodes))
	copy(remaining, nodes)

	result := make([]Node, 0, n)
	for len(result) < n && len(remaining) > 0 {
		total := 0
		for _, nd := range remaining {
			total += nd.Weight
		}
		r := rand.IntN(total)
		cumulative := 0
		for i, nd := range remaining {
			cumulative += nd.Weight
			if r < cumulative {
				result = append(result, nd)
				remaining = append(remaining[:i], remaining[i+1:]...)
				break
			}
		}
	}
	return result
}

// FormatCircuit returns a human-readable description of the circuit for
// debugging (never log this in production as it reveals topology).
func FormatCircuit(c *Circuit) string {
	if c == nil {
		return "<nil circuit>"
	}
	parts := make([]string, c.Len())
	for i := 0; i < c.Len(); i++ {
		node, _ := c.Hop(i)
		parts[i] = node.Addr
	}
	return strings.Join(parts, " → ")
}
