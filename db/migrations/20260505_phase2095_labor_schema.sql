-- ============================================================
-- Phase 20.95 WS1 — Labor logging schema subset
-- Additive-only: adds tasks.time_spent_minutes (nullable) and
-- a new gr33ncore.task_labor_log table. The time_spent_minutes
-- column is maintained by the labor log handler as a running
-- SUM(task_labor_log.minutes) on every insert/delete — there is
-- no trigger. Phase 20.9 full WS1 will layer auto-cost on top.
-- ============================================================

ALTER TABLE gr33ncore.tasks
    ADD COLUMN IF NOT EXISTS time_spent_minutes INTEGER;

COMMENT ON COLUMN gr33ncore.tasks.time_spent_minutes IS
    'denormalised SUM(task_labor_log.minutes) maintained by handler';

CREATE TABLE IF NOT EXISTS gr33ncore.task_labor_log (
    id                    BIGSERIAL PRIMARY KEY,
    farm_id               BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    task_id               BIGINT NOT NULL REFERENCES gr33ncore.tasks(id) ON DELETE CASCADE,
    user_id               UUID   REFERENCES gr33ncore.profiles(user_id) ON DELETE SET NULL,
    started_at            TIMESTAMPTZ NOT NULL,
    ended_at              TIMESTAMPTZ,
    minutes               INTEGER NOT NULL CHECK (minutes >= 0),
    hourly_rate_snapshot  NUMERIC(10,2),
    currency              CHAR(3) CHECK (currency IS NULL OR currency ~ '^[A-Z]{3}$'),
    notes                 TEXT,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_task_labor_log_task
    ON gr33ncore.task_labor_log (task_id);
CREATE INDEX IF NOT EXISTS idx_task_labor_log_farm
    ON gr33ncore.task_labor_log (farm_id);

DO $$ BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_trigger WHERE tgname = 'trg_task_labor_log_updated_at'
  ) THEN
    CREATE TRIGGER trg_task_labor_log_updated_at
      BEFORE UPDATE ON gr33ncore.task_labor_log
      FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();
  END IF;
END $$;
