#!/usr/bin/env bash
# Build this checkout and run sanitized fixtures; never loads a kubeconfig.
set -euo pipefail
cd "$(dirname "$0")/.."
port="${1:-8185}"
if [[ ! "$port" =~ ^[0-9]+$ ]] || (( port < 1024 || port > 65535 )); then
  echo 'Usage: scripts/preview-local.sh [port 1024..65535]' >&2
  exit 2
fi
(cd frontend && npm run build -- --emptyOutDir)
mkdir -p bin
export GOCACHE="${GOCACHE:-${TMPDIR:-/tmp}/ironcore-dashboard-go-cache}"
go build -o bin/ironcore-dashboard-preview ./cmd/server
printf '\nLocal fixture preview (no cluster connection): http://127.0.0.1:%s/overview\nSRE Console: http://127.0.0.1:%s/sre/landscape\n\n' "$port" "$port"
exec bin/ironcore-dashboard-preview --addr "127.0.0.1:$port" --sre-fixture-dir internal/sre/testdata
