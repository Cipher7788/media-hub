//! Exit node binary for the privacy proxy.
//!
//! The exit node accepts inbound TCP connections from the proxy circuit and
//! forwards them to their final destination. It enforces:
//!
//! * A per-IP connection limit (default: 50 simultaneous connections per IP).
//! * An allow-list of permitted target ports (default: 80 and 443).
//! * A target address validator that rejects RFC 1918 / loopback destinations.
//!
//! # Usage
//!
//! ```text
//! exit-node [options]
//!   --addr   <host:port>   Listen address           (default: 0.0.0.0:4001)
//!   --max    <n>           Max connections per IP    (default: 50)
//! ```

mod relay;

use relay::{ConnectionLimiter, TargetValidator};
use std::io::{self, BufRead, Write};
use std::net::{SocketAddr, TcpListener, TcpStream};
use std::sync::Arc;
use std::thread;

/// Default listen address for the exit node.
const DEFAULT_ADDR: &str = "0.0.0.0:4001";
/// Default maximum simultaneous connections per source IP.
const DEFAULT_MAX_CONN_PER_IP: usize = 50;

fn main() -> io::Result<()> {
    let addr = DEFAULT_ADDR;
    let max_conn = DEFAULT_MAX_CONN_PER_IP;

    let limiter = ConnectionLimiter::new(max_conn);
    let validator = Arc::new(TargetValidator::new(vec![80, 443]));

    let listener = TcpListener::bind(addr)?;
    eprintln!("exit-node listening on {addr}");

    for stream in listener.incoming() {
        match stream {
            Ok(client) => {
                let peer = client.peer_addr()?;
                let limiter = Arc::clone(&limiter);
                let validator = Arc::clone(&validator);

                thread::spawn(move || {
                    if let Err(e) = handle_client(client, peer, limiter, validator) {
                        eprintln!("client {peer} error: {e}");
                    }
                });
            }
            Err(e) => eprintln!("accept error: {e}"),
        }
    }

    Ok(())
}

/// Handles a single inbound connection.
///
/// The client must send a single line containing the target address in
/// `host:port` format, then the exit node opens a connection to that target
/// and relays traffic bidirectionally.
fn handle_client(
    mut client: TcpStream,
    peer: SocketAddr,
    limiter: Arc<ConnectionLimiter>,
    validator: Arc<TargetValidator>,
) -> io::Result<()> {
    let ip = peer.ip();

    if !limiter.acquire(ip) {
        let _ = client.write_all(b"ERR connection limit exceeded\n");
        return Ok(());
    }

    let result = (|| -> io::Result<()> {
        // Read the target address from the first line.
        let mut buf_reader = io::BufReader::new(client.try_clone()?);
        let mut target_line = String::new();
        buf_reader.read_line(&mut target_line)?;
        let target_addr: SocketAddr = target_line
            .trim()
            .parse()
            .map_err(|e| io::Error::new(io::ErrorKind::InvalidInput, e))?;

        // Validate the target address.
        if let Err(reason) = validator.validate(&target_addr) {
            let _ = client.write_all(format!("ERR {reason}\n").as_bytes());
            return Err(io::Error::new(io::ErrorKind::PermissionDenied, reason));
        }

        client.write_all(b"OK\n")?;

        let stats = relay::relay(client, target_addr)?;
        eprintln!(
            "relay {peer} → {target_addr}: in={} out={} dur={:.1}s",
            stats.bytes_inbound,
            stats.bytes_outbound,
            stats.duration.as_secs_f64()
        );
        Ok(())
    })();

    limiter.release(ip);
    result
}
