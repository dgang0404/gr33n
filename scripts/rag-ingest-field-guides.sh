#!/usr/bin/env bash
# Index field/trades guides into gr33ncore.rag_embedding_chunks (source_type=field_guide).
# Default source: gr33ncrops.agronomy_field_guides (AGRONOMY_FIELD_GUIDES_SOURCE=db).
# Legacy file manifest: AGRONOMY_FIELD_GUIDES_SOURCE=file (deprecated — see docs/crop-catalog-db-cutover-runbook.md).
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

export AGRONOMY_FIELD_GUIDES_SOURCE="${AGRONOMY_FIELD_GUIDES_SOURCE:-db}"

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
