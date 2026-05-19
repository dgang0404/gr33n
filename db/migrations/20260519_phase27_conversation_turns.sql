-- ============================================================
-- Phase 27 WS5 follow-up — conversation_turns (Farm Guardian history)
-- ============================================================
-- DB-backed multi-turn history for POST /v1/chat. Each row is one
-- (user_message, assistant_message) pair within a session. Sessions are
-- identified by a server-generated UUID returned to the client in the
-- response (or in the SSE `done` event). Turns within a session are strictly
-- ordered by turn_index starting at 0.
--
-- Trust boundary: same as the rest of gr33ncore — sessions are owned by
-- auth.users (JWT subject). Farm-scoped (grounded) turns also store the
-- farm_id so the handler can verify membership on history reads. Plain
-- (non-grounded) turns leave farm_id NULL.
-- ============================================================

CREATE TABLE IF NOT EXISTS gr33ncore.conversation_turns (
    id                  BIGSERIAL PRIMARY KEY,
    session_id          UUID NOT NULL,
    user_id             UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    farm_id             BIGINT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    turn_index          INTEGER NOT NULL CHECK (turn_index >= 0),
    user_message        TEXT NOT NULL,
    assistant_message   TEXT NOT NULL,
    llm_model           TEXT NOT NULL,
    grounded            BOOLEAN NOT NULL DEFAULT false,
    context_count       INTEGER NOT NULL DEFAULT 0 CHECK (context_count >= 0),
    citations           JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_conversation_turns_session_index UNIQUE (session_id, turn_index)
);

COMMENT ON TABLE gr33ncore.conversation_turns IS
  'Per-session (user_message, assistant_message) history for Farm Guardian (Phase 27 WS5). '
  'Same farm_id trust boundary as gr33ncore.rag_embedding_chunks.';

COMMENT ON COLUMN gr33ncore.conversation_turns.session_id IS
  'Opaque session identifier; generated server-side when the client omits it.';

COMMENT ON COLUMN gr33ncore.conversation_turns.farm_id IS
  'NULL for plain (non-grounded) chats. Set when the turn used RAG retrieval against the farm.';

COMMENT ON COLUMN gr33ncore.conversation_turns.citations IS
  'JSON array of synthesis.Citation entries (ref, chunk_id, source_type, source_id, excerpt); empty for plain turns.';

CREATE INDEX IF NOT EXISTS idx_conversation_turns_user_recent
    ON gr33ncore.conversation_turns (user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_conversation_turns_session
    ON gr33ncore.conversation_turns (session_id, turn_index);

CREATE INDEX IF NOT EXISTS idx_conversation_turns_farm
    ON gr33ncore.conversation_turns (farm_id)
    WHERE farm_id IS NOT NULL;
