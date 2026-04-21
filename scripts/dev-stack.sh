#!/usr/bin/env bash
# One entry point for local Docker Postgres + bootstrap + verify (+ optional UI).
# Handles "permission denied" on docker.sock by retrying with: sg docker -c '...'
#
# Usage (always from repo root, or run via make):
#   ./scripts/dev-stack.sh
#   ./scripts/dev-stack.sh --serve              # also run make dev-auth-test
#   ./scripts/dev-stack.sh --reset-volumes      # wipe Compose volumes (DESTROYS DB DATA)
#   ./scripts/dev-stack.sh --quick              # docker compose build uses cache (faster rebuilds)
#   ./scripts/dev-stack.sh --skip-seed          # only bring DB up + migrations from bootstrap without seed
#
# Requires: .env with DATABASE_URL matching Compose (see .env.example).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

SERVE=0
RESET_VOL=0
QUICK=0
SKIP_SEED=0

usage() {
  cat <<'EOF'
Usage: scripts/dev-stack.sh [options]

  (default)     Build/start Compose db, wait for Postgres, bootstrap --seed, check-stack
  --serve       Then run make dev-auth-test (API + UI)
  --reset-volumes  docker compose down -v first (wipes DB volumes for this project)
  --quick       Use cached Docker layers (omit db --no-cache rebuild)
  --skip-seed   Bootstrap schema/migrations without master_seed.sql
  -h, --help    This message

Requires .env with DATABASE_URL matching Docker Compose (see .env.example).
EOF
  exit 0
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --serve) SERVE=1 ;;
    --reset-volumes) RESET_VOL=1 ;;
    --quick) QUICK=1 ;;
    --skip-seed) SKIP_SEED=1 ;;
    -h|--help) usage ;;
    *)
      echo "unknown option: $1" >&2
      echo "Try: ./scripts/dev-stack.sh --help" >&2
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
docker compose version >/dev/null 2>&1 || die "install Compose v2 (Ubuntu: sudo apt-get install -y docker-compose-v2)"

# Wrap docker compose when the current shell is not in group docker yet.
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
  die "cannot talk to Docker (permission denied on /var/run/docker.sock). Run: sudo usermod -aG docker \"\$USER\" then log out/in, or use: sg docker -c 'cd $(printf '%q' "$ROOT") && ./scripts/dev-stack.sh'"
}

echo "==> gr33n dev-stack (repo: $ROOT)"

if [[ ! -f "$ROOT/.env" ]]; then
  die "missing .env — copy .env.example to .env and set DATABASE_URL (Compose: postgres://gr33n:gr33n@127.0.0.1:5432/gr33n?sslmode=disable)"
fi

if [[ "$RESET_VOL" -eq 1 ]]; then
  echo "==> Stopping stack and removing Compose volumes (DATA LOSS for this compose project)"
  compose down -v || true
fi

echo "==> Building / starting db service (first pgvector build can take several minutes)"
if [[ "$QUICK" -eq 1 ]]; then
  compose up -d db --build
else
  compose build db --no-cache
  compose up -d db
fi

echo "==> Waiting for Postgres (user gr33n)"
ready=0
for _ in $(seq 1 90); do
  if compose exec -T db pg_isready -U gr33n -d gr33n >/dev/null 2>&1; then
    ready=1
    echo "    Postgres is accepting connections."
    break
  fi
  sleep 2
done
[[ "$ready" -eq 1 ]] || die "db did not become ready — try: compose logs db (via: docker compose logs db)"

echo "==> Bootstrap (sources .env)"
if [[ "$SKIP_SEED" -eq 1 ]]; then
  ./scripts/bootstrap-local.sh
else
  ./scripts/bootstrap-local.sh --seed
fi

echo "==> check-stack"
./scripts/check-local-stack.sh || {
  echo "" >&2
  echo "check-stack reported issues — API may still be down until you run the dev server." >&2
}

echo ""
echo "==> Done."
if [[ "$SERVE" -eq 1 ]]; then
  echo "Starting API + UI (AUTH_MODE from .env; use dev-auth-test-style secrets for auth_test)..."
  exec make dev-auth-test
fi

echo "Next:"
echo "  make dev-auth-test          # API + UI"
echo "  curl -s http://localhost:8080/health"
