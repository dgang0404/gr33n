#!/usr/bin/env bash
# Phase 212 — run Install B host API on :8081 (AUTH_MODE=auth_test, AI_ENABLED=false).
# Run from Install B root after phase212-install-b-setup.sh.
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
set -a
# shellcheck disable=1091
source "$ROOT/.env"
set +a
export PORT="${PORT:-8081}"
export CORS_ORIGIN="${CORS_ORIGIN:-http://localhost:5174}"
export AI_ENABLED="${AI_ENABLED:-false}"
export AUTH_MODE="${AUTH_MODE:-auth_test}"
echo "Install B API → :$PORT  AI_ENABLED=$AI_ENABLED  DATABASE_URL=$DATABASE_URL"
exec go run -tags dev ./cmd/api/
