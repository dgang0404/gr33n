#!/usr/bin/env bash
# Phase 83 WS3 — Guardian bootstrap: verify platform catalog, RAG ingest, readiness report.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"

FARM_ID=1
DRY_RUN=0
SKIP_FIELD=0
SKIP_PLATFORM=0
SKIP_OPS=0
SMOKE=0
IMPORT_PACK=0
MIN_FIELD_CHUNKS=12
API_URL="${GR33N_API_URL:-http://localhost:8080}"

usage() {
  cat <<EOF
Usage: $(basename "$0") [--dry-run] [--farm-id N] [options]

One-command Guardian RAG bootstrap after migrate + Phase 84 catalog seed.

  --dry-run              Print steps only
  --farm-id N            Target farm (default: 1)
  --skip-field-guides    Skip field_guide ingest (DB source default)
  --skip-platform-docs   Skip platform_doc ingest
  --skip-operational     Skip operational domain ingest
  --no-smoke             Skip API health checks (default without --smoke)
  --smoke                Probe /v1/chat/health after ingest
  --import-pack          Also run import-agronomy-seed-pack.sh for this farm
  --min-field-chunks N   Minimum field_guide chunks (default: 12)

Requires .env: DATABASE_URL. Ingest requires EMBEDDING_API_KEY (skipped with message if unset).

See docs/crop-catalog-db-cutover-runbook.md
EOF
}

load_env() {
  if [[ ! -f "$ROOT/.env" ]]; then
    echo "error: missing .env — copy .env.example" >&2
    exit 1
  fi
  set -a
  # shellcheck disable=1091
  source "$ROOT/.env"
  set +a
  if [[ -n "${PORT:-}" ]]; then
    API_URL="http://localhost:${PORT}"
  fi
  export AGRONOMY_FIELD_GUIDES_SOURCE="${AGRONOMY_FIELD_GUIDES_SOURCE:-db}"
  export CROP_CATALOG_SOURCE="${CROP_CATALOG_SOURCE:-db}"
}

psql_q() {
  psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -tAc "$1" | tr -d '[:space:]'
}

verify_platform_catalog() {
  echo "==> verify platform crop catalog (Postgres)"
  ./scripts/check-crop-catalog-db.sh
}

report_chunks() {
  local farm="$1"
  local fg pd total ops
  fg=$(psql_q "SELECT count(*) FROM gr33ncore.rag_embedding_chunks WHERE farm_id=${farm} AND source_type='field_guide'")
  pd=$(psql_q "SELECT count(*) FROM gr33ncore.rag_embedding_chunks WHERE farm_id=${farm} AND source_type='platform_doc'")
  total=$(psql_q "SELECT count(*) FROM gr33ncore.rag_embedding_chunks WHERE farm_id=${farm}")
  ops=$((total - fg - pd))
  [[ "$ops" -lt 0 ]] && ops=0

  local builtins
  builtins=$(psql_q "SELECT count(*) FROM gr33ncrops.crop_profiles WHERE farm_id IS NULL AND is_builtin")

  echo ""
  echo "Guardian bootstrap — farm_id=${farm}"
  echo "  crop_profiles (builtin): ${builtins} OK"
  echo "  rag chunks field_guide:  ${fg} $([[ "$fg" -ge "$MIN_FIELD_CHUNKS" ]] && echo OK || echo "WARN (min ${MIN_FIELD_CHUNKS})")"
  echo "  rag chunks platform_doc: ${pd} $([[ "$pd" -gt 0 ]] && echo OK || echo WARN)"
  echo "  rag chunks operational:  ${ops} $([[ "$ops" -gt 0 ]] && echo OK || echo "WARN (greenfield ok)")"
  if [[ -n "${EMBEDDING_API_KEY:-}" ]]; then
    echo "  embedding: configured OK"
  else
    echo "  embedding: SKIP (EMBEDDING_API_KEY unset — ingest skipped)"
  fi
  if [[ -n "${LLM_BASE_URL:-}" ]] || [[ -n "${LLM_API_KEY:-}" ]]; then
    echo "  llm: ${LLM_MODEL:-default} configured"
  else
    echo "  llm: not configured (Guardian structured tools still work)"
  fi

  local fail=0
  [[ "$builtins" -lt 46 ]] && fail=1
  if [[ -n "${EMBEDDING_API_KEY:-}" ]] && [[ "$fg" -lt "$MIN_FIELD_CHUNKS" ]]; then
    fail=1
  fi
  return "$fail"
}

