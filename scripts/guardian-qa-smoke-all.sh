#!/usr/bin/env bash
# Run all non-redundant Guardian QA smoke suites sequentially (laptop / self-hosted).
# Batches Q&A so each eval gets its own suite wall-clock budget (see GUARDIAN_EVAL_SUITE_TIMEOUT_HOURS).
#
# Usage (repo root):
#   make guardian-qa-smoke-all
#   GUARDIAN_QA_UI=1 make guardian-qa-smoke-all          # + multi-turn UI quick (~50 min)
#   GUARDIAN_QA_UI_FULL=1 make guardian-qa-smoke-all     # + full change-requests-ui (~2–3 hr)
#   GUARDIAN_QA_FAIL_FAST=1 make guardian-qa-smoke-all   # stop on first suite failure
#   GUARDIAN_QA_KILL_STALE=1 make guardian-qa-smoke-all  # best-effort kill leftover guardian-eval PIDs
set -uo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

MODEL="${MODEL:-phi3:mini}"
FARM_ID="${FARM_ID:-1}"
API_URL="${GR33N_API_URL:-http://127.0.0.1:8080}"
LOG="${GUARDIAN_QA_ALL_LOG:-/tmp/guardian-qa-smoke-all.log}"
FAIL_FAST="${GUARDIAN_QA_FAIL_FAST:-0}"
ARCHIVE_DIR="${GUARDIAN_QA_RUNS_DIR:-data/guardian_qa_runs}"
# Durable copy survives /tmp wipe — grade mid-run kills from here.
DURABLE_LOG="${ARCHIVE_DIR}/smoke-all-latest.log"

if [[ -f .env ]]; then set -a && . ./.env && set +a; fi
# shellcheck disable=SC1091
source scripts/source-local-env.sh --refresh-eval-token

ROLLUP_LABELS=()
ROLLUP_ARCHIVES=()

kill_stale_eval() {
  if [[ "${GUARDIAN_QA_KILL_STALE:-}" != "1" ]]; then
    return 0
  fi
  echo "==> GUARDIAN_QA_KILL_STALE=1 — best-effort pkill of leftover guardian-eval (never ollama serve)"
  pkill -f '[g]o run ./cmd/guardian-eval/' 2>/dev/null || true
  pkill -f '[g]uardian-eval/' 2>/dev/null || true
}

preflight_guardian() {
  echo ""
  echo "================================================================"
  echo "==> Preflight — API health + Guardian farm_counsel warmup"
  echo "    ${API_URL}/health then make guardian-qa-preflight"
  echo "================================================================"
  if ! curl -sf "${API_URL}/health" >/dev/null; then
    echo "==> Preflight: API not reachable at ${API_URL}/health" >&2
    return 1
  fi
  if make guardian-qa-preflight MODEL="${MODEL}" FARM_ID="${FARM_ID}"; then
    echo "==> Preflight: Guardian ready"
    return 0
  fi
  echo "==> Preflight: FAILED (Guardian not ready — start make dev-auth-test + Ollama)" >&2
  return 1
}

newest_archive_since() {
  local since="$1"
  local f
  while IFS= read -r f; do
    if [[ -n "${since}" && "${f}" -ot "${since}" ]]; then
      continue
    fi
    echo "${f}"
    return 0
  done < <(find "${ARCHIVE_DIR}" -maxdepth 1 -name '*.json' -type f -printf '%T@ %p\n' 2>/dev/null | sort -rn | cut -d' ' -f2-)
  return 1
}

capture_archive() {
  local label="$1"
  local marker="$2"
  local arch
  arch="$(newest_archive_since "${marker}")" || return 0
  ROLLUP_LABELS+=("${label}")
  ROLLUP_ARCHIVES+=("${arch}")
}

run_suite() {
  local target="$1"
  local label="$2"
  local marker
  marker="$(mktemp)"
  touch "${marker}"
  echo ""
  echo "================================================================"
  echo "==> ${label}"
  echo "    make ${target} MODEL=${MODEL} FARM_ID=${FARM_ID}"
  echo "================================================================"
  if make "${target}" MODEL="${MODEL}" FARM_ID="${FARM_ID}"; then
    echo "==> ${label}: OK"
    capture_archive "${label}" "${marker}"
    rm -f "${marker}"
    return 0
  fi
  echo "==> ${label}: FAILED" >&2
  capture_archive "${label} (FAILED)" "${marker}"
  rm -f "${marker}"
  return 1
}

