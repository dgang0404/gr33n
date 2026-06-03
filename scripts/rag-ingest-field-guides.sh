#!/usr/bin/env bash
# Index authored field/trades guides (docs/rag/field-guide-manifest.yaml) into
# gr33ncore.rag_embedding_chunks with source_type=field_guide (Phase 37 WS2).
#
# Usage (repo root):
#   ./scripts/rag-ingest-field-guides.sh
#   ./scripts/rag-ingest-field-guides.sh --dry-run
#   ./scripts/rag-ingest-field-guides.sh --farm-id 1
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

DRY_RUN=0
FARM_ID=1
while [[ $# -gt 0 ]]; do
  case "$1" in
    --dry-run) DRY_RUN=1 ;;
    --farm-id)
      shift
      FARM_ID="${1:?--farm-id requires a value}"
      ;;
    -h|--help)
      echo "Usage: scripts/rag-ingest-field-guides.sh [--dry-run] [--farm-id N]"
      exit 0
      ;;
    *)
      echo "unknown option: $1" >&2
      exit 1
      ;;
  esac
  shift
done

if [[ ! -f "$ROOT/.env" ]]; then
  echo "error: missing .env — copy .env.example" >&2
  exit 1
fi

set -a
# shellcheck disable=1091
source "$ROOT/.env"
set +a

if [[ "$DRY_RUN" -eq 1 ]]; then
  echo "==> rag-ingest field guides (dry-run, farm_id=$FARM_ID)"
  go run ./cmd/rag-ingest \
    -farm-id "$FARM_ID" \
    -field-guides \
    -repo-root "$ROOT" \
    -dry-run
  exit 0
fi

if [[ -z "${EMBEDDING_API_KEY:-}" ]]; then
  cat <<'EOF'
==> Skipping field guide rag-ingest (EMBEDDING_API_KEY not set in .env)

    Set EMBEDDING_API_KEY and re-run:

      ./scripts/rag-ingest-field-guides.sh

EOF
  exit 0
fi

echo "==> rag-ingest field guides (farm_id=$FARM_ID)"
go run ./cmd/rag-ingest \
  -farm-id "$FARM_ID" \
  -field-guides \
  -repo-root "$ROOT"
echo "==> rag-ingest field guides done."
