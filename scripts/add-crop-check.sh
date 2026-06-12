#!/usr/bin/env bash
# Phase 95 WS2 — pre-migrate integrator validation (no DATABASE_URL required).
#
# Usage (repo root):
#   ./scripts/add-crop-check.sh
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

echo "== Phase 95 — add-crop pre-flight =="
echo ""
echo "Integrator checklist (before migrate on each site):"
echo "  1. Edit data/crop_library.yaml (+ docs/field-guides/crop-*.md if supported crop)"
echo "  2. Bump version: in YAML (monotonic catalog_version for all rows)"
echo "  3. ./scripts/generate-crop-catalog-seed.sql.sh -o db/seed/crop_catalog_from_yaml.sql"
echo "  4. ./scripts/generate-crop-catalog-seed.sql.sh -o db/migrations/YYYYMMDD_catalog_<slug>.sql"
echo "  5. make add-crop-check          ← you are here"
echo "  6. make migrate"
echo "  7. make check-catalog-release"
echo "  8. make rag-ingest-field-guides (if guide body changed)"
echo "  9. Restart API (CROP_CATALOG_SOURCE=db)"
echo " 10. Smoke: picker + GET /commons/crop-catalog/{crop_key}"
echo ""
echo "See docs/catalog-integrator-playbook.md"
echo ""

make check-crop-library
make check-crop-catalog
make check-catalog-seed-drift

echo ""
echo "OK — add-crop pre-flight passed (run make check-catalog-release after migrate)"
