#!/usr/bin/env bash
# Run median v0.6.0 E2E against Docker RustFS + Docker-managed POSIX bind mount.
# Usage: ./run.sh [all|go|js]
set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
REPO="$(cd "$ROOT/.." && pwd)"
COMPOSE=(docker compose -f "$ROOT/docker-compose.yml" -f "$ROOT/docker-compose.bind.yml")

export MEDIAN_E2E_S3_ACCESS_KEY="${MEDIAN_E2E_S3_ACCESS_KEY:-median_e2e}"
export MEDIAN_E2E_S3_SECRET_KEY="${MEDIAN_E2E_S3_SECRET_KEY:-median_e2e_secret}"
export MEDIAN_E2E_S3_BUCKET="${MEDIAN_E2E_S3_BUCKET:-median-e2e}"
export MEDIAN_E2E_S3_REGION="${MEDIAN_E2E_S3_REGION:-us-east-1}"
export MEDIAN_E2E_S3_ENDPOINT="${MEDIAN_E2E_S3_ENDPOINT:-http://127.0.0.1:9000}"

HOST_LOCAL="${MEDIAN_E2E_HOST_LOCAL:-$ROOT/.local-data}"
mkdir -p "$HOST_LOCAL"
chmod 777 "$HOST_LOCAL" || true
export MEDIAN_E2E_LOCAL_ROOT="${MEDIAN_E2E_LOCAL_ROOT:-$HOST_LOCAL}"
export MEDIAN_E2E_HOST_LOCAL="$HOST_LOCAL"

mode="${1:-all}"
KEEP="${MEDIAN_E2E_KEEP:-0}"

cleanup() {
  if [[ "$KEEP" != "1" ]]; then
    "${COMPOSE[@]}" down -v --remove-orphans >/dev/null 2>&1 || true
    rm -rf "$HOST_LOCAL"
  fi
}
trap cleanup EXIT

# Generate bind override so Local adapter uses a path shared with the localfs service.
cat > "$ROOT/docker-compose.bind.yml" <<EOF
services:
  localfs:
    volumes:
      - ${HOST_LOCAL}:/data/local
EOF

echo "==> starting rustfs + localfs (bind $HOST_LOCAL)"
"${COMPOSE[@]}" up -d --wait rustfs localfs

echo "==> waiting for RustFS health at $MEDIAN_E2E_S3_ENDPOINT"
for i in $(seq 1 60); do
  if curl -fsS "$MEDIAN_E2E_S3_ENDPOINT/health" >/dev/null 2>&1; then
    break
  fi
  sleep 1
  if [[ "$i" -eq 60 ]]; then
    echo "RustFS did not become healthy" >&2
    "${COMPOSE[@]}" logs rustfs >&2 || true
    exit 1
  fi
done

run_go() {
  echo "==> Go E2E"
  (
    cd "$ROOT/go"
    export GOTOOLCHAIN="${GOTOOLCHAIN:-auto}"
    go test -count=1 -timeout 15m ./...
  )
}

run_js() {
  echo "==> JS E2E"
  (
    cd "$REPO/packages/js"
    if [[ ! -d node_modules ]]; then
      npm ci
    fi
    npm run build
    cd "$ROOT/js"
    npm install
    npm test
  )
}

case "$mode" in
  go) run_go ;;
  js) run_js ;;
  all)
    run_go
    run_js
    ;;
  *)
    echo "usage: $0 [all|go|js]" >&2
    exit 2
    ;;
esac

echo "==> E2E OK"
