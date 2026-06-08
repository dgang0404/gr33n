-- Phase 63 — Guardian session memory (farm-scoped, operator-visible summaries).
CREATE TABLE IF NOT EXISTS gr33ncore.session_summaries (
    session_id    UUID PRIMARY KEY REFERENCES gr33ncore.conversation_sessions(id) ON DELETE CASCADE,
    farm_id       BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    user_id       UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    summary_text  TEXT NOT NULL,
    topics        TEXT[] NOT NULL DEFAULT '{}',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE gr33ncore.session_summaries IS
  'Phase 63 — 2–3 sentence Guardian session recap with topic tags for cross-session context injection.';

CREATE INDEX IF NOT EXISTS idx_session_summaries_farm_user_created
    ON gr33ncore.session_summaries (farm_id, user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_session_summaries_user_created
    ON gr33ncore.session_summaries (user_id, created_at DESC);
