#!/usr/bin/env bash
set -Eeuo pipefail
IMAGE="${RUNTIME_SMOKE_IMAGE:-fitness-runtime-smoke:local}"
CONTAINER="${RUNTIME_SMOKE_CONTAINER:-fitness-runtime-smoke}"
PORT="${RUNTIME_SMOKE_PORT:-18080}"
cleanup() {
  docker rm -f "$CONTAINER" >/dev/null 2>&1 || true
}
trap cleanup EXIT INT TERM
cleanup
docker build --file benzhi.Dockerfile --tag "$IMAGE" .
docker run --rm --name "$CONTAINER" --publish "${PORT}:8080" "$IMAGE"
