#!/usr/bin/env bash
# Phase 83 WS5 — incremental operational RAG ingest for one farm.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

FARM_ID=1
DRY_RUN=0
WATERMARK_FILE=""

usage() {
  cat <<EOF
Usage: $(basename "$0") [--dry-run] [--farm-id N] [--watermark-file PATH]

Incremental ingest: tasks, crop-cycles, programs, schedules, automation-rules,
executable-actions, inventory, costs, alerts.

Uses RAG_INGEST_UPDATED_AFTER from --watermark-file or env when set.

See scripts/enterprise/README.md (Phase 83 WS5 cron example)
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --dry-run) DRY_RUN=1 ;;
    --farm-id) FARM_ID="$2"; shift ;;
    --watermark-file) WATERMARK_FILE="$2"; shift ;;
    -h|--help) usage; exit 0 ;;
    *) echo "unknown option: $1" >&2; exit 1 ;;
  esac
  shift
done

if [[ ! -f "$ROOT/.env" ]]; then
  echo "error: missing .env" >&2
  exit 1
fi
set -a
# shellcheck disable=1091
source "$ROOT/.env"
set +a

UPDATED_AFTER="${RAG_INGEST_UPDATED_AFTER:-}"
if [[ -n "$WATERMARK_FILE" && -f "$WATERMARK_FILE" ]]; then
  UPDATED_AFTER=$(cat "$WATERMARK_FILE")
fi

ARGS=(-farm-id "$FARM_ID" -tasks -crop-cycles -programs -schedules -automation-rules \
  -executable-actions -inventory-definitions -inventory-batches -cost-transactions -alerts)
if [[ -n "$UPDATED_AFTER" ]]; then
  ARGS+=(-updated-after "$UPDATED_AFTER")
fi
if [[ "$DRY_RUN" -eq 1 ]]; then
  ARGS+=(-dry-run)
fi

if [[ "$DRY_RUN" -eq 0 && -z "${EMBEDDING_API_KEY:-}" ]]; then
  echo "==> Skipping operational rag-ingest (EMBEDDING_API_KEY not set)"
  exit 0
fi

echo "==> rag-ingest operational farm_id=${FARM_ID}"
go run ./cmd/rag-ingest "${ARGS[@]}"

if [[ -n "$WATERMARK_FILE" && "$DRY_RUN" -eq 0 ]]; then
  date -u +%Y-%m-%dT%H:%M:%SZ > "$WATERMARK_FILE"
fi
