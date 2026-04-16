-- Phase 14 WS7 (stretch): starter schemas for future domain modules — no product APIs yet.
-- Enable per farm via gr33ncore.farm_active_modules (module_schema_name = schema name below).
-- Rollback (dev only): DROP SCHEMA gr33naquaponics CASCADE; DROP SCHEMA gr33nanimals CASCADE; DROP SCHEMA gr33ncrops CASCADE;

CREATE SCHEMA IF NOT EXISTS gr33ncrops;

CREATE TABLE IF NOT EXISTS gr33ncrops.plants (
    id                   BIGSERIAL PRIMARY KEY,
    farm_id              BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    display_name         TEXT NOT NULL,
    variety_or_cultivar  TEXT,
    meta                 JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at           TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_gr33ncrops_plants_farm
    ON gr33ncrops.plants (farm_id)
    WHERE deleted_at IS NULL;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_gr33ncrops_plants_updated_at') THEN
    CREATE TRIGGER trg_gr33ncrops_plants_updated_at
      BEFORE UPDATE ON gr33ncrops.plants
      FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();
  END IF;
END $$;

CREATE SCHEMA IF NOT EXISTS gr33nanimals;

CREATE TABLE IF NOT EXISTS gr33nanimals.animal_groups (
    id          BIGSERIAL PRIMARY KEY,
    farm_id     BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    label       TEXT NOT NULL,
    species     TEXT,
    meta        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_gr33nanimals_groups_farm
    ON gr33nanimals.animal_groups (farm_id)
    WHERE deleted_at IS NULL;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_gr33nanimals_animal_groups_updated_at') THEN
    CREATE TRIGGER trg_gr33nanimals_animal_groups_updated_at
      BEFORE UPDATE ON gr33nanimals.animal_groups
      FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();
  END IF;
END $$;

CREATE SCHEMA IF NOT EXISTS gr33naquaponics;

CREATE TABLE IF NOT EXISTS gr33naquaponics.loops (
    id          BIGSERIAL PRIMARY KEY,
    farm_id     BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    label       TEXT NOT NULL,
    meta        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_gr33naquaponics_loops_farm
    ON gr33naquaponics.loops (farm_id)
    WHERE deleted_at IS NULL;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_gr33naquaponics_loops_updated_at') THEN
    CREATE TRIGGER trg_gr33naquaponics_loops_updated_at
      BEFORE UPDATE ON gr33naquaponics.loops
      FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();
  END IF;
END $$;
