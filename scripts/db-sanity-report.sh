#!/usr/bin/env bash
# Read-only Postgres report: extensions, duplicate zones (seed hazards), farm count, RAG chunk count.
# Exit 1 if DB unreachable, vector missing, or duplicate zone names exist (master_seed.sql assumes unique names).
# Usage (repo root): ./scripts/db-sanity-report.sh
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

ENV_FILE="$ROOT/.env"
SQL_FILE="$ROOT/scripts/sql/db_sanity_report.sql"

die() {
  echo "error: $*" >&2
  exit 1
}

[[ -f "$ENV_FILE" ]] || die "missing .env — copy .env.example"
[[ -f "$SQL_FILE" ]] || die "missing $SQL_FILE"

DATABASE_URL=""
if grep -qE '^[[:space:]]*DATABASE_URL=' "$ENV_FILE"; then
  DATABASE_URL="$(grep -E '^[[:space:]]*DATABASE_URL=' "$ENV_FILE" | head -1 | sed 's/^[[:space:]]*DATABASE_URL=//' | sed 's/^"\(.*\)"$/\1/' | sed "s/^'\(.*\)'$/\1/")"
fi
[[ -n "${DATABASE_URL:-}" ]] || die "DATABASE_URL empty in .env"

command -v psql >/dev/null 2>&1 || die "psql not found"

echo "==> DB sanity report ($(echo "$DATABASE_URL" | sed -E 's|//([^:]+):([^@]*)@|//\1:***@|'))"
# ponytail: 3× retry on psql exit 3 (ON_ERROR_STOP) — dev API/automation can lock sensors mid-report; upgrade: pause API before sanity
_run_sanity_sql() {
  psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f "$SQL_FILE"
}
attempt=1
while ! _run_sanity_sql; do
  rc=$?
  if [[ "$rc" -ne 3 || "$attempt" -ge 3 ]]; then
    exit "$rc"
  fi
  echo "warn: sanity SQL failed (exit $rc) — retry $((attempt + 1))/3 in 2s (stop API if this persists)" >&2
  sleep 2
  attempt=$((attempt + 1))
done

dup="$(psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -tAc "
SELECT count(*) FROM (
  SELECT farm_id, name FROM gr33ncore.zones GROUP BY farm_id, name HAVING count(*) > 1
) t;" | tr -d '[:space:]')"

if [[ "${dup:-0}" != "0" ]]; then
  echo ""
  echo "error: duplicate zone names detected ($dup distinct farm_id+name groups). This breaks db/seeds/master_seed.sql" >&2
  echo "  Fix: merge or delete duplicate zones, or docker compose down -v && ./scripts/dev-stack.sh for a fresh DB (destructive)." >&2
  exit 1
fi

sensor_dup="$(psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -tAc "
SELECT count(*) FROM (
  SELECT farm_id, name FROM gr33ncore.sensors
  WHERE deleted_at IS NULL
  GROUP BY farm_id, name HAVING count(*) > 1
) t;" | tr -d '[:space:]')"

if [[ "${sensor_dup:-0}" != "0" ]]; then
  echo ""
  echo "error: duplicate active sensor names detected ($sensor_dup groups). Run migration phase48 or ./scripts/dev-reset-farm.sh" >&2
  exit 1
fi

farm1_sensors="$(psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -tAc "
SELECT count(*) FROM gr33ncore.sensors WHERE farm_id = 1 AND deleted_at IS NULL;" | tr -d '[:space:]')"

if [[ -n "${farm1_sensors:-}" && "${farm1_sensors}" -gt 24 ]]; then
  echo ""
  echo "warn: farm 1 has ${farm1_sensors} active sensors (>24) — consider ./scripts/dev-reset-farm.sh --profile small_indoor" >&2
fi

