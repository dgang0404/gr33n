-- Phase 138 — per-farm counsel/quick model policy and grounded timeout override.

ALTER TABLE gr33ncore.farms
  ADD COLUMN IF NOT EXISTS guardian_counsel_model TEXT NULL,
  ADD COLUMN IF NOT EXISTS guardian_quick_model TEXT NULL,
  ADD COLUMN IF NOT EXISTS guardian_grounded_timeout_seconds INTEGER NULL
    CHECK (guardian_grounded_timeout_seconds IS NULL OR guardian_grounded_timeout_seconds > 0);

COMMENT ON COLUMN gr33ncore.farms.guardian_counsel_model IS
  'Farm counsel (grounded) chat model — Phase 138. Falls back to guardian_preferred_model when unset.';

COMMENT ON COLUMN gr33ncore.farms.guardian_quick_model IS
  'Quick chat model for this farm when farm context is off — Phase 138.';

COMMENT ON COLUMN gr33ncore.farms.guardian_grounded_timeout_seconds IS
  'Per-farm grounded chat HTTP timeout override (seconds). NULL uses GUARDIAN_GROUNDED_TIMEOUT_SECONDS / env default.';

UPDATE gr33ncore.farms
SET guardian_counsel_model = guardian_preferred_model
WHERE guardian_counsel_model IS NULL
  AND guardian_preferred_model IS NOT NULL
  AND TRIM(guardian_preferred_model) <> '';
