-- ============================================================
-- Queries: gr33ncore.conversation_turns (Phase 27 WS5 follow-up)
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
    citations
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
    sqlc.arg(citations)
)
RETURNING id, session_id, user_id, farm_id, turn_index, created_at;

-- name: ListConversationTurnsBySession :many
-- Ordered history for a session, scoped to the calling user so a session_id
-- guess can't leak another operator's chat.
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
    created_at
FROM gr33ncore.conversation_turns
WHERE session_id = sqlc.arg(session_id)
  AND user_id    = sqlc.arg(user_id)
ORDER BY turn_index ASC;

-- name: ListRecentConversationSessions :many
-- One row per distinct session_id for the calling user, most recently active first.
-- LATERAL joins keep first/last messages typed (vs ARRAY_AGG which sqlc renders as interface{}).
SELECT
    agg.session_id,
    agg.turn_count::int          AS turn_count,
    agg.last_turn_at              AS last_turn_at,
    agg.any_grounded              AS any_grounded,
    first_t.user_message          AS first_user_message,
    last_t.assistant_message      AS last_assistant_message,
    last_t.farm_id                AS last_farm_id
FROM (
    SELECT
        session_id,
        COUNT(*)        AS turn_count,
        MAX(created_at) AS last_turn_at,
        BOOL_OR(grounded) AS any_grounded
    FROM gr33ncore.conversation_turns
    WHERE user_id = sqlc.arg(user_id)
    GROUP BY session_id
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
ORDER BY agg.last_turn_at DESC
LIMIT sqlc.arg(match_limit);
