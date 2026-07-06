#!/usr/bin/env bash
# Verify local dev prerequisites: DATABASE_URL from .env, pgvector, optional API /health.
# Usage (from repo root): ./scripts/check-local-stack.sh
# Exit 0 if DB connects and vector extension exists; non-zero otherwise.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

ENV_FILE="$ROOT/.env"
RED='\033[0;31m'
GRN='\033[0;32m'
YLW='\033[0;33m'
RST='\033[0m'

die() {
  echo -e "${RED}error:${RST} $*" >&2
  exit 1
}

warn() {
  echo -e "${YLW}warning:${RST} $*" >&2
}

ok() {
  echo -e "${GRN}ok${RST}  $*"
}

[[ -f "$ENV_FILE" ]] || die "missing .env — copy .env.example to .env"

# First non-comment DATABASE_URL= line (strip optional quotes).
DATABASE_URL=""
if grep -qE '^[[:space:]]*DATABASE_URL=' "$ENV_FILE"; then
  DATABASE_URL="$(grep -E '^[[:space:]]*DATABASE_URL=' "$ENV_FILE" | head -1 | sed 's/^[[:space:]]*DATABASE_URL=//' | sed 's/^"\(.*\)"$/\1/' | sed "s/^'\(.*\)'$/\1/")"
fi

[[ -n "${DATABASE_URL:-}" ]] || die "DATABASE_URL is empty in .env"

if ! command -v psql >/dev/null 2>&1; then
  die "psql not found — install postgresql-client or use Docker Compose DB only"
fi

echo "==> DATABASE_URL host (from .env): $(echo "$DATABASE_URL" | sed -E 's|//([^:]+):([^@]*)@|//\1:***@|')"
echo "==> Postgres connection"
if ! psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -tAc "SELECT 1" >/dev/null 2>&1; then
  die "cannot connect with DATABASE_URL — fix credentials / host / DB name (see INSTALL.md §2)"
fi
ok "connected"

echo "==> Extension: vector (required for API startup)"
VEC="$(psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -tAc "SELECT count(*) FROM pg_extension WHERE extname='vector'" | tr -d '[:space:]')"
if [[ "${VEC:-0}" != "1" ]]; then
  warn "vector extension missing on this database."
  echo "    Fix options:"
  echo "    • Docker: from repo root run  docker compose up -d db --build"
  echo "      then set .env DATABASE_URL=postgres://gr33n:gr33n@127.0.0.1:5433/gr33n?sslmode=disable"
  echo "    • Native Ubuntu/Debian: ./scripts/install-system-deps-debian.sh (PG16+pgvector), then INSTALL.md §2"
  echo "    • Then load schema and migrations: ./scripts/bootstrap-local.sh --seed"
  exit 1
fi
ok "vector present"

if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
  echo "==> Docker Compose (db service)"
  if docker compose ps db 2>/dev/null | grep -qE 'Up|running'; then
    ok "docker compose db container is up"
  else
    warn "docker is installed but db container not running — try: make compose-db-up"
  fi
else
  warn "docker not found — use native Postgres + pgvector (see INSTALL.md)"
fi

PORT_VAL="8080"
if grep -qE '^[[:space:]]*PORT=' "$ENV_FILE"; then
  PORT_VAL="$(grep -E '^[[:space:]]*PORT=' "$ENV_FILE" | head -1 | sed 's/^[[:space:]]*PORT=//' | tr -d \"\' )"
fi

echo "==> API http://localhost:${PORT_VAL}/health"
if curl -sf "http://127.0.0.1:${PORT_VAL}/health" >/dev/null 2>&1 || curl -sf "http://localhost:${PORT_VAL}/health" >/dev/null 2>&1; then
  ok "API responds (healthy)"
else
  warn "API not reachable on port ${PORT_VAL} — run: make dev-auth-test   (or make dev) from repo root"
fi

OLLAMA_URL="http://127.0.0.1:11434"
if grep -qE '^[[:space:]]*OLLAMA_HOST=' "$ENV_FILE"; then
  OLLAMA_HOST="$(grep -E '^[[:space:]]*OLLAMA_HOST=' "$ENV_FILE" | head -1 | sed 's/^[[:space:]]*OLLAMA_HOST=//' | tr -d \"\' )"
  if [[ -n "$OLLAMA_HOST" ]]; then
    OLLAMA_URL="http://${OLLAMA_HOST}"
  fi
elif grep -qE '^[[:space:]]*LLM_BASE_URL=' "$ENV_FILE"; then
  LLM_BASE="$(grep -E '^[[:space:]]*LLM_BASE_URL=' "$ENV_FILE" | head -1 | sed 's/^[[:space:]]*LLM_BASE_URL=//' | tr -d \"\' )"
  if [[ "$LLM_BASE" =~ ^https?:// ]]; then
    OLLAMA_URL="${LLM_BASE%/v1}"
    OLLAMA_URL="${OLLAMA_URL%/}"
  fi
fi

AI_ENABLED_VAL="true"
if grep -qE '^[[:space:]]*AI_ENABLED=' "$ENV_FILE"; then
  AI_ENABLED_VAL="$(grep -E '^[[:space:]]*AI_ENABLED=' "$ENV_FILE" | head -1 | sed 's/^[[:space:]]*AI_ENABLED=//' | tr -d \"\' | tr '[:upper:]' '[:lower:]')"
fi

echo "==> Ollama ${OLLAMA_URL}/api/tags"
if curl -sf "${OLLAMA_URL}/api/tags" >/dev/null 2>&1; then
  ok "Ollama responds"
else
  if [[ "$AI_ENABLED_VAL" == "true" ]]; then
    warn "Ollama not reachable but AI_ENABLED=true — Farm Guardian needs Ollama."
    echo "    Start Ollama: systemctl start ollama   (Linux) or open the Ollama app (macOS/Windows)"
    echo "    Then log in and open Guardian — awakening preloads models automatically."
    echo "    Laptop timeouts: make guardian-laptop-tune ARGS=\"--apply\"   (repo root)"
  else
    warn "Ollama not reachable — OK in Lite mode (AI_ENABLED=false)"
  fi
fi

echo ""
echo "Next: make dev-auth-test   # or make dev"
exit 0
