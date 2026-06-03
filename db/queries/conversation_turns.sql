-- ============================================================
-- Queries: gr33ncore.conversation_turns + gr33ncore.conversation_sessions
-- ============================================================

-- name: InsertConversationTurn :one
-- Atomically assigns the next turn_index for the session (0-based) and
-- inserts the turn. The UNIQUE (session_id, turn_index) constraint plus
-- COALESCE(MAX+1, 0) keeps numbering monotonic without a sequence per session.
INSERT INTO gr33ncore.conversation_turns (
    session_id,
    user_id,
    farm_id,
    turn_index,
    user_message,
    assistant_message,
    llm_model,
    grounded,
    context_count,
    citations,
    prompt_tokens,
    completion_tokens
) VALUES (
    sqlc.arg(session_id),
    sqlc.arg(user_id),
    sqlc.narg(farm_id),
    COALESCE(
        (SELECT MAX(turn_index) + 1
           FROM gr33ncore.conversation_turns
          WHERE session_id = sqlc.arg(session_id)),
        0
    ),
    sqlc.arg(user_message),
    sqlc.arg(assistant_message),
    sqlc.arg(llm_model),
    sqlc.arg(grounded),
    sqlc.arg(context_count),
    sqlc.arg(citations),
    sqlc.arg(prompt_tokens),
    sqlc.arg(completion_tokens)
)
RETURNING id, session_id, user_id, farm_id, turn_index, created_at;

-- name: ListConversationTurnsBySession :many
SELECT
    id,
    turn_index,
    farm_id,
    user_message,
    assistant_message,
    llm_model,
    grounded,
    context_count,
    citations,
    prompt_tokens,
    completion_tokens,
    created_at
FROM gr33ncore.conversation_turns
WHERE session_id = sqlc.arg(session_id)
  AND user_id    = sqlc.arg(user_id)
ORDER BY turn_index ASC;

-- name: ListRecentConversationSessions :many
-- LEFT JOIN gr33ncore.conversation_sessions so the title surfaces when the
-- operator has renamed the session; otherwise we fall back to the first user
-- message in the API layer.
SELECT
    agg.session_id,
    agg.turn_count::int          AS turn_count,
    agg.last_turn_at              AS last_turn_at,
    agg.any_grounded              AS any_grounded,
    agg.total_prompt_tokens       AS total_prompt_tokens,
    agg.total_completion_tokens   AS total_completion_tokens,
    first_t.user_message          AS first_user_message,
    last_t.assistant_message      AS last_assistant_message,
    last_t.farm_id                AS last_farm_id,
    sess.title                    AS title
FROM (
    SELECT
        session_id,
        COUNT(*)                             AS turn_count,
        MAX(ct.created_at)::timestamptz       AS last_turn_at,
        BOOL_OR(grounded)                    AS any_grounded,
        SUM(prompt_tokens)::int              AS total_prompt_tokens,
        SUM(completion_tokens)::int          AS total_completion_tokens
    FROM gr33ncore.conversation_turns ct
    WHERE ct.user_id = sqlc.arg(user_id)
    GROUP BY ct.session_id
) agg
JOIN LATERAL (
    SELECT user_message
    FROM gr33ncore.conversation_turns
    WHERE session_id = agg.session_id
      AND user_id    = sqlc.arg(user_id)
    ORDER BY turn_index ASC
    LIMIT 1
) first_t ON true
JOIN LATERAL (
    SELECT assistant_message, farm_id
    FROM gr33ncore.conversation_turns
    WHERE session_id = agg.session_id
      AND user_id    = sqlc.arg(user_id)
    ORDER BY turn_index DESC
    LIMIT 1
) last_t ON true
LEFT JOIN gr33ncore.conversation_sessions sess
    ON sess.id = agg.session_id AND sess.user_id = sqlc.arg(user_id)
ORDER BY agg.last_turn_at DESC
LIMIT sqlc.arg(match_limit);

-- name: UpsertConversationSession :exec
-- Called on every turn insert so the session row exists and updated_at
-- tracks the latest turn for ordering. The set_updated_at trigger handles
-- updated_at on the conflict branch.
INSERT INTO gr33ncore.conversation_sessions (id, user_id)
VALUES (sqlc.arg(id), sqlc.arg(user_id))
ON CONFLICT (id) DO UPDATE
    SET updated_at = NOW();

