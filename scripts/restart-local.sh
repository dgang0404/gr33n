#!/usr/bin/env bash
# Post-reboot / quick dev bring-up: start Compose Postgres only, wait (with progress), optional sanity report.
# Does NOT run bootstrap migrations or seed — use scripts/bootstrap-local.sh or scripts/dev-stack.sh when the schema changes.
# Does NOT start API/UI unless --serve (because first `go run` compile can take minutes — not an infinite loop).
#
# Usage (repo root):
#   ./scripts/restart-local.sh
#   ./scripts/restart-local.sh --serve              # then runs make dev-auth-test (API + UI)
#   ./scripts/restart-local.sh --skip-report        # faster if you trust the DB
#   ./scripts/restart-local.sh --quick              # docker compose up without --no-cache rebuild
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

SERVE=0
SKIP_REPORT=0
QUICK=1

usage() {
  cat <<'EOF'
Usage: scripts/restart-local.sh [options]

  (default)      docker compose up -d db, wait for Postgres, db sanity report
  --serve        Run make dev-auth-test after checks (API + UI; Go may compile 1–5+ min cold)
  --skip-report  Skip scripts/db-sanity-report.sh
  --no-quick     docker compose build db --no-cache before up (slow; rare)
  -h, --help     This message

Tip: Slow startup after reboot is usually `go run` compiling the API. Run once:
  go build -tags dev -o ./bin/api ./cmd/api/
then start ./bin/api with the same env as make run-auth-test (advanced).

Requires .env with DATABASE_URL matching Compose (see .env.example).
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --serve) SERVE=1 ;;
    --skip-report) SKIP_REPORT=1 ;;
    --no-quick) QUICK=0 ;;
    -h|--help) usage; exit 0 ;;
    *)
      echo "unknown option: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
  shift
done

die() {
  echo "error: $*" >&2
  exit 1
}

need() {
  command -v "$1" >/dev/null 2>&1 || die "missing '$1' in PATH"
}

need docker
docker compose version >/dev/null 2>&1 || die "need Compose v2 (docker compose)"

[[ -f "$ROOT/.env" ]] || die "missing .env — copy .env.example"

compose() {
  if docker info >/dev/null 2>&1; then
    docker compose "$@"
    return
  fi
  if sg docker -c 'docker info' >/dev/null 2>&1; then
    # shellcheck disable=SC2145
    sg docker -c "cd $(printf '%q' "$ROOT") && docker compose $(printf '%q ' "$@")"
    return
  fi
  die "cannot talk to Docker — try: newgrp docker or sg docker -c 'cd $(printf '%q' "$ROOT") && ./scripts/restart-local.sh'"
}

echo "==> restart-local: Compose db + wait + checks (repo: $ROOT)"

if [[ "$QUICK" -eq 1 ]]; then
  compose up -d db
else
  compose build db --no-cache
  compose up -d db
fi

echo -n "==> Waiting for Postgres (gr33n)"
ready=0
for i in $(seq 1 90); do
  if compose exec -T db pg_isready -U gr33n -d gr33n >/dev/null 2>&1; then
    ready=1
    echo ""
    echo "    Ready after ~$((i * 2))s."
    break
  fi
  echo -n "."
  sleep 2
done

[[ "$ready" -eq 1 ]] || die "Postgres not ready — docker compose logs db"

set -a
# shellcheck disable=1091
source "$ROOT/.env"
set +a

if [[ "$SKIP_REPORT" -eq 0 ]]; then
  echo ""
  ./scripts/db-sanity-report.sh
else
  echo "==> Skipping db sanity report (--skip-report)"
  ./scripts/check-local-stack.sh
fi

echo ""
echo "==> Done."
if [[ "$SERVE" -eq 1 ]]; then
  echo "Starting API + UI (make dev-auth-test) — first compile may take several minutes."
  exec make dev-auth-test
fi

echo "Next:"
echo "  ./scripts/check-local-stack.sh   # optional API /health ping"
echo "  make dev-auth-test               # API + UI (AUTH_MODE from Makefile)"
echo "  make dev                         # dev auth bypass"
