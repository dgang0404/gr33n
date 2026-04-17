-- ============================================================
-- Phase 20.9 WS1 — Labor auto-cost default rate
-- Additive-only: adds gr33ncore.profiles.hourly_rate +
-- hourly_rate_currency (both nullable) so the autologger can
-- default a labor log's rate when the operator leaves it blank.
-- task_labor_log.hourly_rate_snapshot / currency were already
-- added in Phase 20.95 WS1 and stay the authoritative historic
-- source (profile rate can change; snapshots must not).
-- ============================================================

ALTER TABLE gr33ncore.profiles
    ADD COLUMN IF NOT EXISTS hourly_rate NUMERIC(10,2) CHECK (hourly_rate IS NULL OR hourly_rate >= 0);

ALTER TABLE gr33ncore.profiles
    ADD COLUMN IF NOT EXISTS hourly_rate_currency CHAR(3)
        CHECK (hourly_rate_currency IS NULL OR hourly_rate_currency ~ '^[A-Z]{3}$');

COMMENT ON COLUMN gr33ncore.profiles.hourly_rate IS
    'Default hourly wage used by the labor autologger when task_labor_log.hourly_rate_snapshot is NULL.';
COMMENT ON COLUMN gr33ncore.profiles.hourly_rate_currency IS
    'ISO-4217 currency paired with hourly_rate; both NULL means no autologged cost will be emitted for this user.';
