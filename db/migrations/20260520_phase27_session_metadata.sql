-- ============================================================
-- Phase 27 WS5 follow-up — session metadata + token-usage accounting
-- ============================================================
-- Adds a side table for per-session metadata that does not belong on every
-- turn (title, soft delete marker, monotonic updated_at). The existing
-- conversation_turns table still owns the (user_message, assistant_message)
-- history; sessions just keeps mutable per-session fields the operator can
-- edit (rename, archive) and the API can read cheaply without touching turn
-- rows.
--
-- Also adds prompt_tokens / completion_tokens columns to conversation_turns
-- so the UI can render per-turn token usage and the API can later enforce
-- cost guards on it.
-- ============================================================

CREATE TABLE IF NOT EXISTS gr33ncore.conversation_sessions (
    id          UUID PRIMARY KEY,
    user_id     UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    title       TEXT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE gr33ncore.conversation_sessions IS
  'Per-session metadata for Farm Guardian (Phase 27 WS5 follow-up). '
  'session.id matches conversation_turns.session_id one-to-many.';

COMMENT ON COLUMN gr33ncore.conversation_sessions.title IS
  'Operator-supplied label; NULL means the UI falls back to the first user message.';

CREATE INDEX IF NOT EXISTS idx_conversation_sessions_user_updated
    ON gr33ncore.conversation_sessions (user_id, updated_at DESC);

DROP TRIGGER IF EXISTS trg_conversation_sessions_updated_at ON gr33ncore.conversation_sessions;
CREATE TRIGGER trg_conversation_sessions_updated_at
    BEFORE UPDATE ON gr33ncore.conversation_sessions
    FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();

ALTER TABLE gr33ncore.conversation_turns
    ADD COLUMN IF NOT EXISTS prompt_tokens     INTEGER NOT NULL DEFAULT 0 CHECK (prompt_tokens     >= 0),
    ADD COLUMN IF NOT EXISTS completion_tokens INTEGER NOT NULL DEFAULT 0 CHECK (completion_tokens >= 0);

COMMENT ON COLUMN gr33ncore.conversation_turns.prompt_tokens IS
  'Tokens billed for the request (system + history + user). 0 when the LLM did not report usage (e.g. streaming on a backend without stream_options.include_usage).';

COMMENT ON COLUMN gr33ncore.conversation_turns.completion_tokens IS
  'Tokens billed for the assistant reply.';