gpio_conflicts="$(psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -tAc "
SELECT count(*) FROM (
  SELECT farm_id, device_id, gpio_pin
  FROM (
    SELECT s.farm_id,
           (s.config->'wiring'->>'device_id')::bigint AS device_id,
           (s.config->'wiring'->>'gpio_pin')::int AS gpio_pin,
           'sensor' AS entity_kind,
           COALESCE(s.config->'wiring'->>'source', '') AS source
    FROM gr33ncore.sensors s
    WHERE s.deleted_at IS NULL
      AND s.config->'wiring'->>'device_id' IS NOT NULL
      AND s.config->'wiring'->>'gpio_pin' IS NOT NULL
    UNION ALL
    SELECT a.farm_id,
           (a.config->'wiring'->>'device_id')::bigint,
           (a.config->'wiring'->>'gpio_pin')::int,
           'actuator',
           COALESCE(a.config->'wiring'->>'source', 'gpio_relay')
    FROM gr33ncore.actuators a
    WHERE a.deleted_at IS NULL
      AND a.config->'wiring'->>'device_id' IS NOT NULL
      AND a.config->'wiring'->>'gpio_pin' IS NOT NULL
  ) gpio_usage
  GROUP BY farm_id, device_id, gpio_pin
  HAVING count(*) > 1
     AND NOT (count(*) = count(*) FILTER (WHERE entity_kind = 'sensor' AND source = 'dht22'))
) t;" | tr -d '[:space:]')"

if [[ "${gpio_conflicts:-0}" != "0" ]]; then
  echo ""
  echo "error: GPIO pin conflicts on edge devices ($gpio_conflicts groups). Fix wiring in Sensors / Controls or SQL." >&2
  exit 1
fi

i2c_conflicts="$(psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -tAc "
SELECT count(*) FROM (
  SELECT farm_id, device_id, i2c_channel
  FROM (
    SELECT s.farm_id,
           (s.config->'wiring'->>'device_id')::bigint AS device_id,
           (s.config->'wiring'->>'i2c_channel')::int AS i2c_channel
    FROM gr33ncore.sensors s
    WHERE s.deleted_at IS NULL
      AND s.config->'wiring'->>'device_id' IS NOT NULL
      AND s.config->'wiring'->>'i2c_channel' IS NOT NULL
  ) ch_usage
  GROUP BY farm_id, device_id, i2c_channel
  HAVING count(*) > 1
) t;" | tr -d '[:space:]')"

if [[ "${i2c_conflicts:-0}" != "0" ]]; then
  echo ""
  echo "error: I2C channel conflicts on edge devices ($i2c_conflicts groups)." >&2
  exit 1
fi

bad_source="$(psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -tAc "
SELECT count(*) FROM (
  SELECT id FROM gr33ncore.sensors
  WHERE deleted_at IS NULL
    AND config->'wiring'->>'source' IS NOT NULL
    AND config->'wiring'->>'source' NOT IN ('dht22', 'ads1115', 'mhz19', 'bh1750', 'derived', 'gpio_digital')
  UNION ALL
  SELECT id FROM gr33ncore.actuators
  WHERE deleted_at IS NULL
    AND config->'wiring'->>'gpio_pin' IS NOT NULL
    AND COALESCE(config->'wiring'->>'source', 'gpio_relay') NOT IN ('gpio_relay')
) t;" | tr -d '[:space:]')"

if [[ "${bad_source:-0}" != "0" ]]; then
  echo ""
  echo "error: unsupported wiring.source on $bad_source sensor/actuator row(s)." >&2
  exit 1
fi

derived_bad="$(psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -tAc "
SELECT count(*) FROM gr33ncore.sensors s
CROSS JOIN LATERAL jsonb_each_text(s.config->'wiring'->'inputs') AS inp(key, value)
WHERE s.deleted_at IS NULL
  AND s.config->'wiring'->>'source' = 'derived'
  AND NOT EXISTS (
      SELECT 1 FROM gr33ncore.sensors t
      WHERE t.farm_id = s.farm_id AND t.id = (inp.value)::bigint AND t.deleted_at IS NULL
  );" | tr -d '[:space:]')"

if [[ "${derived_bad:-0}" != "0" ]]; then
  echo ""
  echo "error: derived sensor wiring references $derived_bad missing input sensor id(s)." >&2
  exit 1
fi

echo ""
echo "ok  no duplicate zone or sensor names; no wiring pin/channel conflicts"
