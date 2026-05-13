#!/usr/bin/env bash
# Read-only Postgres report: extensions, duplicate zones (seed hazards), farm count, RAG chunk count.
# Exit 1 if DB unreachable, vector missing, or duplicate zone names exist (master_seed.sql assumes unique names).
# Usage (repo root): ./scripts/db-sanity-report.sh
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

ENV_FILE="$ROOT/.env"
SQL_FILE="$ROOT/scripts/sql/db_sanity_report.sql"

die() {
  echo "error: $*" >&2
  exit 1
}

[[ -f "$ENV_FILE" ]] || die "missing .env — copy .env.example"
[[ -f "$SQL_FILE" ]] || die "missing $SQL_FILE"

DATABASE_URL=""
if grep -qE '^[[:space:]]*DATABASE_URL=' "$ENV_FILE"; then
  DATABASE_URL="$(grep -E '^[[:space:]]*DATABASE_URL=' "$ENV_FILE" | head -1 | sed 's/^[[:space:]]*DATABASE_URL=//' | sed 's/^"\(.*\)"$/\1/' | sed "s/^'\(.*\)'$/\1/")"
fi
[[ -n "${DATABASE_URL:-}" ]] || die "DATABASE_URL empty in .env"

command -v psql >/dev/null 2>&1 || die "psql not found"

echo "==> DB sanity report ($(echo "$DATABASE_URL" | sed -E 's|//([^:]+):([^@]*)@|//\1:***@|'))"
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f "$SQL_FILE"

dup="$(psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -tAc "
SELECT count(*) FROM (
  SELECT farm_id, name FROM gr33ncore.zones GROUP BY farm_id, name HAVING count(*) > 1
) t;" | tr -d '[:space:]')"

if [[ "${dup:-0}" != "0" ]]; then
  echo ""
  echo "error: duplicate zone names detected ($dup distinct farm_id+name groups). This breaks db/seeds/master_seed.sql" >&2
  echo "  Fix: merge or delete duplicate zones, or docker compose down -v && ./scripts/dev-stack.sh for a fresh DB (destructive)." >&2
  exit 1
fi

echo ""
echo "ok  no duplicate zone names"
