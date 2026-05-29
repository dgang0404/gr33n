#!/usr/bin/env bash
# Phase 31 WS3 — end-to-end: pending_command → pi_client → actuator_events → clear pending.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DB_URL="${DATABASE_URL:-postgres://gr33n:gr33n@127.0.0.1:5433/gr33n?sslmode=disable}"
API_URL="${GR33N_API_URL:-http://localhost:8080}"
FARM_ID="${GR33N_FARM_ID:-1}"
PI_KEY="${PI_API_KEY:-}"
LOGIN_USER="${GR33N_LOGIN_USER:-dev@gr33n.local}"
LOGIN_PASS="${GR33N_LOGIN_PASS:-devpassword}"
MODE="direct"
COMMAND="on"
WAIT_SEC="${GR33N_ACTUATOR_WAIT_SEC:-45}"

usage() {
  cat <<EOF
Phase 31 WS3 — safe actuator round-trip smoke

Usage: $(basename "$0") [--direct|--guardian] [--command on|off] [--wait SEC]

Requires: API running (make dev-auth-test), seeded demo relay (demo-veg-relay-01),
          PI_API_KEY in .env, python3 + psql.

  --direct     Enqueue pending_command via Postgres (default) — bench / CI-friendly
  --guardian   Insert Guardian proposal + POST /v1/chat/confirm (needs login password)
  --command    on (default) or off
  --wait       Seconds to run pi_client and poll for actuator_events (default 45)

Manual steps (same flow):
  1. Terminal A: ./scripts/run-edge-actuator-client.sh
  2. Terminal B: ./scripts/enqueue-demo-pending-command.sh on
  3. Watch client logs + dashboard Devices / actuator history

See docs/pi-integration-guide.md §9 and make edge-actuator-smoke-help
EOF
}

