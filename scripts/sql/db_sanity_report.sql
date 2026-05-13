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
\echo '==> Duplicate zone names per farm (breaks master_seed.sql subqueries)'
SELECT farm_id, name, count(*) AS cnt
FROM gr33ncore.zones
GROUP BY farm_id, name
HAVING count(*) > 1
ORDER BY farm_id, name;

\echo ''
\echo '==> RAG chunks row count (informational)'
SELECT count(*) AS rag_embedding_chunks FROM gr33ncore.rag_embedding_chunks;

\echo ''
\echo 'Done (read-only).'
