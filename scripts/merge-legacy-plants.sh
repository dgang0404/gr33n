#!/usr/bin/env bash
# Phase 103 — audit and merge legacy free-text plant rows into catalog crop_key slots.
#
# Usage (repo root):
#   ./scripts/merge-legacy-plants.sh              # audit only
#   ./scripts/merge-legacy-plants.sh --apply      # run gr33ncrops.merge_legacy_plants()
#   ./scripts/merge-legacy-plants.sh --apply --audit  # merge then audit
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

ENV_FILE="$ROOT/.env"
AUDIT_SQL="$ROOT/scripts/sql/legacy_plants_audit.sql"
APPLY=0
POST_AUDIT=0

die() { echo "error: $*" >&2; exit 1; }

while [[ $# -gt 0 ]]; do
  case "$1" in
    --apply) APPLY=1; shift ;;
    --audit) POST_AUDIT=1; shift ;;
    -h|--help)
      cat <<EOF
Usage: ./scripts/merge-legacy-plants.sh [--apply] [--audit]

  (default)   Read-only audit: missing crop_key, duplicate slots, broken cycle links
  --apply     Call gr33ncrops.merge_legacy_plants() (idempotent; safe to re-run)
  --audit     With --apply, print audit report after merge

Requires DATABASE_URL in .env. Prefer \`make migrate\` on new installs — this targets
existing farms that accumulated typo plant rows before Phase 85 catalog binding.
EOF
      exit 0
      ;;
    *) die "unknown arg: $1 (try --help)" ;;
  esac
done

[[ -f "$ENV_FILE" ]] || die "missing .env — copy .env.example"
[[ -f "$AUDIT_SQL" ]] || die "missing $AUDIT_SQL"

DATABASE_URL=""
if grep -qE '^[[:space:]]*DATABASE_URL=' "$ENV_FILE"; then
  DATABASE_URL="$(grep -E '^[[:space:]]*DATABASE_URL=' "$ENV_FILE" | head -1 | sed 's/^[[:space:]]*DATABASE_URL=//' | sed 's/^"\(.*\)"$/\1/' | sed "s/^'\(.*\)'$/\1/")"
fi
[[ -n "${DATABASE_URL:-}" ]] || die "DATABASE_URL empty in .env"
command -v psql >/dev/null 2>&1 || die "psql not found"

run_audit() {
  echo "==> Legacy plants audit"
  psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f "$AUDIT_SQL"
}

if [[ "$APPLY" -eq 0 ]]; then
  run_audit
  echo ""
  echo "ok  audit complete (read-only). Run with --apply to merge."
  exit 0
fi

echo "==> Applying gr33ncrops.merge_legacy_plants()"
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -c "SELECT * FROM gr33ncrops.merge_legacy_plants();"

unresolved="$(psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -tAc "
SELECT count(*) FROM gr33ncrops.plants WHERE deleted_at IS NULL AND crop_key IS NULL;" | tr -d '[:space:]')"

dupes="$(psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -tAc "
SELECT count(*) FROM (
  SELECT 1 FROM gr33ncrops.plants
  WHERE deleted_at IS NULL AND crop_key IS NOT NULL
  GROUP BY farm_id, crop_key HAVING count(*) > 1
) t;" | tr -d '[:space:]')"

broken="$(psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -tAc "
SELECT count(*) FROM gr33nfertigation.crop_cycles c
LEFT JOIN gr33ncrops.plants p ON p.id = c.plant_id AND p.deleted_at IS NULL
WHERE c.plant_id IS NOT NULL AND p.id IS NULL;" | tr -d '[:space:]')"

echo ""
echo "merge summary: unresolved=$unresolved duplicate_groups=$dupes broken_cycle_links=$broken"

if [[ "${dupes:-0}" != "0" || "${broken:-0}" != "0" ]]; then
  echo "error: merge left duplicate crop_key groups or broken cycle links" >&2
  exit 1
fi

if [[ "$POST_AUDIT" -eq 1 || "${unresolved:-0}" != "0" ]]; then
  echo ""
  run_audit
fi

if [[ "${unresolved:-0}" != "0" ]]; then
  echo ""
  echo "warn: $unresolved plant row(s) still lack crop_key — pick catalog crop in UI or add alias, then re-run --apply" >&2
  exit 1
fi

echo ""
echo "ok  legacy plant merge complete"
