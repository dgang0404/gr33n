-- Phase 56 WS1 — link crop cycles to Plants catalog.
-- Phase 56 WS2 — stage transition history.

ALTER TABLE gr33nfertigation.crop_cycles
    ADD COLUMN IF NOT EXISTS plant_id BIGINT REFERENCES gr33ncrops.plants(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_crop_cycles_plant_id
    ON gr33nfertigation.crop_cycles (plant_id)
    WHERE plant_id IS NOT NULL;

COMMENT ON COLUMN gr33nfertigation.crop_cycles.plant_id IS
    'Optional FK to gr33ncrops.plants — strain catalog link (Phase 56).';

-- Best-effort backfill from strain_or_variety text → plant display_name.
UPDATE gr33nfertigation.crop_cycles c
SET plant_id = p.id
FROM gr33ncrops.plants p
WHERE c.plant_id IS NULL
  AND c.farm_id = p.farm_id
  AND c.strain_or_variety IS NOT NULL
  AND (
    lower(trim(c.strain_or_variety)) = lower(trim(p.display_name))
    OR lower(trim(c.strain_or_variety)) LIKE lower(trim(p.display_name)) || ' (%'
  );

CREATE TABLE IF NOT EXISTS gr33nfertigation.crop_cycle_stage_events (
    id            BIGSERIAL PRIMARY KEY,
    crop_cycle_id BIGINT NOT NULL REFERENCES gr33nfertigation.crop_cycles(id) ON DELETE CASCADE,
    growth_stage  gr33nfertigation.growth_stage_enum NOT NULL,
    entered_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_crop_cycle_stage_events_cycle
    ON gr33nfertigation.crop_cycle_stage_events (crop_cycle_id, entered_at ASC);

COMMENT ON TABLE gr33nfertigation.crop_cycle_stage_events IS
    'Stage transitions for farmer-friendly grow timelines (Phase 56 WS2).';

-- Seed one event per existing cycle from current_stage + started_at.
INSERT INTO gr33nfertigation.crop_cycle_stage_events (crop_cycle_id, growth_stage, entered_at)
SELECT c.id, c.current_stage, COALESCE(c.started_at::timestamptz, c.created_at)
FROM gr33nfertigation.crop_cycles c
WHERE c.current_stage IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM gr33nfertigation.crop_cycle_stage_events e WHERE e.crop_cycle_id = c.id
  );
