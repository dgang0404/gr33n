#!/usr/bin/env bash
# Phase 131 — grep API logs for Guardian eval / tool evidence.
# Usage: ./scripts/guardian-qa-scrape-logs.sh [--log /tmp/gr33n-api.log] [--expect walk_farm] [--eval-id smoke-morning-walk]
set -euo pipefail

LOG="${GUARDIAN_EVAL_LOG:-/tmp/gr33n-api.log}"
EXPECT=""
EVAL_ID=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --log) LOG="$2"; shift 2 ;;
    --expect) EXPECT="$2"; shift 2 ;;
    --eval-id) EVAL_ID="$2"; shift 2 ;;
    -h|--help)
      echo "Usage: $0 [--log PATH] [--expect TOOL] [--eval-id FIXTURE_ID]"
      exit 0
      ;;
    *) echo "unknown option: $1" >&2; exit 1 ;;
  esac
done

[[ -f "$LOG" ]] || { echo "log not found: $LOG" >&2; exit 1; }

PATTERN="guardian:"
[[ -n "$EXPECT" ]] && PATTERN="tool_id=${EXPECT}|${EXPECT}"
[[ -n "$EVAL_ID" ]] && PATTERN="${EVAL_ID}|${PATTERN}"

echo "==> Scraping $LOG for: $PATTERN"
if grep -E "$PATTERN" "$LOG" | tail -20; then
  exit 0
fi
echo "(no matches)"
exit 1
