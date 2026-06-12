#!/usr/bin/env bash
# Phase 99 — UI ↔ backend domain enum parity (growth stages, lighting presets, dead UI copies).
#
# Usage (repo root):
#   ./scripts/check-ui-domain-parity.sh
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

echo "== Phase 99 — UI domain parity =="

echo ">> Go parity tests (domain enums, lighting presets, OpenAPI GrowthStageEnum)"
go test ./internal/platform/domainenums/... ./internal/handler/lighting/... -run 'Parity|GrowthStages|PresetList' -count=1

echo ">> Vitest UI domain parity"
(
  cd ui
  npm ci --legacy-peer-deps --silent
  npm test -- --run src/__tests__/ui-domain-parity.test.js
)

echo ">> No hardcoded lighting preset keys in production UI"
PRESET_PATTERN='peas_22_2|veg_18_6|flower_12_12|seedling_16_8'
if rg -n "$PRESET_PATTERN" ui/src \
  --glob '*.vue' --glob '*.js' \
  --glob '!**/__tests__/**' \
  --glob '!**/lightingPresets.js' \
  --glob '!**/bootstrapCatalog.fallback.js' \
  2>/dev/null; then
  echo "FAIL: hardcoded lighting preset keys found — use loadLightingPresets()" >&2
  exit 1
fi

echo "OK — UI domain parity checks passed"
