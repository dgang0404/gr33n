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
\echo 'Done (read-only).'
