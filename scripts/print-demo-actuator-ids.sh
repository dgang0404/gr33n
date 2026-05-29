#!/usr/bin/env bash
# Print master_seed demo relay device + actuator ids for pi_client actuators[] (farm_id=1).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DB_URL="${DATABASE_URL:-postgres://gr33n:gr33n@127.0.0.1:5433/gr33n?sslmode=disable}"
FARM_ID="${GR33N_FARM_ID:-1}"
DEVICE_UID="${GR33N_DEVICE_UID:-demo-veg-relay-01}"

if [[ -f "$ROOT/.env" ]]; then
  # shellcheck disable=SC1091
  val="$(grep -E '^DATABASE_URL=' "$ROOT/.env" | tail -1 | cut -d= -f2- || true)"
  [[ -n "$val" ]] && DB_URL="$val"
fi

echo "farm_id=${FARM_ID} demo relay (device_uid=${DEVICE_UID}):"
echo "DATABASE_URL=${DB_URL}"
echo ""

psql "$DB_URL" -v ON_ERROR_STOP=1 <<SQL
SELECT d.id AS device_id,
       d.name AS device_name,
       d.device_uid,
       a.id AS actuator_id,
       a.name AS actuator_name,
       a.actuator_type
FROM gr33ncore.devices d
JOIN gr33ncore.actuators a
  ON a.device_id = d.id AND a.deleted_at IS NULL
WHERE d.farm_id = ${FARM_ID}
  AND d.device_uid = '${DEVICE_UID}'
  AND d.deleted_at IS NULL
ORDER BY a.id;
SQL

echo ""
echo "Use in pi_client actuators[] (bench GPIO pin is operator choice; default BCM 17):"
echo '  - actuator_id: <actuator_id>'
echo '  device_id:     <device_id>'
echo '  device_type:   light'
echo '  gpio_pin:      17'
