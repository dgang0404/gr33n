#!/usr/bin/env bash
# Print master_seed sensor names → numeric ids for pi_client config (farm_id=1).
# Uses lowest id per name when duplicate seed rows exist.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DB_URL="${DATABASE_URL:-postgres://gr33n:gr33n@127.0.0.1:5433/gr33n?sslmode=disable}"

if [[ -f "$ROOT/.env" ]]; then
  # shellcheck disable=SC1091
  val="$(grep -E '^DATABASE_URL=' "$ROOT/.env" | tail -1 | cut -d= -f2- || true)"
  [[ -n "$val" ]] && DB_URL="$val"
fi

echo "farm_id=1 sensor ids (use in pi_client/config.yaml or config.demo-stub.yaml):"
echo "DATABASE_URL=${DB_URL}"
echo ""

psql "$DB_URL" -v ON_ERROR_STOP=1 <<'SQL'
SELECT s.id, s.name, s.sensor_type
FROM gr33ncore.sensors s
INNER JOIN (
    SELECT name, MIN(id) AS id
    FROM gr33ncore.sensors
    WHERE farm_id = 1 AND deleted_at IS NULL
      AND name IN (
          'PAR Sensor Indoor', 'Lux Sensor Indoor', 'Air Temp Indoor', 'Root Zone Temp',
          'Air Humidity Indoor', 'Soil Moisture Outdoor', 'Media Moisture Indoor',
          'EC Sensor', 'pH Sensor', 'CO2 Sensor Indoor'
      )
    GROUP BY name
) pick ON pick.id = s.id
ORDER BY s.id;
SQL

echo ""
echo "Pi client mapping (fresh master_seed insert order):"
echo "  3 Air Temp Indoor (temperature)  5 Air Humidity Indoor (humidity)"
echo "  1 PAR  8 EC  9 pH  10 CO2  6/7 soil moisture — see config.demo-stub.yaml"
