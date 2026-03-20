# Architecture Overview

## Executive Summary

This repository hosts a privacy-focused proxy system designed to protect users
from tracking, censorship, and traffic analysis. It consists of two compiled
services and a React front-end:

| Component | Language | Role |
|-----------|----------|------|
| Proxy engine | Go | HTTP/HTTPS proxy with firewall, DPI, anonymisation, and tracking prevention |
| Exit node | Rust | Last-hop TCP relay; validates and forwards traffic to the public internet |
| Media Hub UI | React | Personal media-collection web application (existing) |

---

## Stack Selection

### Go – Proxy Engine

Go was chosen for the core proxy because:

* Its `net/http` standard library provides a production-quality HTTP/1.1 and
  HTTP/2 stack with hijacking support (needed for CONNECT tunnels).
* Goroutines give very low overhead per connection.
* The language's strong typing and built-in race detector make concurrent code
  easier to reason about and test.

### Rust – Exit Node

Rust was chosen for the exit node because:

* Memory safety without a garbage collector eliminates entire classes of
  security vulnerabilities (use-after-free, buffer overflow) that are critical
  at a network trust boundary.
* The `std::net` TCP API is zero-cost and maps directly onto OS sockets.
* The release binary is small and self-contained with no runtime dependency.

---

## System Architecture

```
                        ┌──────────────────────────────────────────┐
                        │               Client Browser              │
                        └───────────────────┬──────────────────────┘
                                            │ HTTP / HTTPS CONNECT
                                            ▼
                        ┌──────────────────────────────────────────┐
                        │              Proxy Engine (Go)            │
                        │                                           │
                        │  ┌─────────┐  ┌─────┐  ┌────────────┐   │
                        │  │Firewall │  │ DPI │  │Anonymiser  │   │
                        │  │& Rules  │→ │     │→ │(header     │   │
                        │  └─────────┘  └─────┘  │ sanitise + │   │
                        │                        │ circuit)   │   │
                        │                        └──────┬─────┘   │
                        │  ┌─────────────┐             │          │
                        │  │  Tracking   │◄────────────┘          │
                        │  │  Prevention │                         │
                        │  └──────┬──────┘                        │
                        │         │                                 │
                        │  ┌──────▼──────┐                        │
                        │  │  Monitor /  │  Prometheus :9090       │
                        │  │  Threat Det │──────────────────────►  │
                        │  └─────────────┘                        │
                        └───────────────────┬──────────────────────┘
                                            │ TCP (relay protocol)
                                            ▼
                        ┌──────────────────────────────────────────┐
                        │              Exit Node (Rust)             │
                        │                                           │
                        │  ┌───────────┐  ┌───────────────────┐   │
                        │  │Connection │  │ Target Validator  │   │
                        │  │ Limiter   │  │ (no RFC1918/loop) │   │
                        │  └───────────┘  └──────────┬────────┘   │
                        │                            │             │
                        └────────────────────────────┼─────────────┘
                                                     │ TCP
                                                     ▼
                                          Public Internet Target
```

---

## Phase Descriptions

### Phase 1 – Core Proxy Engine (`proxy/internal/core`)

Implements `net/http`-based HTTP and HTTPS proxying:

* **Plain HTTP** – rewrites requests, strips hop-by-hop headers, and forwards
  via a shared `http.Transport`.
* **HTTPS tunnels** – handles `CONNECT` by hijacking the client connection and
  opening a raw TCP connection to the target.
* **Middleware chain** – all subsequent phases plug in via `Proxy.Use()`.

### Phase 2 – Firewall & DPI Layer (`proxy/internal/firewall`)

* **Firewall** – evaluates an ordered rule list (CIDR blocks, domain blocks,
  content-type rules) against every request.
* **DPI Inspector** – classifies application-layer protocols from HTTP headers
  and URI patterns; blocks disallowed protocols (e.g. BitTorrent).

### Phase 3 – Anonymisation Engine (`proxy/internal/anonymize`)

* **Header Sanitiser** – removes `X-Forwarded-For`, `Via`, `Referer`, and
  other headers that identify the client.
* **Circuit Builder** – selects a weighted-random ordered set of relay nodes
  to form an anonymisation circuit.

### Phase 4 – Tracking Prevention (`proxy/internal/tracking`)

* Blocks requests to a built-in list of known tracker and ad-network domains.
* Strips UTM, fbclid, gclid, and other tracking query parameters.
* Removes browser-fingerprinting headers (`Sec-CH-UA-*`, `Sec-Fetch-*`).

### Phase 5 – Monitoring & Threat Detection (`proxy/internal/monitor`)

* **Metrics** – Prometheus-compatible counters and histograms for request
  rate, latency, blocked requests, and bytes forwarded.
* **Threat Detector** – per-source IP rate limiting with four severity levels
  (Low / Medium / High / Critical) and pluggable alert handlers.

### Exit Nodes (`exit-node/src`)

* TCP relay that forwards traffic from the proxy circuit to the public internet.
* **Connection Limiter** – caps simultaneous connections per source IP.
* **Target Validator** – rejects loopback, link-local, and RFC 1918 targets.
