-- Phase 19 / WS3 — Alert → Task linkage.
--
-- Lets an operator turn a specific alert into a tracked task with a
-- persistent back-reference. ON DELETE SET NULL so an expired alert can
-- be cleaned up without destroying task history.

ALTER TABLE gr33ncore.tasks
    ADD COLUMN IF NOT EXISTS source_alert_id BIGINT
        REFERENCES gr33ncore.alerts_notifications(id) ON DELETE SET NULL;

-- Partial index — only tasks that came from an alert participate.
CREATE INDEX IF NOT EXISTS idx_tasks_source_alert_id
    ON gr33ncore.tasks (source_alert_id)
    WHERE source_alert_id IS NOT NULL;
