#!/usr/bin/env bash
# Phase 48 WS5 — apply Timescale retention policies when TIMESCALE_RETENTION_DAYS is set.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

die() { echo "error: $*" >&2; exit 1; }

DAYS="${TIMESCALE_RETENTION_DAYS:-}"
[[ -n "$DAYS" ]] || die "TIMESCALE_RETENTION_DAYS not set — skipping retention (dev/staging only)"

[[ "$DAYS" =~ ^[0-9]+$ ]] || die "TIMESCALE_RETENTION_DAYS must be a positive integer"
[[ "$DAYS" -gt 0 ]] || die "TIMESCALE_RETENTION_DAYS must be > 0"

ENV_FILE="$ROOT/.env"
[[ -f "$ENV_FILE" ]] || die "missing .env"
DATABASE_URL=""
if grep -qE '^[[:space:]]*DATABASE_URL=' "$ENV_FILE"; then
  DATABASE_URL="$(grep -E '^[[:space:]]*DATABASE_URL=' "$ENV_FILE" | head -1 | sed 's/^[[:space:]]*DATABASE_URL=//' | sed 's/^"\(.*\)"$/\1/' | sed "s/^'\(.*\)'$/\1/")"
fi
[[ -n "${DATABASE_URL:-}" ]] || die "DATABASE_URL empty"
command -v psql >/dev/null 2>&1 || die "psql not found"

echo "==> Applying Timescale retention (${DAYS} days) — dev/staging gate"
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -v retention_days="$DAYS" \
  -f "$ROOT/scripts/sql/apply_dev_timescale_retention.sql"
echo "ok  retention policies applied"