run_smoke() {
  echo "==> smoke: chat health"
  if ! curl -sf "${API_URL}/health" >/dev/null 2>&1; then
    echo "  WARN API not reachable at ${API_URL} — start make dev-auth-test for --smoke"
    return 0
  fi
  local out
  out=$(curl -sf "${API_URL}/v1/chat/health?farm_id=${FARM_ID}" 2>/dev/null || true)
  if [[ -n "$out" ]]; then
    echo "  chat health: ${out:0:120}"
  else
    echo "  WARN /v1/chat/health failed (JWT may be required on your build)"
  fi
}

run_dry() {
  cat <<EOF
Guardian bootstrap (dry-run) — farm_id=${FARM_ID}
  1. verify platform catalog (check-crop-catalog-db.sh)
EOF
  if [[ "$IMPORT_PACK" -eq 1 ]]; then
    echo "  2. import-agronomy-seed-pack.sh --farm-ids ${FARM_ID}"
  fi
  [[ "$SKIP_FIELD" -eq 0 ]] && echo "  • rag-ingest -field-guides -farm-id ${FARM_ID} (AGRONOMY_FIELD_GUIDES_SOURCE=db)"
  [[ "$SKIP_PLATFORM" -eq 0 ]] && echo "  • rag-ingest -platform-docs -farm-id ${FARM_ID}"
  [[ "$SKIP_OPS" -eq 0 ]] && echo "  • rag-ingest operational domains -farm-id ${FARM_ID}"
  echo "  • report chunk counts (min field_guide=${MIN_FIELD_CHUNKS})"
  [[ "$SMOKE" -eq 1 ]] && echo "  • GET /v1/chat/health?farm_id=${FARM_ID}"
  echo ""
  echo "Dry-run only — no commands executed."
}

run_ingest() {
  if [[ -z "${EMBEDDING_API_KEY:-}" ]]; then
    cat <<'EOF'
==> Skipping RAG ingest (EMBEDDING_API_KEY not set)

    Platform catalog verification still runs. Set EMBEDDING_API_KEY and re-run for embeddings.

EOF
    return 0
  fi

  local flags=(-farm-id "$FARM_ID")
  if [[ "$SKIP_FIELD" -eq 0 ]]; then
    echo "==> ingest field guides (DB)"
    go run ./cmd/rag-ingest "${flags[@]}" -field-guides -repo-root "$ROOT"
  fi
  if [[ "$SKIP_PLATFORM" -eq 0 ]]; then
    echo "==> ingest platform docs"
    go run ./cmd/rag-ingest "${flags[@]}" -platform-docs -repo-root "$ROOT"
  fi
  if [[ "$SKIP_OPS" -eq 0 ]]; then
    echo "==> ingest operational domains"
    go run ./cmd/rag-ingest "${flags[@]}" \
      -crop-cycles -programs -schedules -automation-rules -executable-actions \
      -inventory-definitions -inventory-batches -cost-transactions -alerts -tasks
  fi
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --dry-run) DRY_RUN=1; shift ;;
    --farm-id) FARM_ID="$2"; shift 2 ;;
    --skip-field-guides) SKIP_FIELD=1; shift ;;
    --skip-platform-docs) SKIP_PLATFORM=1; shift ;;
    --skip-operational) SKIP_OPS=1; shift ;;
    --smoke) SMOKE=1; shift ;;
    --no-smoke) SMOKE=0; shift ;;
    --import-pack) IMPORT_PACK=1; shift ;;
    --min-field-chunks) MIN_FIELD_CHUNKS="$2"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Unknown option: $1" >&2; usage >&2; exit 1 ;;
  esac
done

if [[ "$FARM_ID" -le 0 ]]; then
  echo "error: --farm-id must be > 0" >&2
  exit 1
fi

load_env

if [[ "$DRY_RUN" -eq 1 ]]; then
  run_dry
  exit 0
fi

verify_platform_catalog

if [[ "$IMPORT_PACK" -eq 1 ]]; then
  echo "==> import agronomy seed pack audit"
  GR33N_FARM_IDS="$FARM_ID" "$ROOT/scripts/enterprise/import-agronomy-seed-pack.sh"
fi

run_ingest

report_chunks "$FARM_ID" || { echo "bootstrap FAILED readiness thresholds" >&2; exit 1; }

if [[ "$SMOKE" -eq 1 ]]; then
  run_smoke
fi

echo ""
echo "==> guardian bootstrap done (farm_id=${FARM_ID})"
