//! TCP relay logic for the exit node.
//!
//! A [`Relay`] accepts an inbound TCP connection and forwards its traffic
//! to a target address, bidirectionally. Each relay pair runs in two threads:
//! one copying from client → target, the other from target → client.
//!
//! The [`ConnectionLimiter`] tracks the number of active connections per
//! source IP and rejects new connections once the configured maximum is
//! reached.

use std::collections::HashMap;
use std::io;
use std::net::{IpAddr, SocketAddr, TcpStream};
use std::sync::{Arc, Mutex};
use std::time::{Duration, Instant};

// ── Connection Limiter ────────────────────────────────────────────────────────

/// Tracks active connection counts per source IP address and enforces a
/// per-IP limit to prevent resource exhaustion.
pub struct ConnectionLimiter {
    max_per_ip: usize,
    counts: Mutex<HashMap<IpAddr, usize>>,
}

impl ConnectionLimiter {
    /// Creates a new limiter that allows at most `max_per_ip` simultaneous
    /// connections from a single IP address.
    pub fn new(max_per_ip: usize) -> Arc<Self> {
        Arc::new(Self {
            max_per_ip,
            counts: Mutex::new(HashMap::new()),
        })
    }

    /// Attempts to register a new connection from `addr`.
    ///
    /// Returns `true` if the connection is accepted, `false` if the per-IP
    /// limit has been reached.
    pub fn acquire(&self, addr: IpAddr) -> bool {
        let mut counts = self.counts.lock().expect("limiter lock poisoned");
        let count = counts.entry(addr).or_insert(0);
        if *count >= self.max_per_ip {
            return false;
        }
        *count += 1;
        true
    }

    /// Releases a connection slot for `addr`. Must be called exactly once
    /// for each successful [`acquire`](Self::acquire).
    pub fn release(&self, addr: IpAddr) {
        let mut counts = self.counts.lock().expect("limiter lock poisoned");
        if let Some(count) = counts.get_mut(&addr) {
            if *count > 0 {
                *count -= 1;
            }
            if *count == 0 {
                counts.remove(&addr);
            }
        }
    }

    /// Returns the current connection count for `addr`.
    #[allow(dead_code)] // used in tests
    pub fn count(&self, addr: IpAddr) -> usize {
        let counts = self.counts.lock().expect("limiter lock poisoned");
        *counts.get(&addr).unwrap_or(&0)
    }
}

// ── Relay ─────────────────────────────────────────────────────────────────────

/// Statistics collected for a completed relay session.
#[derive(Debug, Default, Clone)]
pub struct RelayStats {
    /// Bytes copied from the client to the target.
    pub bytes_inbound: u64,
    /// Bytes copied from the target to the client.
    pub bytes_outbound: u64,
    /// Wall-clock duration of the relay session.
    pub duration: Duration,
}

/// Relays traffic between `client` and `target` bidirectionally.
///
/// Returns [`RelayStats`] once both directions have been closed.
///
/// # Errors
///
/// Returns an [`io::Error`] if either the outbound connection to `target`
/// or the initial stream configuration fails.
pub fn relay(client: TcpStream, target_addr: SocketAddr) -> io::Result<RelayStats> {
    let start = Instant::now();

    let target = TcpStream::connect_timeout(&target_addr, Duration::from_secs(10))?;

    // Enable TCP_NODELAY on both sides for lower latency.
    client.set_nodelay(true)?;
    target.set_nodelay(true)?;

    let client_r = client.try_clone()?;
    let client_w = client;
    let target_r = target.try_clone()?;
    let target_w = target;

    // Spawn two threads for the bidirectional copy.
    let inbound = std::thread::spawn(move || copy_stream(client_r, target_w));
    let outbound = std::thread::spawn(move || copy_stream(target_r, client_w));

    let bytes_inbound = inbound.join().unwrap_or(Ok(0)).unwrap_or(0);
    let bytes_outbound = outbound.join().unwrap_or(Ok(0)).unwrap_or(0);

    Ok(RelayStats {
        bytes_inbound,
        bytes_outbound,
        duration: start.elapsed(),
    })
}

/// Copies all bytes from `src` to `dst` and returns the number of bytes
/// transferred. The destination is shut down for writes when the source
/// reaches EOF.
fn copy_stream(mut src: TcpStream, mut dst: TcpStream) -> io::Result<u64> {
    let n = io::copy(&mut src, &mut dst)?;
    // Best-effort shutdown; ignore errors (the peer may have already closed).
    let _ = dst.shutdown(std::net::Shutdown::Write);
    Ok(n)
}

// ── Allowed-Target Validator ──────────────────────────────────────────────────

/// Validates that relay targets are within the permitted port range and not
/// on the loopback interface (to prevent SSRF-style attacks).
pub struct TargetValidator {
    allowed_ports: Vec<u16>,
}

impl TargetValidator {
    /// Creates a validator that only permits connections to the specified ports.
    pub fn new(allowed_ports: Vec<u16>) -> Self {
        Self { allowed_ports }
    }

