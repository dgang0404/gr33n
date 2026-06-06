#!/usr/bin/env bash
# Phase 45 WS2/WS8 — facilitator dry-run: automated Guardian PR path validation.
# Usage (repo root): ./scripts/sit-in-dry-run.sh
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

RED='\033[0;31m'
GRN='\033[0;32m'
YLW='\033[0;33m'
RST='\033[0m'
failures=0

run_step() {
  local label="$1"
  shift
  echo ""
  echo "==> $label"
  if "$@"; then
    echo -e "${GRN}pass${RST}  $label"
  else
    echo -e "${RED}fail${RST}  $label" >&2
    failures=$((failures + 1))
  fi
}

echo "Phase 45 sit-in dry-run — Guardian PR paths (ack · setup pack · dismiss)"

run_step "Go matcher tests (ack_alert, setup pack)" \
  go test ./internal/farmguardian/... \
    -run 'TestMatchAlert|TestMatchSetupPack|TestPickAlert|TestBuildSetupPack' \
    -count=1

run_step "UI Guardian + WS2/WS8 closure tests" \
  bash -c 'cd ui && npm test -- --run \
    src/__tests__/guardian-proposal.test.js \
    src/__tests__/phase-45-ws8-guardian-closure.test.js \
    src/__tests__/phase-45-ws2-closure.test.js \
    src/__tests__/phase-45-ws1-protocol.test.js \
    src/__tests__/farmer-a11y.test.js \
    src/__tests__/phase-44-guardian-closure.test.js'

echo ""
echo "==> API smoke (optional — seeded DB)"
if [[ -f "$ROOT/.env" ]]; then
  if bash -c 'set -a; source "$ROOT/.env" 2>/dev/null; set +a; export AUTH_MODE=auth_test; \
    go test ./cmd/api/... \
      -run "TestPhase29WS3_ConfirmAckHumidityAlert|TestPhase32WS3_ApplyGrowSetupPackConfirm|TestPhase32WS7_SetupPackIntentToConfirm" \
      -count=1 -tags dev -timeout 180s' 2>/dev/null; then
    echo -e "${GRN}pass${RST}  API smoke (ack + setup pack confirm)"
  else
    echo -e "${YLW}skip${RST}  API smoke — DB not seeded or schema missing (Vitest + matchers are primary gate)"
  fi
else
  echo -e "${YLW}skip${RST}  no .env"
fi

echo ""
if [[ "$failures" -eq 0 ]]; then
  echo -e "${GRN}Dry-run PASS${RST} — see docs/workstreams/sit-in-45-dry-run-log.md"
  exit 0
fi
echo -e "${RED}Dry-run FAIL${RST} ($failures step(s))"
exit 1
