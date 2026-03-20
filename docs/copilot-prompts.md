# GitHub Copilot Prompts for Each Phase

This document provides ready-to-use GitHub Copilot prompt starters for
continuing development on each phase of the proxy project.

---

## Phase 1 – Core Proxy Engine

```
// proxy/internal/core/proxy.go
// TODO: Add support for HTTP/2 upstream connections. The proxy should
// negotiate HTTP/2 with upstreams when the upstream advertises h2 via ALPN,
// while still supporting HTTP/1.1 clients.
```

```
// proxy/internal/core/proxy.go
// TODO: Implement connection pooling statistics: expose the number of idle
// connections, active connections, and total wait time via the metrics package.
```

```
// proxy/internal/core/proxy.go
// TODO: Add support for upstream proxy chaining – when an UPSTREAM_PROXY
// environment variable is set, route all outbound connections through it
// (useful for corporate networks).
```

---

## Phase 2 – Firewall & DPI

```
// proxy/internal/firewall/firewall.go
// TODO: Implement a domain block-list loader that reads a hosts-format file
// (e.g. from a URL like https://someblockist.txt) and registers a BlockDomain
// rule for each entry. Refresh the list every 24 hours.
```

```
// proxy/internal/firewall/firewall.go
// TODO: Extend the DPI inspector to detect TLS Client Hello fingerprinting
// (JA3 hashes). Log the JA3 hash for each CONNECT request so operators can
// identify unusual TLS client stacks.
```

---

## Phase 3 – Anonymisation Engine

```
// proxy/internal/anonymize/anonymizer.go
// TODO: Implement onion routing encryption for the circuit. Each relay node
// should receive only the address of the next hop, encrypted with that node's
// public key using X25519 key exchange and AES-256-GCM.
```

```
// proxy/internal/anonymize/anonymizer.go
// TODO: Add a CircuitManager that automatically rebuilds circuits every
// 10 minutes and on error, and exposes the current circuit hop count via
// the /metrics endpoint.
```

---

## Phase 4 – Tracking Prevention

```
// proxy/internal/tracking/prevention.go
// TODO: Integrate an automatically updated EasyList/EasyPrivacy block list.
// Parse the Adblock Plus filter syntax and convert domain rules into
// BlockDomain firewall rules.
```

```
// proxy/internal/tracking/prevention.go
// TODO: Implement canvas fingerprint injection – intercept responses that
// set Canvas API data and return randomised but visually identical pixel data
// to prevent canvas-based fingerprinting.
```

---

## Phase 5 – Monitoring & Threat Detection

```
// proxy/internal/monitor/monitor.go
// TODO: Add a sliding-window anomaly detector that uses exponential moving
// average (EMA) to detect sudden spikes in error rate or latency, and raises
// a ThreatMedium alert when the EMA exceeds 2 standard deviations.
```

```
// proxy/internal/monitor/monitor.go
// TODO: Implement an HTTP handler for /api/v1/threats that returns a JSON
// array of the last 100 threat alerts, including timestamp, source IP,
// threat level, and message.
```

---

## Exit Nodes (Rust)

```rust
// exit-node/src/relay.rs
// TODO: Add TLS termination to the exit node so that the proxy–exit-node
// channel is encrypted. Use rustls (no OpenSSL dependency) to accept a TLS
// connection from the proxy and forward plaintext to the target.
```

```rust
// exit-node/src/relay.rs
// TODO: Implement a health-check endpoint that listens on a separate port
// (default: 9091) and responds to GET /healthz with {"status":"ok"} and
// current connection count.
```

---

## Testing & Deployment

```
// scripts/test.sh
// TODO: Add an integration test that:
// 1. Starts the proxy on a random port using os/exec.
// 2. Configures curl to use it as an HTTP proxy.
// 3. Verifies that blocked domains return 403.
// 4. Verifies that tracking parameters are stripped from forwarded requests.
```

```yaml
# .github/workflows/ci.yml
# TODO: Add a job that builds and runs the Docker Compose stack, waits for
# the health endpoint to return 200, then runs the integration test suite.
```
