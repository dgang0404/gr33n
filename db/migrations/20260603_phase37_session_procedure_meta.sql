-- Phase 37 — store guided-procedure progress on conversation_sessions.
ALTER TABLE gr33ncore.conversation_sessions
    ADD COLUMN IF NOT EXISTS meta JSONB NOT NULL DEFAULT '{}'::jsonb;

COMMENT ON COLUMN gr33ncore.conversation_sessions.meta IS
  'Session-scoped JSON (Phase 37): active_procedure { id, step_n, status } for resumable field walkthroughs.';
