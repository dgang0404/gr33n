#!/usr/bin/env bash
# Phase 48 — reset one farm's demo config without wiping Docker volumes.
# Usage:
#   ./scripts/dev-reset-farm.sh --farm-id 1 --profile small_indoor
#   ./scripts/dev-reset-farm.sh --farm-id 1 --profile demo_showcase --include-readings
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

ENV_FILE="$ROOT/.env"
FARM_ID=1
PROFILE="${DEV_SEED_PROFILE:-small_indoor}"
INCLUDE_READINGS=0

die() { echo "error: $*" >&2; exit 1; }

while [[ $# -gt 0 ]]; do
  case "$1" in
    --farm-id) FARM_ID="${2:-}"; shift 2 ;;
    --profile) PROFILE="${2:-}"; shift 2 ;;
    --include-readings) INCLUDE_READINGS=1; shift ;;
    -h|--help)
      cat <<EOF
Usage: ./scripts/dev-reset-farm.sh [--farm-id N] [--profile small_indoor|demo_showcase] [--include-readings]

Re-applies idempotent master_seed.sql and applies profile trim/restore.
Does not wipe auth users or Docker volumes — use make dev-stack-fresh for that.

Profiles: see docs/dev-farm-profiles.md
EOF
      exit 0
      ;;
    *) die "unknown arg: $1 (try --help)" ;;
  esac
done

[[ -f "$ENV_FILE" ]] || die "missing .env — copy .env.example"
case "$PROFILE" in
  small_indoor|demo_showcase) ;;
  *) die "profile must be small_indoor or demo_showcase (got: $PROFILE)" ;;
esac

DATABASE_URL=""
if grep -qE '^[[:space:]]*DATABASE_URL=' "$ENV_FILE"; then
  DATABASE_URL="$(grep -E '^[[:space:]]*DATABASE_URL=' "$ENV_FILE" | head -1 | sed 's/^[[:space:]]*DATABASE_URL=//' | sed 's/^"\(.*\)"$/\1/' | sed "s/^'\(.*\)'$/\1/")"
fi
[[ -n "${DATABASE_URL:-}" ]] || die "DATABASE_URL empty in .env"
command -v psql >/dev/null 2>&1 || die "psql not found"

echo "==> dev-reset farm_id=$FARM_ID profile=$PROFILE include_readings=$INCLUDE_READINGS"

if [[ "$INCLUDE_READINGS" -eq 1 ]]; then
  psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -v farm_id="$FARM_ID" \
    -f "$ROOT/scripts/sql/dev_reset_farm_readings.sql"
fi

echo "==> Purging smoke-test pollution (farm 1 + extra test farms)"
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f "$ROOT/scripts/sql/dev_purge_smoke_pollution.sql"

echo "==> Re-applying master_seed.sql (idempotent)"
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f "$ROOT/db/seeds/master_seed.sql"

case "$PROFILE" in
  small_indoor)
    psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f "$ROOT/db/seeds/small_indoor_farm1.sql"
    ;;
  demo_showcase)
    psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f "$ROOT/db/seeds/demo_showcase_restore_farm1.sql"
    ;;
esac

echo "==> Sanity check"
"$ROOT/scripts/db-sanity-report.sh" || true

echo "ok  dev-reset complete (farm $FARM_ID, profile $PROFILE)"
