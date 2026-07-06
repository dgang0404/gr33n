-- Phase 134 — Guardian answer feedback on conversation turns.

ALTER TABLE gr33ncore.conversation_turns
    ADD COLUMN IF NOT EXISTS feedback_rating text
        CHECK (feedback_rating IS NULL OR feedback_rating IN ('up', 'down')),
    ADD COLUMN IF NOT EXISTS feedback_reason text,
    ADD COLUMN IF NOT EXISTS feedback_at timestamptz;

CREATE INDEX IF NOT EXISTS idx_conversation_turns_farm_feedback
    ON gr33ncore.conversation_turns (farm_id, feedback_at DESC)
    WHERE farm_id IS NOT NULL AND feedback_rating IS NOT NULL;

COMMENT ON COLUMN gr33ncore.conversation_turns.feedback_rating IS
    'Phase 134 — operator thumbs up/down on assistant answer';
COMMENT ON COLUMN gr33ncore.conversation_turns.feedback_reason IS
    'Phase 134 — optional free-text reason (especially on down)';
