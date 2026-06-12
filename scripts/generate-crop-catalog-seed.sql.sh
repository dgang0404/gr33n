#!/usr/bin/env bash
# Phase 84 WS-B — generate platform crop catalog + field guide seed SQL.
#
# Usage (repo root):
#   ./scripts/generate-crop-catalog-seed.sql.sh --validate
#   ./scripts/generate-crop-catalog-seed.sql.sh -o db/migrations/20260616_phase84_crop_catalog_seed.sql
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

VALIDATE_ONLY=0
OUTPUT=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --validate|--validate-only)
      VALIDATE_ONLY=1
      ;;
    -o|--output)
      shift
      OUTPUT="${1:?-o requires a path}"
      ;;
    -h|--help)
      cat <<'EOF'
Usage: scripts/generate-crop-catalog-seed.sql.sh [--validate] [-o FILE]

  --validate    Validate crop_library.yaml + field guide manifest
  -o FILE       Write idempotent catalog seed SQL to FILE
EOF
      exit 0
      ;;
    *)
      echo "unknown option: $1" >&2
      exit 1
      ;;
  esac
  shift
done

ARGS=(--repo-root "$ROOT")
if [[ "$VALIDATE_ONLY" -eq 1 ]]; then
  ARGS+=(--validate-only)
elif [[ -n "$OUTPUT" ]]; then
  ARGS+=(-o "$OUTPUT")
fi

go run ./cmd/generate-crop-catalog-seed "${ARGS[@]}"
