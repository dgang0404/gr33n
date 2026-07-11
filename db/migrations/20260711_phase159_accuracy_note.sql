-- Phase 159 — persist Guardian accuracy_note on conversation turns.
ALTER TABLE gr33ncore.conversation_turns
    ADD COLUMN IF NOT EXISTS accuracy_note TEXT NULL;

COMMENT ON COLUMN gr33ncore.conversation_turns.accuracy_note IS
    'Phase 159 — live AnswerAccuracyNote code from chat finalize; shown as farmer-facing banner on reload.';