print_archive_rollup() {
  echo ""
  echo "================================================================"
  echo "==> Archive rollup (${ARCHIVE_DIR})"
  echo "================================================================"
  if [[ ${#ROLLUP_ARCHIVES[@]} -eq 0 ]]; then
    echo "  (no archives captured this run)"
    return 0
  fi
  local total_pass=0 total_all=0
  local i label arch
  for i in "${!ROLLUP_ARCHIVES[@]}"; do
    label="${ROLLUP_LABELS[$i]}"
    arch="${ROLLUP_ARCHIVES[$i]}"
    if [[ ! -f "${arch}" ]]; then
      echo "  ${label}: (missing ${arch})"
      continue
    fi
    read -r passed total suite model <<<"$(python3 - "${arch}" <<'PY'
import json, sys
path = sys.argv[1]
with open(path) as f:
    arch = json.load(f)
scores = arch.get("scores") or []
passed = sum(1 for s in scores if s.get("passed"))
print(passed, len(scores), arch.get("suite", "?"), arch.get("model", "?"))
PY
)"
    total_pass=$((total_pass + passed))
    total_all=$((total_all + total))
    echo "  ${label}: ${passed}/${total} passed (suite=${suite} model=${model})"
    echo "    $(basename "${arch}")"
  done
  if [[ ${total_all} -gt 0 ]]; then
    echo ""
    echo "  Total this run: ${total_pass}/${total_all} prompts passed heuristic"
  fi
}

SUITES=(
  "guardian-qa-smoke|Core smoke (5 prompts, ~90 min CPU)"
  "guardian-qa-smoke-nf-batch1|NF batch 1 (5 prompts, ~150 min CPU)"
  "guardian-qa-smoke-nf-batch2|NF batch 2 (7 prompts incl. history-compare, ~210 min CPU)"
  "guardian-qa-phase127|Phase 127 grounding (4 prompts, ~90 min CPU)"
  "guardian-qa-change-requests-pending|Change requests + Pending tab (4 write-intents, ~100 min CPU)"
)

if [[ "${GUARDIAN_QA_UI:-}" == "1" ]]; then
  SUITES+=("guardian-qa-change-requests-ui-quick|Multi-turn UI quick (ack + schedule, ~50 min CPU)")
fi
if [[ "${GUARDIAN_QA_UI_FULL:-}" == "1" ]]; then
  SUITES+=("guardian-qa-change-requests-ui|Multi-turn UI full (5 scenarios, ~2–3 hr CPU)")
fi

echo "Guardian QA smoke-all — MODEL=${MODEL} FARM_ID=${FARM_ID}"
echo "Log: ${LOG}"
echo "Durable log: ${DURABLE_LOG}"
echo "Archives: ${ARCHIVE_DIR}/"
echo "Suites: ${#SUITES[@]} (+ preflight before each; manual checklist after)"
echo "Skip manual: GUARDIAN_QA_SKIP_MANUAL=1"
echo "Suite timeout: GUARDIAN_EVAL_SUITE_TIMEOUT_HOURS=${GUARDIAN_EVAL_SUITE_TIMEOUT_HOURS:-12} (default 12h per batch)"

failures=0
mkdir -p "${ARCHIVE_DIR}"
: >"${LOG}"
: >"${DURABLE_LOG}"

tee_log() {
  tee -a "${LOG}" "${DURABLE_LOG}"
}

print_manual_checklist() {
  if [[ "${GUARDIAN_QA_SKIP_MANUAL:-}" == "1" ]]; then
    echo "==> Manual checklist skipped (GUARDIAN_QA_SKIP_MANUAL=1)"
    return 0
  fi
  echo ""
  echo "================================================================"
  echo "==> Manual UI checklist — same prompts for spot-check in browser"
  echo "    make guardian-qa-manual SUITE=smoke-all"
  echo "================================================================"
  if make guardian-qa-manual SUITE=smoke-all 2>&1 | tee_log; then
    echo "==> Manual checklist: printed"
    return 0
  fi
  echo "==> Manual checklist: FAILED" >&2
  return 1
}

kill_stale_eval 2>&1 | tee_log

for entry in "${SUITES[@]}"; do
  target="${entry%%|*}"
  label="${entry#*|}"
  if ! preflight_guardian 2>&1 | tee_log; then
    failures=$((failures + 1))
    if [[ "${FAIL_FAST}" == "1" ]]; then
      echo "FAIL_FAST=1 — stopping after preflight failure" >&2
      break
    fi
  fi
  if run_suite "${target}" "${label}" 2>&1 | tee_log; then
    :
  else
    failures=$((failures + 1))
    if [[ "${FAIL_FAST}" == "1" ]]; then
      echo "FAIL_FAST=1 — stopping after first suite failure" >&2
      break
    fi
  fi
done

print_archive_rollup 2>&1 | tee_log

if ! print_manual_checklist; then
  failures=$((failures + 1))
fi

echo ""
echo "Guardian QA smoke-all finished — ${failures} failure(s) (preflight + ${#SUITES[@]} suites + manual)"
echo "Compare archives in ${ARCHIVE_DIR}/"
if [[ "${failures}" -gt 0 ]]; then
  exit 1
fi
