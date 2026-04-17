-- Phase 19 / WS4 — Schedule preconditions (interlock lite).
--
-- Before executing a schedule's executable_actions, the worker consults
-- this JSON list. Each entry is an evaluation against the latest reading
-- for a sensor on the same farm. If any fails, the worker records an
-- automation_runs row with status='skipped' and does NOT touch actuators.
--
-- Shape: [{ "sensor_id": 12, "op": "gte", "value": 10.0 }, ...]
-- Supported ops: lt | lte | eq | gte | gt | ne
-- Empty list ([]) means no interlock — same behavior as today.

ALTER TABLE gr33ncore.schedules
    ADD COLUMN IF NOT EXISTS preconditions JSONB NOT NULL DEFAULT '[]'::jsonb;

DO $$ BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'chk_schedule_preconditions_is_array'
    ) THEN
        ALTER TABLE gr33ncore.schedules
            ADD CONSTRAINT chk_schedule_preconditions_is_array
            CHECK (jsonb_typeof(preconditions) = 'array');
    END IF;
END $$;
