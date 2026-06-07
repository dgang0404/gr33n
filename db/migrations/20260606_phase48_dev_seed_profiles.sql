-- Phase 48 — dev seed profiles: farm meta_data, sensor name idempotency, profile helpers.

ALTER TABLE gr33ncore.farms
    ADD COLUMN IF NOT EXISTS meta_data JSONB NOT NULL DEFAULT '{}'::jsonb;

COMMENT ON COLUMN gr33ncore.farms.meta_data IS
    'Operator metadata; dev_seed_profile (small_indoor | demo_showcase) for local/staging hygiene (Phase 48).';

-- Soft-delete duplicate active sensors before partial unique index (keep lowest id).
WITH ranked AS (
    SELECT id,
           ROW_NUMBER() OVER (PARTITION BY farm_id, name ORDER BY id) AS rn
    FROM gr33ncore.sensors
    WHERE deleted_at IS NULL
)
UPDATE gr33ncore.sensors s
SET deleted_at = NOW(),
    updated_at = NOW()
FROM ranked r
WHERE s.id = r.id
  AND r.rn > 1;

CREATE UNIQUE INDEX IF NOT EXISTS uq_sensors_farm_name_active
    ON gr33ncore.sensors (farm_id, name)
    WHERE deleted_at IS NULL;

CREATE OR REPLACE FUNCTION gr33ncore.set_dev_seed_profile(p_farm_id BIGINT, p_profile TEXT)
RETURNS void
LANGUAGE plpgsql
AS $$
BEGIN
    IF p_profile IS NULL OR btrim(p_profile) = '' THEN
        RAISE EXCEPTION 'dev_seed_profile required';
    END IF;
    UPDATE gr33ncore.farms
    SET meta_data = COALESCE(meta_data, '{}'::jsonb)
        || jsonb_build_object('dev_seed_profile', btrim(p_profile)),
        updated_at = NOW()
    WHERE id = p_farm_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'farm % not found', p_farm_id;
    END IF;
END;
$$;

COMMENT ON FUNCTION gr33ncore.set_dev_seed_profile(BIGINT, TEXT) IS
    'Stamp farms.meta_data.dev_seed_profile (Phase 48 WS1/WS4).';
