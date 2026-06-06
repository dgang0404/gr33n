#!/usr/bin/env bash
# Phase 45 WS8 prep — facilitator checks before farmer sit-in sessions.
# Usage (repo root): ./scripts/sit-in-preflight.sh [--mobile]
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

MOBILE=0
while [[ $# -gt 0 ]]; do
  case "$1" in
    --mobile) MOBILE=1 ;;
    -h|--help)
      echo "Usage: scripts/sit-in-preflight.sh [--mobile]"
      echo "  Runs check-local-stack.sh, then prints sit-in protocol links."
      exit 0
      ;;
    *) echo "unknown option: $1" >&2; exit 1 ;;
  esac
  shift
done

echo "==> Stack checks (DB, vector, API /health)"
"$ROOT/scripts/check-local-stack.sh"

ENV_FILE="$ROOT/.env"
PORT_VAL="8080"
if [[ -f "$ENV_FILE" ]] && grep -qE '^[[:space:]]*PORT=' "$ENV_FILE"; then
  PORT_VAL="$(grep -E '^[[:space:]]*PORT=' "$ENV_FILE" | head -1 | sed 's/^[[:space:]]*PORT=//' | tr -d \"\' )"
fi

echo ""
echo "==> Phase 45 sit-in facilitator kit"
echo "Protocol:     docs/workstreams/farmer-sit-in-protocol.md"
echo "Scorecard:    docs/workstreams/sit-in-45-session-log-template.md"
echo "Guardian PR:  docs/plans/phase_45_guardian_pr_spec.md"
echo "Friction log: docs/workstreams/phase-45-ws2-friction-backlog.md"
echo ""
echo "Session A — demo farm 1: http://localhost:5173/  (log in as operator+)"
echo "Session B — new farm via Dashboard setup wizard"
echo "Three PR paths: ack_alert · apply_grow_setup_pack · dismiss (protocol §4)"

if curl -sf "http://127.0.0.1:${PORT_VAL}/health" >/dev/null 2>&1; then
  echo ""
  echo "==> Optional: Guardian field health (farm 1)"
  if curl -sf "http://127.0.0.1:${PORT_VAL}/v1/chat/health?farm_id=1" 2>/dev/null | head -c 200; then
    echo ""
  else
    echo "(skipped — needs JWT; open UI and confirm Guardian drawer opens)"
  fi
fi

if [[ "$MOBILE" -eq 1 ]]; then
  echo ""
  "$ROOT/scripts/mobile-sit-in-prep.sh"
fi

echo ""
echo "Ready when ≥2 testers are scheduled. File friction in phase-45-ws2-friction-backlog.md."
