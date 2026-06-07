-- Read-only sanity checks for local / staging Postgres (farm_id=1 demo assumptions).
-- Run: psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f scripts/sql/db_sanity_report.sql
\set ON_ERROR_STOP on

\echo '==> Extensions (expect vector + timescaledb + postgis)'
SELECT extname, extversion
FROM pg_extension
WHERE extname IN ('vector', 'timescaledb', 'postgis')
ORDER BY extname;

\echo ''
\echo '==> Farm count'
SELECT count(*) AS farms FROM gr33ncore.farms;

\echo ''
\echo '==> Dev seed profile (farm 1)'
SELECT id, name, COALESCE(meta_data->>'dev_seed_profile', '(unset)') AS dev_seed_profile
FROM gr33ncore.farms
WHERE id = 1;

\echo ''
\echo '==> Active sensors per farm'
SELECT farm_id, count(*) AS active_sensors
FROM gr33ncore.sensors
WHERE deleted_at IS NULL
GROUP BY farm_id
ORDER BY farm_id;

\echo ''
\echo '==> Duplicate zone names per farm (breaks master_seed.sql subqueries)'
SELECT farm_id, name, count(*) AS cnt
FROM gr33ncore.zones
GROUP BY farm_id, name
HAVING count(*) > 1
ORDER BY farm_id, name;

\echo ''
\echo '==> Duplicate active sensor names per farm (Phase 48)'
SELECT farm_id, name, count(*) AS cnt
FROM gr33ncore.sensors
WHERE deleted_at IS NULL
GROUP BY farm_id, name
HAVING count(*) > 1
ORDER BY farm_id, name;

\echo ''
\echo '==> sensor_readings approximate row count'
SELECT reltuples::bigint AS sensor_readings_approx
FROM pg_class
WHERE oid = 'gr33ncore.sensor_readings'::regclass;

\echo ''
\echo '==> Profile bloat warning (farm 1 active sensors > 24)'
SELECT farm_id, count(*) AS active_sensors,
       CASE WHEN count(*) > 24 THEN 'WARN bloat' ELSE 'ok' END AS status
FROM gr33ncore.sensors
WHERE deleted_at IS NULL AND farm_id = 1
GROUP BY farm_id;

\echo ''
\echo '==> RAG chunks row count (informational)'
SELECT count(*) AS rag_embedding_chunks FROM gr33ncore.rag_embedding_chunks;

\echo ''
\echo '==> Phase 50 — sensors/actuators with wiring metadata'
SELECT farm_id, count(*) AS sensors_with_wiring
FROM gr33ncore.sensors
WHERE deleted_at IS NULL
  AND config->'wiring'->>'source' IS NOT NULL
GROUP BY farm_id
ORDER BY farm_id;

SELECT farm_id, count(*) AS actuators_with_wiring
FROM gr33ncore.actuators
WHERE deleted_at IS NULL
  AND config->'wiring'->>'gpio_pin' IS NOT NULL
GROUP BY farm_id
ORDER BY farm_id;

\echo ''
\echo '==> Phase 50 — GPIO pin conflicts per edge device (sensors + actuators)'
SELECT farm_id, device_id, gpio_pin, count(*) AS entities, array_agg(label ORDER BY entity_id) AS conflicts
FROM (
    SELECT s.farm_id,
           (s.config->'wiring'->>'device_id')::bigint AS device_id,
           (s.config->'wiring'->>'gpio_pin')::int AS gpio_pin,
           s.id AS entity_id,
           'sensor' AS entity_kind,
           COALESCE(s.config->'wiring'->>'source', '') AS source,
           'sensor:' || s.id || ':' || s.name AS label
    FROM gr33ncore.sensors s
    WHERE s.deleted_at IS NULL
      AND s.config->'wiring'->>'device_id' IS NOT NULL
      AND s.config->'wiring'->>'gpio_pin' IS NOT NULL
    UNION ALL
    SELECT a.farm_id,
           (a.config->'wiring'->>'device_id')::bigint,
           (a.config->'wiring'->>'gpio_pin')::int,
           a.id,
           'actuator',
           COALESCE(a.config->'wiring'->>'source', 'gpio_relay'),
           'actuator:' || a.id || ':' || a.name
    FROM gr33ncore.actuators a
    WHERE a.deleted_at IS NULL
      AND a.config->'wiring'->>'device_id' IS NOT NULL
      AND a.config->'wiring'->>'gpio_pin' IS NOT NULL
) gpio_usage
GROUP BY farm_id, device_id, gpio_pin
HAVING count(*) > 1
   AND NOT (
     count(*) = count(*) FILTER (WHERE entity_kind = 'sensor' AND source = 'dht22')
   )
ORDER BY farm_id, device_id, gpio_pin;

\echo ''
\echo '==> Phase 50 — I2C channel conflicts per edge device (sensors)'
SELECT farm_id, device_id, i2c_channel, count(*) AS sensors, array_agg(sensor_id ORDER BY sensor_id) AS sensor_ids
FROM (
    SELECT s.farm_id,
           (s.config->'wiring'->>'device_id')::bigint AS device_id,
           (s.config->'wiring'->>'i2c_channel')::int AS i2c_channel,
           s.id AS sensor_id
    FROM gr33ncore.sensors s
    WHERE s.deleted_at IS NULL
      AND s.config->'wiring'->>'device_id' IS NOT NULL
      AND s.config->'wiring'->>'i2c_channel' IS NOT NULL
) ch_usage
GROUP BY farm_id, device_id, i2c_channel
HAVING count(*) > 1
ORDER BY farm_id, device_id, i2c_channel;

\echo ''
\echo '==> Phase 50 — unsupported wiring.source values'
SELECT 'sensor' AS entity_type, id, name, config->'wiring'->>'source' AS source
FROM gr33ncore.sensors
WHERE deleted_at IS NULL
  AND config->'wiring'->>'source' IS NOT NULL
  AND config->'wiring'->>'source' NOT IN ('dht22', 'ads1115', 'mhz19', 'bh1750', 'derived', 'gpio_digital')
UNION ALL
SELECT 'actuator', id, name, COALESCE(config->'wiring'->>'source', 'gpio_relay')
FROM gr33ncore.actuators
WHERE deleted_at IS NULL
  AND config->'wiring'->>'gpio_pin' IS NOT NULL
  AND COALESCE(config->'wiring'->>'source', 'gpio_relay') NOT IN ('gpio_relay')
ORDER BY entity_type, id;

\echo ''
\echo '==> Phase 50 — derived sensors with missing input sensor ids'
SELECT s.id AS sensor_id, s.name, inp.key AS input_key, (inp.value)::bigint AS missing_sensor_id
FROM gr33ncore.sensors s
CROSS JOIN LATERAL jsonb_each_text(s.config->'wiring'->'inputs') AS inp(key, value)
WHERE s.deleted_at IS NULL
  AND s.config->'wiring'->>'source' = 'derived'
  AND NOT EXISTS (
      SELECT 1 FROM gr33ncore.sensors t
      WHERE t.farm_id = s.farm_id
        AND t.id = (inp.value)::bigint
        AND t.deleted_at IS NULL
  )
ORDER BY s.id, inp.key;

\echo ''
\echo 'Done (read-only).'
