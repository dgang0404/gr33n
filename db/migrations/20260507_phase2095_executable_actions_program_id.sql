-- ============================================================
-- Phase 20.95 WS3 — executable_actions.program_id
--
-- Adds program_id BIGINT REFERENCES gr33nfertigation.programs(id) and
-- tightens chk_executable_source from "at-least-one" (schedule_id OR rule_id)
-- to "exactly-one" across (schedule_id, rule_id, program_id).
--
-- Safety: the OLD constraint was at-least-one, which technically allows rows
-- with BOTH schedule_id and rule_id set. Before dropping the old constraint
-- and installing the new num_nonnulls(...) = 1 version, we RAISE EXCEPTION
-- if any such legacy row exists so the migration fails loudly rather than
-- silently rejecting a legitimate row.
--
-- Optional idempotent backfill: programs.metadata->'steps' may describe
-- actuator/notification steps. Since we don't yet have a stable step schema,
-- this migration intentionally does NOT auto-create executable_actions rows.
-- The CRUD handler will grow program_id support in Phase 20.7 WS3.
-- ============================================================

ALTER TABLE gr33ncore.executable_actions
    ADD COLUMN IF NOT EXISTS program_id BIGINT
        REFERENCES gr33nfertigation.programs(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_executable_actions_program
    ON gr33ncore.executable_actions (program_id);

-- Pre-check: abort if any existing row violates exactly-one.
DO $$
DECLARE
    bad_count BIGINT;
BEGIN
    SELECT COUNT(*) INTO bad_count
      FROM gr33ncore.executable_actions
     WHERE num_nonnulls(schedule_id, rule_id, program_id) <> 1;
    IF bad_count > 0 THEN
        RAISE EXCEPTION
          'Phase 20.95 WS3: % executable_actions rows violate exactly-one '
          '(schedule_id, rule_id, program_id). Fix these rows before retrying.',
          bad_count;
    END IF;
END $$;

ALTER TABLE gr33ncore.executable_actions
    DROP CONSTRAINT IF EXISTS chk_executable_source;

ALTER TABLE gr33ncore.executable_actions
    ADD CONSTRAINT chk_executable_source
    CHECK (num_nonnulls(schedule_id, rule_id, program_id) = 1);