-- name: GetConversationSessionMeta :one
SELECT meta
FROM gr33ncore.conversation_sessions
WHERE id = sqlc.arg(id)
  AND user_id = sqlc.arg(user_id);

-- name: UpdateConversationSessionMeta :exec
UPDATE gr33ncore.conversation_sessions
SET meta = sqlc.arg(meta)
WHERE id = sqlc.arg(id)
  AND user_id = sqlc.arg(user_id);

-- name: UpdateConversationSessionTitle :one
-- Operator rename. Returns the row so the API can confirm ownership in one
-- query — RowsAffected = 0 means "not found / not yours" → 404.
UPDATE gr33ncore.conversation_sessions
SET title = sqlc.narg(title)
WHERE id = sqlc.arg(id)
  AND user_id = sqlc.arg(user_id)
RETURNING id, title, updated_at;

-- name: DeleteConversationTurnsBySession :exec
DELETE FROM gr33ncore.conversation_turns
WHERE session_id = sqlc.arg(session_id)
  AND user_id    = sqlc.arg(user_id);

-- name: DeleteConversationSession :execrows
-- Drops the metadata row last. Caller deletes turns first so the user-facing
-- session disappears cleanly even when no metadata row was ever created
-- (sessions auto-created via UpsertConversationSession backfilled rows for
-- pre-existing turn data).
DELETE FROM gr33ncore.conversation_sessions
WHERE id = sqlc.arg(id)
  AND user_id = sqlc.arg(user_id);

-- name: DeleteStaleConversationTurns :execrows
-- Phase 27 WS5 follow-up — TTL pruning. Removes turns belonging to sessions
-- whose latest activity (MAX(created_at) per session) is older than the
-- cutoff. Cutoff is computed in Go (NOW() - ttl) so the SQL is parameter-only
-- and portable across timezones.
DELETE FROM gr33ncore.conversation_turns
WHERE session_id IN (
    SELECT ct.session_id
    FROM gr33ncore.conversation_turns ct
    GROUP BY ct.session_id
    HAVING MAX(ct.created_at) < sqlc.arg(cutoff)
);

-- name: DeleteStaleConversationSessions :execrows
-- Phase 27 WS5 follow-up — TTL pruning. Drops session metadata rows whose
-- last activity (updated_at) is older than the cutoff. Run AFTER
-- DeleteStaleConversationTurns so the visible-to-the-API row count flips to
-- zero in lockstep with the turn data.
DELETE FROM gr33ncore.conversation_sessions
WHERE updated_at < sqlc.arg(cutoff);

-- name: SumChatTokensSinceForUser :one
-- Phase 27 WS5 follow-up — cost guards. Rolling-window token total for a
-- single user across every session they own. `since` is the window start
-- computed in Go (NOW() - window) so the SQL is parameter-only and the
-- caller decides on the window length.
SELECT
    COALESCE(SUM(prompt_tokens), 0)::bigint     AS prompt_tokens,
    COALESCE(SUM(completion_tokens), 0)::bigint AS completion_tokens,
    COALESCE(SUM(prompt_tokens + completion_tokens), 0)::bigint AS total_tokens
FROM gr33ncore.conversation_turns
WHERE user_id = sqlc.arg(user_id)
  AND created_at >= sqlc.arg(since);

-- name: SumChatTokensSinceForFarm :one
-- Phase 27 WS5 follow-up — cost guards. Rolling-window token total for a
-- single farm across every user who chatted with that farm's data. Plain
-- (non-grounded) turns have farm_id IS NULL and are excluded.
SELECT
    COALESCE(SUM(prompt_tokens), 0)::bigint     AS prompt_tokens,
    COALESCE(SUM(completion_tokens), 0)::bigint AS completion_tokens,
    COALESCE(SUM(prompt_tokens + completion_tokens), 0)::bigint AS total_tokens
FROM gr33ncore.conversation_turns
WHERE farm_id = sqlc.arg(farm_id)
  AND created_at >= sqlc.arg(since);
