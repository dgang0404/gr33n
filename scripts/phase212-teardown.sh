#!/usr/bin/env bash
# Phase 212 WS7 — stop Install B + receiver; optional clone delete.
# Does not wipe Install A DB. Revert Insert Commons .env knobs yourself if desired.
set -euo pipefail

INSTALL_B_DIR="${INSTALL_B_DIR:-$HOME/gr33n-platform-b}"
DELETE_CLONE="${DELETE_CLONE:-0}"

kill_port() {
  local port="$1"
  local pids
  pids=$(ss -ltnp 2>/dev/null | awk -v p=":$port" '$4 ~ p {print}' | grep -oP 'pid=\K[0-9]+' | sort -u || true)
  if [[ -z "${pids:-}" ]]; then
    return 0
  fi
  # shellcheck disable=SC2086
  echo "    killing pid(s) on :$port → $pids"
  # shellcheck disable=SC2086
  kill $pids 2>/dev/null || true
}

echo "==> Stop Insert Commons receiver (:8765)"
pkill -f 'cmd/insert-commons-receiver' 2>/dev/null || true
pkill -f 'insert-commons-receiver' 2>/dev/null || true
kill_port 8765

echo "==> Stop Install B host API (:8081)"
pkill -f 'gr33n-platform-b.*cmd/api' 2>/dev/null || true
kill_port 8081
if ss -ltn 2>/dev/null | grep -q ':8081'; then
  echo "    WARN: :8081 still listening after kill — check manually"
fi

if [[ -d "$INSTALL_B_DIR" ]]; then
  echo "==> Stop Install B Compose DB"
  (cd "$INSTALL_B_DIR" && docker compose -f docker-compose.phase212-b.yml down) || true
fi

if [[ "$DELETE_CLONE" == "1" ]]; then
  echo "==> Removing $INSTALL_B_DIR"
  rm -rf "$INSTALL_B_DIR"
else
  echo "==> Clone kept at $INSTALL_B_DIR (DELETE_CLONE=1 to remove)"
fi

echo "Done. Install A (:8080 / :5433) left running if it was up."
echo "Optional: unset INSERT_COMMONS_* in Install A .env and restart API."
