-- ============================================================
-- Insert Commons receiver (ingest persistence)
-- ============================================================

-- name: InsertInsertCommonsReceivedPayload :one
INSERT INTO gr33ncore.insert_commons_received_payloads (
    payload_hash,
    farm_pseudonym,
    schema_version,
    generated_at,
    payload,
    source_idempotency_key
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING id;

-- name: GetInsertCommonsReceivedPayloadIDByHash :one
SELECT id
FROM gr33ncore.insert_commons_received_payloads
WHERE payload_hash = $1;

-- name: GetInsertCommonsReceivedPayloadByFarmIdempotency :one
SELECT id, payload_hash
FROM gr33ncore.insert_commons_received_payloads
WHERE farm_pseudonym = $1 AND source_idempotency_key = $2;

-- name: DeleteInsertCommonsReceivedPayloadsBefore :exec
DELETE FROM gr33ncore.insert_commons_received_payloads
WHERE received_at < $1;

-- name: InsertCommonsReceiverStats :one
SELECT
    COUNT(*)::bigint AS total_payloads,
    COUNT(DISTINCT farm_pseudonym)::bigint AS distinct_pseudonyms,
    MIN(received_at) AS oldest_received_at,
    MAX(received_at) AS newest_received_at
FROM gr33ncore.insert_commons_received_payloads;

-- name: InsertCommonsReceiverDailyCounts :many
SELECT (timezone('UTC', received_at))::date AS day,
    COUNT(*)::bigint AS ingest_count
FROM gr33ncore.insert_commons_received_payloads
WHERE received_at >= $1
GROUP BY 1
ORDER BY 1 DESC;
