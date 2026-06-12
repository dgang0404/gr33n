#!/usr/bin/env bash
# Phase 84 WS-K — verify platform crop catalog + field guides in Postgres after migrate.
#
# Usage (repo root, DATABASE_URL in .env or env):
#   ./scripts/check-crop-catalog-db.sh
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

if [[ -f "$ROOT/.env" ]]; then
  set -a
  # shellcheck disable=1091
  source "$ROOT/.env"
  set +a
fi

if [[ -z "${DATABASE_URL:-}" ]]; then
  echo "error: DATABASE_URL required (copy .env.example → .env)" >&2
  exit 1
fi

psql_q() {
  psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -tAc "$1" | tr -d '[:space:]'
}

entries=$(psql_q "SELECT count(*) FROM gr33ncrops.crop_catalog_entries")
aliases=$(psql_q "SELECT count(*) FROM gr33ncrops.crop_catalog_aliases")
guides=$(psql_q "SELECT count(*) FROM gr33ncrops.agronomy_field_guides WHERE published")
supported=$(psql_q "SELECT count(*) FROM gr33ncrops.crop_catalog_entries WHERE supported")
builtin=$(psql_q "SELECT count(*) FROM gr33ncrops.crop_profiles WHERE farm_id IS NULL AND is_builtin")
meta_sub=$(psql_q "SELECT count(*) FROM gr33ncrops.crop_profiles WHERE farm_id IS NULL AND is_builtin AND meta ? 'substrate' AND BTRIM(meta->>'substrate') <> ''")

fail=0
check_ge() {
  local name=$1 val=$2 min=$3
  if [[ "$val" -lt "$min" ]]; then
    echo "FAIL $name: got $val want >= $min" >&2
    fail=1
  else
    echo "OK   $name: $val"
  fi
}

check_eq() {
  local name=$1 val=$2 want=$3
  if [[ "$val" != "$want" ]]; then
    echo "FAIL $name: got $val want $want" >&2
    fail=1
  else
    echo "OK   $name: $val"
  fi
}

check_ge "crop_catalog_entries" "$entries" 50
check_ge "crop_catalog_aliases" "$aliases" 30
check_ge "agronomy_field_guides (published)" "$guides" 50
check_ge "supported catalog entries" "$supported" 46
check_ge "builtin crop_profiles" "$builtin" 46
check_ge "builtin profiles with substrate meta (WS-I)" "$meta_sub" 40
check_eq "supported catalog = builtin profiles" "$supported" "$builtin"

if [[ "$fail" -ne 0 ]]; then
  echo "crop catalog DB check failed — run: make migrate" >&2
  exit 1
fi

echo "crop catalog DB OK (entries=$entries aliases=$aliases guides=$guides builtin=$builtin)"
