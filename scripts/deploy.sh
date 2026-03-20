#!/usr/bin/env bash
# scripts/deploy.sh – Build and push Docker images, then update the stack.
#
# Required environment variables:
#   REGISTRY   – Container registry prefix (e.g. ghcr.io/Cipher7788/media-hub)
#   TAG        – Image tag (e.g. v1.2.3 or git SHA)
#
# Usage:
#   REGISTRY=ghcr.io/Cipher7788/media-hub TAG=v1.0.0 ./scripts/deploy.sh
set -euo pipefail

REGISTRY="${REGISTRY:?REGISTRY must be set}"
TAG="${TAG:?TAG must be set}"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo "==> Building proxy image..."
docker build \
    -f "$REPO_ROOT/docker/Dockerfile.proxy" \
    -t "${REGISTRY}/proxy:${TAG}" \
    -t "${REGISTRY}/proxy:latest" \
    "$REPO_ROOT"

echo "==> Building exit-node image..."
docker build \
    -f "$REPO_ROOT/docker/Dockerfile.exit-node" \
    -t "${REGISTRY}/exit-node:${TAG}" \
    -t "${REGISTRY}/exit-node:latest" \
    "$REPO_ROOT"

echo "==> Pushing images..."
docker push "${REGISTRY}/proxy:${TAG}"
docker push "${REGISTRY}/proxy:latest"
docker push "${REGISTRY}/exit-node:${TAG}"
docker push "${REGISTRY}/exit-node:latest"

echo "==> Deploying stack..."
TAG="${TAG}" docker compose -f "$REPO_ROOT/docker-compose.yml" up -d --remove-orphans

echo "Deploy complete: ${REGISTRY}/*:${TAG}"
