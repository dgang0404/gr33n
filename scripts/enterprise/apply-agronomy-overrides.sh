#!/usr/bin/env bash
# Phase 83 WS2 — apply farm agronomy override pack (EC/VPD/DLI deltas on builtin profiles).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"

FARM_ID=1
FILE="${AGRONOMY_OVERRIDE_FILE:-$ROOT/data/agronomy-override-pack.example.yaml}"
DRY_RUN=0

usage() {
  cat <<EOF
Usage: $(basename "$0") [--dry-run] [--farm-id N] [--file PATH]

Clone builtin crop profiles for a farm and apply stage target deltas from YAML.

  --dry-run    Validate pack only
  --farm-id N  Target farm (default: 1)
  --file PATH  Override pack YAML (default: data/agronomy-override-pack.example.yaml)

Requires DATABASE_URL in .env (except --dry-run).

Unsupported catalog keys (ramps, …) are rejected.

See docs/plans/phase_83_enterprise_agronomy_seed_pack.plan.md WS2
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --dry-run) DRY_RUN=1; shift ;;
    --farm-id) FARM_ID="$2"; shift 2 ;;
    --file) FILE="$2"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Unknown option: $1" >&2; usage >&2; exit 1 ;;
  esac
done

if [[ ! -f "$FILE" ]]; then
  echo "error: override file not found: $FILE" >&2
  exit 1
fi

ARGS=(-farm-id "$FARM_ID" -file "$FILE")
if [[ "$DRY_RUN" -eq 1 ]]; then
  ARGS+=(-dry-run)
else
  if [[ -f "$ROOT/.env" ]]; then
    set -a
    # shellcheck disable=1091
    source "$ROOT/.env"
    set +a
  fi
  if [[ -z "${DATABASE_URL:-}" ]]; then
    echo "error: DATABASE_URL required" >&2
    exit 1
  fi
fi

go run ./cmd/apply-agronomy-overrides "${ARGS[@]}"
