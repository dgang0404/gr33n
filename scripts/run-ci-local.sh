#!/usr/bin/env bash
# ponytail: mirrors mandatory CI jobs (go + ui) from .github/workflows/ci.yml
set -euo pipefail
cd "$(dirname "$0")/.."

export CI=true
export DATABASE_URL="${DATABASE_URL:-postgres://gr33n:gr33n@127.0.0.1:5433/gr33n?sslmode=disable}"
export CROP_CATALOG_SOURCE=db
export AGRONOMY_FIELD_GUIDES_SOURCE=db

step() { echo ""; echo "=== $* ==="; }

step "Bootstrap schema, migrations, and demo seed"
./scripts/bootstrap-local.sh --seed

step "Validate crop library YAML (Phase 82 WS4a)"
make check-crop-catalog-parity

step "Catalog seed drift gate (Phase 95)"
make check-catalog-seed-drift

step "Verify crop catalog seed (Phase 84 WS-B/K)"
make check-crop-catalog-db

step "UI domain parity guards (Phase 99)"
go test ./internal/platform/domainenums/... ./internal/handler/lighting/... -run 'Parity|GrowthStages|PresetList' -count=1
PRESET_PATTERN='peas_22_2|veg_18_6|flower_12_12|seedling_16_8'
if rg -n "$PRESET_PATTERN" ui/src \
  --glob '*.vue' --glob '*.js' \
  --glob '!**/__tests__/**' \
  --glob '!**/lightingPresets.js' \
  --glob '!**/bootstrapCatalog.fallback.js' \
  2>/dev/null; then
  echo "hardcoded lighting preset keys in production UI"
  exit 1
fi

step "Verify pgvector extension (RAG parity)"
count=$(psql "$DATABASE_URL" -tAc "SELECT count(*) FROM pg_extension WHERE extname='vector'" | tr -d '[:space:]')
test "$count" = "1"
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -c "SELECT to_regclass('gr33ncore.rag_embedding_chunks') AS rag_chunks_table;"

step "OpenAPI route parity"
make audit-openapi

step "Environment variable reference parity"
make audit-env

step "Run Go tests (includes cmd/api smoke tests)"
go test -tags dev ./... -count=1

step "Govulncheck (Phase 156)"
go run golang.org/x/vuln/cmd/govulncheck@latest ./...

step "UI deps, audit, and test"
(
  cd ui
  npm ci --legacy-peer-deps
  npm audit --audit-level=high
  npm test -- --run
)

echo ""
echo "=== ALL MANDATORY CI CHECKS PASSED ==="
