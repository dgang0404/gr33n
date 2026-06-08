-- ============================================================
-- Queries: gr33ncore.session_summaries (Phase 63)
-- ============================================================

-- name: UpsertSessionSummary :one
INSERT INTO gr33ncore.session_summaries (
    session_id,
    farm_id,
    user_id,
    summary_text,
    topics
) VALUES (
    sqlc.arg(session_id),
    sqlc.arg(farm_id),
    sqlc.arg(user_id),
    sqlc.arg(summary_text),
    sqlc.arg(topics)
)
ON CONFLICT (session_id) DO UPDATE
    SET summary_text = EXCLUDED.summary_text,
        topics       = EXCLUDED.topics,
        created_at   = NOW()
RETURNING session_id, farm_id, user_id, summary_text, topics, created_at;

-- name: GetSessionSummary :one
SELECT session_id, farm_id, user_id, summary_text, topics, created_at
FROM gr33ncore.session_summaries
WHERE session_id = sqlc.arg(session_id)
  AND user_id    = sqlc.arg(user_id);

-- name: ListSessionSummariesByFarmUser :many
SELECT session_id, farm_id, user_id, summary_text, topics, created_at
FROM gr33ncore.session_summaries
WHERE farm_id = sqlc.arg(farm_id)
  AND user_id = sqlc.arg(user_id)
ORDER BY created_at DESC
LIMIT sqlc.arg(match_limit);

-- name: ListSessionSummaryTopicsForUser :many
SELECT session_id, topics
FROM gr33ncore.session_summaries
WHERE user_id = sqlc.arg(user_id);

-- name: FindMatchingSessionSummary :one
-- Returns the newest summary whose topics overlap any of the requested tags.
SELECT session_id, farm_id, user_id, summary_text, topics, created_at
FROM gr33ncore.session_summaries
WHERE farm_id = sqlc.arg(farm_id)
  AND user_id = sqlc.arg(user_id)
  AND topics && sqlc.arg(topics)::text[]
ORDER BY created_at DESC
LIMIT 1;

-- name: DeleteSessionSummary :exec
DELETE FROM gr33ncore.session_summaries
WHERE session_id = sqlc.arg(session_id)
  AND user_id    = sqlc.arg(user_id);

-- name: DeleteAllSessionSummariesForFarmUser :execrows
DELETE FROM gr33ncore.session_summaries
WHERE farm_id = sqlc.arg(farm_id)
  AND user_id = sqlc.arg(user_id);
