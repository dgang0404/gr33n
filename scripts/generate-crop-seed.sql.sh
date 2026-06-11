#!/usr/bin/env bash
# Phase 82 WS4a — generate idempotent crop profile seed SQL from data/crop_library.yaml.
# Pattern matches db/migrations/20260610_phase64_crop_knowledge_base.sql (WHERE NOT EXISTS).
#
# Usage (repo root):
#   ./scripts/generate-crop-seed.sql.sh              # validate + print SQL to stdout
#   ./scripts/generate-crop-seed.sql.sh --validate   # CI check only
#   ./scripts/generate-crop-seed.sql.sh -o db/migrations/generated_crop_seed.sql
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
Usage: scripts/generate-crop-seed.sql.sh [--validate] [-o FILE]

  --validate    Validate crop_library.yaml (EC mS/cm, growth_stage_enum) and exit
  -o FILE       Write idempotent seed SQL to FILE (default: stdout)

Full profile expansion (Tier A/B stages) lands in WS4b/WS4c — this script
already emits SQL for any crop rows that include stages in the YAML.
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

go run ./cmd/generate-crop-seed "${ARGS[@]}"
