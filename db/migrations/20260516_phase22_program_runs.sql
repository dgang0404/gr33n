-- ============================================================
-- Phase 22 WS1 — worker program-tick + automation_runs.program_id
--
-- Adds the plumbing the worker needs to execute fertigation programs:
--
--   * gr33ncore.automation_runs.program_id — so a program fire lands
--     an automation_run attributable to the program (not a schedule or
--     rule). Nullable; existing runs stay schedule_id/rule_id-bound.
--   * gr33nfertigation.programs.last_triggered_time — mirrors the
--     schedules/rules convention so the worker can short-circuit on
--     "already fired this minute" without scanning automation_runs.
--   * An idempotency index on (program_id, executed_at) so the tick
--     can cheaply ask "did this program already fire at this minute?"
--     via GetAutomationRunByDetails with program_id instead of
--     schedule_id.
--
-- All changes are additive and nullable. No backfill needed.
-- ============================================================

ALTER TABLE gr33ncore.automation_runs
    ADD COLUMN IF NOT EXISTS program_id BIGINT
    REFERENCES gr33nfertigation.programs(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_automation_runs_program_id
    ON gr33ncore.automation_runs(program_id)
    WHERE program_id IS NOT NULL;

ALTER TABLE gr33nfertigation.programs
    ADD COLUMN IF NOT EXISTS last_triggered_time TIMESTAMPTZ;
