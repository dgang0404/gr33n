-- Phase 20 / WS1 — Rule → Task linkage.
--
-- When a rule's create_task action fires, the resulting row in
-- gr33ncore.tasks is tagged with the originating rule via source_rule_id.
-- Mirrors the Phase 19 source_alert_id pattern so analytics can segment
-- "tasks born from rules" vs "tasks born from alerts" vs "manually created".
--
-- ON DELETE SET NULL so retiring a rule doesn't destroy task history.

ALTER TABLE gr33ncore.tasks
    ADD COLUMN IF NOT EXISTS source_rule_id BIGINT
        REFERENCES gr33ncore.automation_rules(id) ON DELETE SET NULL;

-- Partial index — only rule-born tasks participate.
CREATE INDEX IF NOT EXISTS idx_tasks_source_rule_id
    ON gr33ncore.tasks (source_rule_id)
    WHERE source_rule_id IS NOT NULL;
