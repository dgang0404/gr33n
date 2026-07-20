#!/usr/bin/env bash
# Run all non-redundant Guardian QA smoke suites sequentially (laptop / self-hosted).
# Skips single-prompt reruns (smoke-ec-ph, write-ack-only, etc.) — those are for debugging one failure.
#
# Usage (repo root):
#   make guardian-qa-smoke-all
#   GUARDIAN_QA_UI=1 make guardian-qa-smoke-all          # + multi-turn UI quick (~50 min)
#   GUARDIAN_QA_UI_FULL=1 make guardian-qa-smoke-all     # + full change-requests-ui (~2–3 hr)
#   GUARDIAN_QA_FAIL_FAST=1 make guardian-qa-smoke-all   # stop on first suite failure
set -uo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

MODEL="${MODEL:-phi3:mini}"
FARM_ID="${FARM_ID:-1}"
LOG="${GUARDIAN_QA_ALL_LOG:-/tmp/guardian-qa-smoke-all.log}"
FAIL_FAST="${GUARDIAN_QA_FAIL_FAST:-0}"

if [[ -f .env ]]; then set -a && . ./.env && set +a; fi
# shellcheck disable=SC1091
source scripts/source-local-env.sh --refresh-eval-token

run_suite() {
  local target="$1"
  local label="$2"
  echo ""
  echo "================================================================"
  echo "==> ${label}"
  echo "    make ${target} MODEL=${MODEL} FARM_ID=${FARM_ID}"
  echo "================================================================"
  if make "${target}" MODEL="${MODEL}" FARM_ID="${FARM_ID}"; then
    echo "==> ${label}: OK"
    return 0
  fi
  echo "==> ${label}: FAILED" >&2
  return 1
}

SUITES=(
  "guardian-qa-smoke|Q&A smoke (4 prompts, ~90 min CPU)"
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
echo "Archives: data/guardian_qa_runs/"
echo "Suites: ${#SUITES[@]}"

failures=0
: >"${LOG}"

for entry in "${SUITES[@]}"; do
  target="${entry%%|*}"
  label="${entry#*|}"
  if run_suite "${target}" "${label}" 2>&1 | tee -a "${LOG}"; then
    :
  else
    failures=$((failures + 1))
    if [[ "${FAIL_FAST}" == "1" ]]; then
      echo "FAIL_FAST=1 — stopping after first suite failure" >&2
      break
    fi
  fi
done

echo ""
echo "Guardian QA smoke-all finished — ${failures} suite(s) failed (of ${#SUITES[@]})"
echo "Compare archives in data/guardian_qa_runs/"
if [[ "${failures}" -gt 0 ]]; then
  exit 1
fi
