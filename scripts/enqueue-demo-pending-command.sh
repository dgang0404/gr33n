#!/usr/bin/env bash
# Phase 31 WS3 — enqueue devices.config.pending_command (same JSON shape as automation / Guardian).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DB_URL="${DATABASE_URL:-postgres://gr33n:gr33n@127.0.0.1:5433/gr33n?sslmode=disable}"
FARM_ID="${GR33N_FARM_ID:-1}"
DEVICE_UID="${GR33N_DEVICE_UID:-demo-veg-relay-01}"
COMMAND="${1:-}"
SOURCE="${PENDING_SOURCE:-bench}"
REASON="${PENDING_REASON:-Phase 31 WS3 bench test}"

usage() {
  cat <<EOF
Usage: $(basename "$0") on|off [--clear]

Enqueue pending_command on the demo relay device (${DEVICE_UID}) via Postgres.
Same JSON field the automation worker and Guardian enqueue_actuator_command write.

  on|off     Command for the Pi client to execute
  --clear    Remove pending_command without enqueueing

Env: DATABASE_URL, GR33N_FARM_ID, GR33N_DEVICE_UID, PENDING_SOURCE, PENDING_REASON
Guardian-shaped enqueue: PENDING_SOURCE=guardian PENDING_REASON="operator inspection" $(basename "$0") on
EOF
}

load_env() {
  if [[ -f "$ROOT/.env" ]]; then
    # shellcheck disable=SC1091
    val="$(grep -E '^DATABASE_URL=' "$ROOT/.env" | tail -1 | cut -d= -f2- || true)"
    [[ -n "$val" ]] && DB_URL="$val"
  fi
}

lookup_ids() {
  read -r DEVICE_ID ACTUATOR_ID <<<"$(psql "$DB_URL" -v ON_ERROR_STOP=1 -t -A -F' ' -c \
    "SELECT d.id, a.id
     FROM gr33ncore.devices d
     JOIN gr33ncore.actuators a ON a.device_id = d.id AND a.deleted_at IS NULL
     WHERE d.farm_id = ${FARM_ID} AND d.device_uid = '${DEVICE_UID}' AND d.deleted_at IS NULL
     ORDER BY a.id LIMIT 1")"
  if [[ -z "${DEVICE_ID:-}" || "$DEVICE_ID" == "NULL" ]]; then
    echo "Device '${DEVICE_UID}' not found for farm_id=${FARM_ID}. Run make dev-stack --seed." >&2
    exit 1
  fi
}

clear_pending() {
  psql "$DB_URL" -v ON_ERROR_STOP=1 -c \
    "UPDATE gr33ncore.devices SET config = config - 'pending_command', updated_at = NOW() WHERE id = ${DEVICE_ID}"
  echo "Cleared pending_command on device_id=${DEVICE_ID}"
}

enqueue() {
  local cmd="$1"
  DEVICE_ID="$DEVICE_ID" ACTUATOR_ID="$ACTUATOR_ID" SOURCE="$SOURCE" REASON="$REASON" CMD="$cmd" DB_URL="$DB_URL" python3 - <<'PY'
import json, os, subprocess

db_url = os.environ["DB_URL"]
device_id = int(os.environ["DEVICE_ID"])
actuator_id = int(os.environ["ACTUATOR_ID"])
cmd = os.environ["CMD"]
payload = {
    "command": cmd,
    "actuator_id": actuator_id,
    "source": os.environ["SOURCE"],
    "reason": os.environ["REASON"],
}
pending = json.dumps(payload)
sql = f"""
UPDATE gr33ncore.devices
SET config = jsonb_set(coalesce(config, '{{}}'::jsonb), '{{pending_command}}', $pending${pending}$pending$::jsonb),
    updated_at = NOW()
WHERE id = {device_id};
"""
subprocess.run(["psql", db_url, "-v", "ON_ERROR_STOP=1", "-c", sql], check=True)
print(f"Enqueued pending_command={cmd} device_id={device_id} actuator_id={actuator_id} source={payload['source']}")
PY
}

main() {
  load_env
  lookup_ids

  if [[ "${1:-}" == "--clear" ]]; then
    clear_pending
    exit 0
  fi

  if [[ "$COMMAND" != "on" && "$COMMAND" != "off" ]]; then
    usage >&2
    exit 1
  fi

  enqueue "$COMMAND"
}

main "$@"
