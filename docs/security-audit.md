# Security Audit & Threat Model

## Scope

This document covers the proxy engine (Go) and exit node (Rust) components.
The React media-hub front-end is out of scope for this audit.

---

## Threat Model (STRIDE)

| Threat | Asset | Mitigation |
|--------|-------|-----------|
| **S**poofing | Client identity | Header sanitiser strips `X-Forwarded-For`, `Via`, real User-Agent |
| **T**ampering | Request in transit | TLS enforced on exit-node listen port; CONNECT tunnel preserved |
| **R**epudiation | Proxy activity | Structured JSON audit log with per-request records |
| **I**nformation disclosure | Origin IP | Circuit routing obscures origin; headers stripped at proxy layer |
| **D**enial of service | Proxy / exit node | Rate limiter, connection limiter, read/write timeouts |
| **E**scalation of privilege | Internal networks | Firewall blocks RFC 1918 & loopback; exit node target validator |

---

## Trust Boundaries

```
Client → [Proxy Engine] → [Exit Node] → Public Internet
              ↑                ↑
        (untrusted)    (semi-trusted relay)
```

* **Client** – Fully untrusted. All input is treated as potentially malicious.
* **Proxy Engine** – Trusted execution environment. Enforces all policy.
* **Exit Node** – Trusted but minimal. Only forwards; does not inspect payload.
* **Relay Nodes** – Should be treated as semi-trusted; they observe encrypted
  traffic but not both endpoints simultaneously.

---

## Known Risks & Mitigations

### SSRF (Server-Side Request Forgery)

**Risk**: A malicious client could direct the proxy to make requests to
internal services (e.g. cloud metadata APIs at `169.254.169.254`).

**Mitigations**:
1. `BlockCIDR` rules in the firewall cover RFC 1918 and loopback ranges.
2. The exit-node `TargetValidator` rejects private, loopback, and link-local
   destinations at the relay layer.
3. Cloud metadata CIDR `169.254.169.254/32` should be added to firewall rules
   in cloud deployments.

### DNS Rebinding

**Risk**: Attacker controls DNS for a domain that initially resolves to a
public IP but rebinds to a private IP after the firewall check.

**Mitigation**: The firewall's CIDR rules resolve DNS at rule evaluation time.
A DNS-pinning cache or DNSSEC validation should be added in production.
This is marked as a **known limitation**.

### Traffic Analysis

**Risk**: An observer on the network can correlate request timing and size
to deanonymise users even when circuit routing is used.

**Mitigation**: Circuit routing with multiple hops reduces the probability of
a single observer seeing both endpoints. Full traffic-analysis resistance
(padding, mix nets) is out of scope for the current implementation.

### Certificate Pinning Bypass

**Risk**: The proxy intercepts CONNECT tunnels but does not perform MITM TLS
inspection, so it cannot detect certificate spoofing by the upstream.

**Mitigation**: The proxy operates as a transparent tunnel for HTTPS; TLS
validation is left to the client browser, which is the correct design for a
privacy proxy that should not break end-to-end encryption.

### Credential Exposure in Logs

**Risk**: Proxy logs could contain sensitive URL query parameters (e.g. API
keys passed in the URL).

**Mitigation**: The structured logger only records `method`, `host`, and
`path`. Query strings are never logged. Tracking parameters are stripped
before forwarding by the tracking prevention layer.

---

## Security Controls Checklist

- [x] Hop-by-hop headers removed from forwarded requests
- [x] RFC 1918 / loopback ranges blocked by firewall
- [x] BitTorrent and P2P protocols blocked by DPI
- [x] Identifying headers stripped by anonymisation engine
- [x] Tracking query parameters stripped
- [x] Browser fingerprinting headers removed
- [x] Per-IP rate limiting with automatic blocking at High/Critical threat
- [x] Per-IP connection limit at exit node
- [x] Exit node validates targets (no private / loopback addresses)
- [x] Read/write/idle timeouts on all TCP connections
- [x] TCP_NODELAY set on relay connections
- [x] Structured JSON logging (no sensitive data)
- [x] Non-root user in Docker containers
- [x] Minimal base image (`scratch` for proxy, `debian:bookworm-slim` for exit node)
- [x] CGO disabled in Go build (`CGO_ENABLED=0`)
- [ ] DNS-pinning cache (future work)
- [ ] Traffic padding / timing obfuscation (future work)
- [ ] mTLS between proxy and exit node (future work)

---

## Dependency Audit

### Go (proxy engine)

The proxy engine has **zero external dependencies** beyond the Go standard
library. This minimises the supply-chain attack surface.

### Rust (exit node)

The exit node has **zero external crate dependencies**. All functionality is
implemented using `std`.

---

## Responsible Disclosure

If you discover a security vulnerability, please open a private security
advisory via GitHub rather than a public issue.
