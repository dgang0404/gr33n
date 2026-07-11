#!/usr/bin/env bash
# Phase 157 — print counts to help regenerate docs/current-state.md prose.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

echo "==> docs/current-state.md regeneration hints"
echo "    date: $(date -u +%Y-%m-%d)"
echo ""

if [[ -f openapi.yaml ]]; then
  paths="$(rg -c '^  /' openapi.yaml 2>/dev/null || echo 0)"
  tags="$(rg -c '^  - name:' openapi.yaml 2>/dev/null || echo 0)"
  echo "    OpenAPI paths (approx): $paths"
  echo "    OpenAPI tags: $tags"
fi

if [[ -d db/migrations ]]; then
  latest="$(ls -1 db/migrations/*.sql 2>/dev/null | tail -1 | xargs -r basename)"
  count="$(ls -1 db/migrations/*.sql 2>/dev/null | wc -l | tr -d ' ')"
  echo "    migrations: $count files, latest: $latest"
fi

if command -v go >/dev/null 2>&1; then
  echo ""
  echo "    Guardian eval suites (from -manual smoke header):"
  go run ./cmd/guardian-eval/ -manual -suite smoke 2>/dev/null | head -5 || true
fi

echo ""
echo "    Edit docs/current-state.md prose, then link from README + phase-14 index."