    /// Returns `Ok(())` if `addr` is a permitted target, or an `Err` describing
    /// the policy violation.
    pub fn validate(&self, addr: &SocketAddr) -> Result<(), String> {
        let ip = addr.ip();

        // Refuse loopback and link-local addresses.
        if ip.is_loopback() {
            return Err(format!("target {ip} is a loopback address"));
        }
        if is_link_local(ip) {
            return Err(format!("target {ip} is a link-local address"));
        }
        // Refuse private ranges (RFC 1918 / RFC 4193).
        if is_private(ip) {
            return Err(format!("target {ip} is a private address"));
        }

        let port = addr.port();
        if !self.allowed_ports.is_empty() && !self.allowed_ports.contains(&port) {
            return Err(format!("target port {port} is not in the allowed list"));
        }

        Ok(())
    }
}

fn is_link_local(ip: IpAddr) -> bool {
    match ip {
        IpAddr::V4(v4) => v4.is_link_local(),
        IpAddr::V6(v6) => (v6.segments()[0] & 0xffc0) == 0xfe80,
    }
}

fn is_private(ip: IpAddr) -> bool {
    match ip {
        IpAddr::V4(v4) => v4.is_private(),
        IpAddr::V6(v6) => {
            // fc00::/7 (Unique Local Addresses)
            (v6.segments()[0] & 0xfe00) == 0xfc00
        }
    }
}

// ── Tests ─────────────────────────────────────────────────────────────────────

#[cfg(test)]
mod tests {
    use super::*;
    use std::net::{IpAddr, Ipv4Addr, SocketAddr};

    fn ipv4(a: u8, b: u8, c: u8, d: u8) -> IpAddr {
        IpAddr::V4(Ipv4Addr::new(a, b, c, d))
    }

    // ── ConnectionLimiter ────────────────────────────────────────────────────

    #[test]
    fn limiter_allows_up_to_max() {
        let lim = ConnectionLimiter::new(2);
        let ip = ipv4(10, 0, 0, 1);

        assert!(lim.acquire(ip), "first connection should be accepted");
        assert!(lim.acquire(ip), "second connection should be accepted");
        assert!(!lim.acquire(ip), "third connection should be rejected");
    }

    #[test]
    fn limiter_release_frees_slot() {
        let lim = ConnectionLimiter::new(1);
        let ip = ipv4(10, 0, 0, 2);

        assert!(lim.acquire(ip));
        assert!(!lim.acquire(ip)); // at max
        lim.release(ip);
        assert!(lim.acquire(ip)); // slot freed
    }

    #[test]
    fn limiter_count_tracks_correctly() {
        let lim = ConnectionLimiter::new(5);
        let ip = ipv4(10, 0, 0, 3);

        assert_eq!(lim.count(ip), 0);
        lim.acquire(ip);
        lim.acquire(ip);
        assert_eq!(lim.count(ip), 2);
        lim.release(ip);
        assert_eq!(lim.count(ip), 1);
    }

    #[test]
    fn limiter_cleans_up_zero_counts() {
        let lim = ConnectionLimiter::new(2);
        let ip = ipv4(10, 0, 0, 4);

        lim.acquire(ip);
        lim.release(ip);
        // After full release, the entry should be removed.
        assert_eq!(lim.count(ip), 0);
    }

    #[test]
    fn limiter_independent_ips() {
        let lim = ConnectionLimiter::new(1);
        let ip1 = ipv4(10, 0, 0, 5);
        let ip2 = ipv4(10, 0, 0, 6);

        assert!(lim.acquire(ip1));
        assert!(lim.acquire(ip2)); // different IP, own slot
        assert!(!lim.acquire(ip1)); // ip1 at max
    }

    // ── TargetValidator ──────────────────────────────────────────────────────

    fn validator() -> TargetValidator {
        TargetValidator::new(vec![80, 443])
    }

    fn sockaddr(a: u8, b: u8, c: u8, d: u8, port: u16) -> SocketAddr {
        SocketAddr::new(ipv4(a, b, c, d), port)
    }

    #[test]
    fn validator_accepts_public_allowed_port() {
        let v = validator();
        assert!(v.validate(&sockaddr(203, 0, 113, 1, 443)).is_ok());
        assert!(v.validate(&sockaddr(203, 0, 113, 1, 80)).is_ok());
    }

    #[test]
    fn validator_rejects_disallowed_port() {
        let v = validator();
        assert!(v.validate(&sockaddr(203, 0, 113, 1, 8080)).is_err());
    }

    #[test]
    fn validator_rejects_loopback() {
        let v = TargetValidator::new(vec![]);
        assert!(v.validate(&sockaddr(127, 0, 0, 1, 443)).is_err());
    }

    #[test]
    fn validator_rejects_private_rfc1918() {
        let v = TargetValidator::new(vec![]);
        for addr in &[
            sockaddr(10, 0, 0, 1, 443),
            sockaddr(172, 16, 0, 1, 443),
            sockaddr(192, 168, 1, 1, 443),
        ] {
            assert!(
                v.validate(addr).is_err(),
                "expected private address {addr} to be rejected"
            );
        }
    }

    #[test]
    fn validator_empty_allowed_ports_permits_any_public() {
        let v = TargetValidator::new(vec![]); // empty = no port restriction
        assert!(v.validate(&sockaddr(203, 0, 113, 1, 12345)).is_ok());
    }

    // ── is_private / is_link_local helpers ───────────────────────────────────

    #[test]
    fn private_ranges_detected() {
        assert!(is_private(ipv4(10, 1, 2, 3)));
        assert!(is_private(ipv4(172, 20, 0, 1)));
        assert!(is_private(ipv4(192, 168, 0, 1)));
        assert!(!is_private(ipv4(8, 8, 8, 8)));
    }
}
