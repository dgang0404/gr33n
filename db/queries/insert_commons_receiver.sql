-- ============================================================
-- Insert Commons receiver (ingest persistence)
-- ============================================================

-- name: InsertInsertCommonsReceivedPayload :one
INSERT INTO gr33ncore.insert_commons_received_payloads (
    payload_hash,
    farm_pseudonym,
    schema_version,
    generated_at,
    payload
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING id;

-- name: GetInsertCommonsReceivedPayloadIDByHash :one
SELECT id
FROM gr33ncore.insert_commons_received_payloads
WHERE payload_hash = $1;

-- name: DeleteInsertCommonsReceivedPayloadsBefore :exec
DELETE FROM gr33ncore.insert_commons_received_payloads
WHERE received_at < $1;
