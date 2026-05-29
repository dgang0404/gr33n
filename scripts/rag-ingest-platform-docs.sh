#!/usr/bin/env bash
# Index curated operator markdown (docs/rag/platform-doc-manifest.yaml) into
# gr33ncore.rag_embedding_chunks with source_type=platform_doc so Farm Guardian
# can cite how-to / troubleshooting guides — not just farm operational rows.
#
# Requires DATABASE_URL + EMBEDDING_API_KEY in .env (see INSTALL.md).
# Best-effort: exits 0 with a skip message when EMBEDDING_API_KEY is unset.
#
# Usage (repo root):
#   ./scripts/rag-ingest-platform-docs.sh
#   ./scripts/rag-ingest-platform-docs.sh --dry-run
#   ./scripts/rag-ingest-platform-docs.sh --farm-id 1
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
      echo "Usage: scripts/rag-ingest-platform-docs.sh [--dry-run] [--farm-id N]"
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
  echo "==> rag-ingest platform docs (dry-run, farm_id=$FARM_ID)"
  go run ./cmd/rag-ingest \
    -farm-id "$FARM_ID" \
    -platform-docs \
    -repo-root "$ROOT" \
    -dry-run
  exit 0
fi

if [[ -z "${EMBEDDING_API_KEY:-}" ]]; then
  cat <<'EOF'
==> Skipping platform doc rag-ingest (EMBEDDING_API_KEY not set in .env)

    Guardian still works with live snapshot + read tools.
    Platform how-to citations need embeddings — set EMBEDDING_API_KEY and re-run:

      ./scripts/rag-ingest-platform-docs.sh

    Or after demo farm ingest: make dev-stack-fresh-rag
EOF
  exit 0
fi

command -v go >/dev/null 2>&1 || {
  echo "error: go not found in PATH" >&2
  exit 1
}

echo "==> rag-ingest platform docs (farm_id=$FARM_ID)"
go run ./cmd/rag-ingest \
  -farm-id "$FARM_ID" \
  -platform-docs \
  -repo-root "$ROOT"
echo "==> rag-ingest platform docs done."
