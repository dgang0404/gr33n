#!/usr/bin/env bash
# Index the demo farm (farm_id=1) into gr33ncore.rag_embedding_chunks so Farm
# Guardian grounded chat can cite operational data — not just the live snapshot.
#
# Requires DATABASE_URL + EMBEDDING_API_KEY in .env (see INSTALL.md).
# Best-effort: exits 0 with a skip message when EMBEDDING_API_KEY is unset so
# bootstrap scripts don't fail on Lite / offline-only dev boxes.
#
# Usage (repo root):
#   ./scripts/rag-ingest-demo.sh
#   ./scripts/rag-ingest-demo.sh --dry-run
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

DRY_RUN=0
while [[ $# -gt 0 ]]; do
  case "$1" in
    --dry-run) DRY_RUN=1 ;;
    -h|--help)
      echo "Usage: scripts/rag-ingest-demo.sh [--dry-run]"
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

if [[ -z "${EMBEDDING_API_KEY:-}" ]]; then
  cat <<'EOF'
==> Skipping rag-ingest for demo farm (EMBEDDING_API_KEY not set in .env)

    Guardian still works with the live snapshot (zones, cycles, alerts).
    Grounded RAG citations need embeddings — set EMBEDDING_API_KEY and re-run:

      ./scripts/rag-ingest-demo.sh

    Or: make dev-stack-fresh-rag   (wipe DB + seed + ingest when key is set)
EOF
  exit 0
fi

command -v go >/dev/null 2>&1 || {
  echo "error: go not found in PATH" >&2
  exit 1
}

FLAGS=(
  -farm-id 1
  -crop-cycles -programs -schedules -automation-rules -executable-actions
  -inventory-definitions -inventory-batches -cost-transactions -alerts -tasks
)
if [[ "$DRY_RUN" -eq 1 ]]; then
  FLAGS+=(-dry-run)
fi

echo "==> rag-ingest demo farm (farm_id=1) — Guardian RAG corpus"
go run ./cmd/rag-ingest "${FLAGS[@]}"
echo "==> rag-ingest demo farm done."