load_env() {
  if [[ -f "$ROOT/.env" ]]; then
    # shellcheck disable=SC1091
    while IFS= read -r line || [[ -n "$line" ]]; do
      [[ "$line" =~ ^[[:space:]]*# ]] && continue
      [[ "$line" =~ ^DATABASE_URL= ]] && DB_URL="${line#DATABASE_URL=}"
      [[ "$line" =~ ^PI_API_KEY= ]] && PI_KEY="${line#PI_API_KEY=}"
      [[ "$line" =~ ^PORT= ]] && port="${line#PORT=}" && API_URL="http://localhost:${port}"
    done < "$ROOT/.env"
  fi
  if [[ -z "$PI_KEY" ]]; then
    echo "PI_API_KEY not set in .env" >&2
    exit 1
  fi
}

lookup_ids() {
  read -r DEVICE_ID ACTUATOR_ID <<<"$(psql "$DB_URL" -v ON_ERROR_STOP=1 -t -A -F' ' -c \
    "SELECT d.id, a.id
     FROM gr33ncore.devices d
     JOIN gr33ncore.actuators a ON a.device_id = d.id AND a.deleted_at IS NULL
     WHERE d.farm_id = ${FARM_ID} AND d.device_uid = 'demo-veg-relay-01' AND d.deleted_at IS NULL
     ORDER BY a.id LIMIT 1")"
  if [[ -z "${DEVICE_ID:-}" || "$DEVICE_ID" == "NULL" ]]; then
    echo "demo-veg-relay-01 missing — run make dev-stack --seed" >&2
    exit 1
  fi
}

event_count() {
  psql "$DB_URL" -t -A -c \
    "SELECT COUNT(*) FROM gr33ncore.actuator_events WHERE actuator_id = ${ACTUATOR_ID}"
}

pending_present() {
  psql "$DB_URL" -t -A -c \
    "SELECT CASE WHEN config ? 'pending_command' THEN 'yes' ELSE 'no' END
     FROM gr33ncore.devices WHERE id = ${DEVICE_ID}"
}

login_jwt() {
  local resp token
  resp=$(curl -sf -X POST "${API_URL}/auth/login" \
    -H 'Content-Type: application/json' \
    -d "{\"username\":\"${LOGIN_USER}\",\"password\":\"${LOGIN_PASS}\"}") \
    || { echo "Login failed (${LOGIN_USER}). Set GR33N_LOGIN_USER/PASS or use --direct." >&2; exit 1; }
  token=$(python3 -c "import json,sys; print(json.load(sys.stdin).get('token',''))" <<<"$resp")
  if [[ -z "$token" ]]; then
    echo "No JWT in login response" >&2
    exit 1
  fi
  echo "$token"
}

enqueue_guardian() {
  local jwt="$1"
  local proposal_id resp
  proposal_id=$(DEVICE_ID="$DEVICE_ID" ACTUATOR_ID="$ACTUATOR_ID" FARM_ID="$FARM_ID" COMMAND="$COMMAND" DB_URL="$DB_URL" python3 - <<'PY'
import json, os, subprocess

db_url = os.environ["DB_URL"]
farm_id = int(os.environ["FARM_ID"])
device_id = int(os.environ["DEVICE_ID"])
actuator_id = int(os.environ["ACTUATOR_ID"])
command = os.environ["COMMAND"]
args = json.dumps({
    "device_id": device_id,
    "actuator_id": actuator_id,
    "command": command,
    "reason": "Phase 31 WS3 guardian smoke — operator bench test",
})
summary = f"WS3 smoke: turn {command} Veg Room Grow Light"
sql = f"""
INSERT INTO gr33ncore.guardian_action_proposals
    (user_id, farm_id, tool_id, args, summary, risk_tier, expires_at)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    {farm_id},
    'enqueue_actuator_command',
    $args${args}$args$::jsonb,
    '{summary.replace("'", "''")}',
    'high',
    NOW() + INTERVAL '10 minutes'
)
RETURNING proposal_id::text;
"""
out = subprocess.check_output(["psql", "-q", db_url, "-v", "ON_ERROR_STOP=1", "-t", "-A", "-c", sql], text=True)
print(out.strip())
PY
)
  resp=$(curl -sf -X POST "${API_URL}/v1/chat/confirm" \
    -H "Authorization: Bearer ${jwt}" \
    -H 'Content-Type: application/json' \
    -d "{\"proposal_id\":\"${proposal_id}\"}") \
    || { echo "Guardian confirm failed for proposal ${proposal_id}" >&2; exit 1; }
  echo "Guardian confirm OK proposal_id=${proposal_id}"
}

run_client_background() {
  "$ROOT/scripts/run-edge-actuator-client.sh" &
  CLIENT_PID=$!
  trap 'kill "${CLIENT_PID}" 2>/dev/null || true' EXIT
}

wait_for_roundtrip() {
  local before="$1"
  local deadline=$((SECONDS + WAIT_SEC))
  echo "Waiting up to ${WAIT_SEC}s for actuator_events + pending clear..."
  while (( SECONDS < deadline )); do
    local now pending
    now=$(event_count)
    pending=$(pending_present)
    if [[ "$now" -gt "$before" && "$pending" == "no" ]]; then
      echo "OK — new actuator_event (count ${before} → ${now}), pending_command cleared."
      psql "$DB_URL" -c \
        "SELECT event_time, command_sent, source, meta_data::text
         FROM gr33ncore.actuator_events
         WHERE actuator_id = ${ACTUATOR_ID}
         ORDER BY event_time DESC LIMIT 1"
      return 0
    fi
    sleep 1
  done
  echo "FAIL — pending=$(pending_present) events before=${before} after=$(event_count)" >&2
  return 1
}

main() {
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --direct) MODE=direct; shift ;;
      --guardian) MODE=guardian; shift ;;
      --command) COMMAND="$2"; shift 2 ;;
      --wait) WAIT_SEC="$2"; shift 2 ;;
      -h|--help) usage; exit 0 ;;
      *) echo "Unknown option: $1" >&2; usage >&2; exit 1 ;;
    esac
  done

  if [[ "$COMMAND" != "on" && "$COMMAND" != "off" ]]; then
    echo "--command must be on or off" >&2
    exit 1
  fi

  load_env
  lookup_ids

  echo "WS3 actuator smoke mode=${MODE} command=${COMMAND} device_id=${DEVICE_ID} actuator_id=${ACTUATOR_ID}"

  "$ROOT/scripts/enqueue-demo-pending-command.sh" --clear >/dev/null
  BEFORE=$(event_count)

  run_client_background
  sleep 2

  if [[ "$MODE" == "guardian" ]]; then
    JWT=$(login_jwt)
    enqueue_guardian "$JWT"
  else
    PENDING_SOURCE=bench "$ROOT/scripts/enqueue-demo-pending-command.sh" "$COMMAND"
  fi

  if [[ "$(pending_present)" != "yes" ]]; then
    echo "FAIL — pending_command not set after enqueue" >&2
    exit 1
  fi

  wait_for_roundtrip "$BEFORE"
}

main "$@"
