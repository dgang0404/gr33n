#!/usr/bin/env bash
# Phase 95 WS4 — fail CI when crop_library.yaml changed without regenerating committed seed SQL.
#
# Usage (repo root):
#   ./scripts/check-catalog-seed-drift.sh
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

CANONICAL="${ROOT}/db/seed/crop_catalog_from_yaml.sql"
TMP="$(mktemp "${TMPDIR:-/tmp}/crop_catalog_gen.XXXXXX.sql")"
trap 'rm -f "$TMP"' EXIT

if [[ ! -f "$CANONICAL" ]]; then
  echo "error: missing canonical seed $CANONICAL" >&2
  echo "  regenerate: ./scripts/generate-crop-catalog-seed.sql.sh -o db/seed/crop_catalog_from_yaml.sql" >&2
  exit 1
fi

./scripts/generate-crop-catalog-seed.sql.sh -o "$TMP" >/dev/null

if diff -q "$CANONICAL" "$TMP" >/dev/null; then
  ver=$(grep -m1 '^-- crop_library version:' "$CANONICAL" | awk '{print $NF}')
  echo "OK   catalog seed matches YAML (version ${ver:-?})"
  exit 0
fi

echo "FAIL catalog seed drift — data/crop_library.yaml (or field guides) changed without updating committed SQL." >&2
echo "" >&2
echo "Fix:" >&2
echo "  1. Bump version: in data/crop_library.yaml when adding/changing crops" >&2
echo "  2. ./scripts/generate-crop-catalog-seed.sql.sh -o db/seed/crop_catalog_from_yaml.sql" >&2
echo "  3. ./scripts/generate-crop-catalog-seed.sql.sh -o db/migrations/YYYYMMDD_catalog_<slug>.sql" >&2
echo "  4. make migrate && make check-catalog-release" >&2
echo "" >&2
diff -u "$CANONICAL" "$TMP" | head -40 >&2 || true
exit 1
