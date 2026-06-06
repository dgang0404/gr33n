#!/usr/bin/env bash
# Phase 45 WS4 — print LAN URLs for farmer sit-in Session C (PWA on phone).
# Usage (repo root): ./scripts/mobile-sit-in-prep.sh [--api-port 8080] [--ui-port 5173]
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
API_PORT=8080
UI_PORT=5173

while [[ $# -gt 0 ]]; do
  case "$1" in
    --api-port) API_PORT="$2"; shift 2 ;;
    --ui-port) UI_PORT="$2"; shift 2 ;;
    -h|--help)
      echo "Usage: scripts/mobile-sit-in-prep.sh [--api-port 8080] [--ui-port 5173]"
      exit 0
      ;;
    *) echo "unknown option: $1" >&2; exit 1 ;;
  esac
done

LAN_IP="$(hostname -I 2>/dev/null | awk '{print $1}')"
if [[ -z "${LAN_IP:-}" ]]; then
  echo "error: could not detect LAN IP (hostname -I empty)" >&2
  exit 1
fi

UI_URL="http://${LAN_IP}:${UI_PORT}"
API_URL="http://${LAN_IP}:${API_PORT}"

cat <<EOF
Phase 45 — mobile sit-in prep (Session C)

1. Phone and laptop on the same Wi‑Fi.
2. API CORS — in .env set:
     CORS_ORIGIN=${UI_URL}
   Restart the API after changing CORS.
3. UI dev server (from ui/):
     npm run dev -- --host 0.0.0.0 --port ${UI_PORT}
4. On the phone browser open:
     ${UI_URL}
5. Install PWA — Add to Home Screen (iOS Safari) or Install app (Android Chrome).
6. Run Session C from docs/workstreams/farmer-sit-in-protocol.md (Guardian Confirm/Dismiss taps).

Capacitor sideload (optional): ./scripts/cap-lan-build.sh then npm run cap:open:android

Full doc: docs/workstreams/phase-45-ws4-mobile-sit-in-path.md
EOF
