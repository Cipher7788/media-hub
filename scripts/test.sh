#!/usr/bin/env bash
# scripts/test.sh – Run all tests for every component.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo "==> Running Go proxy tests..."
cd "$REPO_ROOT/proxy"
go test -race -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | tail -1

echo ""
echo "==> Running Rust exit-node tests..."
cd "$REPO_ROOT/exit-node"
cargo test

echo ""
echo "==> Running React frontend tests..."
cd "$REPO_ROOT"
npm test -- --watchAll=false --passWithNoTests

echo ""
echo "All tests passed."
