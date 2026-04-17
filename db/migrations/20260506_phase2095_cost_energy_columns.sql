-- ============================================================
-- Phase 20.95 WS2 — Cost / energy column adds (additive-only)
-- ============================================================
-- Adds optional unit-cost metadata to input_definitions, a
-- low-stock marker on input_batches, a wattage column on actuators
-- (DEFAULT 0 — safe for existing rows), a new farm_energy_prices
-- table for per-farm $/kWh pricing, a crop_cycle_id column on
-- cost_transactions so Phase 21 can report $ per cycle, and
-- broadens input_category_enum for livestock categories so
-- animal feed / bedding / veterinary supply purchases cost correctly.
-- ============================================================

-- input cost metadata
ALTER TABLE gr33nnaturalfarming.input_definitions
    ADD COLUMN IF NOT EXISTS unit_cost          NUMERIC(12,4);
ALTER TABLE gr33nnaturalfarming.input_definitions
    ADD COLUMN IF NOT EXISTS unit_cost_currency CHAR(3)
        CHECK (unit_cost_currency IS NULL OR unit_cost_currency ~ '^[A-Z]{3}$');
ALTER TABLE gr33nnaturalfarming.input_definitions
    ADD COLUMN IF NOT EXISTS unit_cost_unit_id  BIGINT REFERENCES gr33ncore.units(id) ON DELETE SET NULL;

-- low-stock trigger marker
ALTER TABLE gr33nnaturalfarming.input_batches
    ADD COLUMN IF NOT EXISTS low_stock_threshold NUMERIC(12,4);

-- actuator wattage for the nightly electricity rollup (Phase 20.7 WS4, later)
ALTER TABLE gr33ncore.actuators
    ADD COLUMN IF NOT EXISTS watts NUMERIC(10,2) DEFAULT 0 NOT NULL;

-- farm-level energy pricing (additive new table)
CREATE TABLE IF NOT EXISTS gr33ncore.farm_energy_prices (
    id               BIGSERIAL PRIMARY KEY,
    farm_id          BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    effective_from   DATE   NOT NULL,
    effective_to     DATE,
    price_per_kwh    NUMERIC(10,4) NOT NULL CHECK (price_per_kwh >= 0),
    currency         CHAR(3)       NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    notes            TEXT,
    created_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_farm_energy_prices_active
    ON gr33ncore.farm_energy_prices (farm_id, effective_from DESC);
DO $$ BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_trigger WHERE tgname = 'trg_farm_energy_prices_updated_at'
  ) THEN
    CREATE TRIGGER trg_farm_energy_prices_updated_at
      BEFORE UPDATE ON gr33ncore.farm_energy_prices
      FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();
  END IF;
END $$;

-- cycle-scoped cost tagging (Phase 21 "$ per cycle" report)
ALTER TABLE gr33ncore.cost_transactions
    ADD COLUMN IF NOT EXISTS crop_cycle_id BIGINT
        REFERENCES gr33nfertigation.crop_cycles(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_cost_tx_crop_cycle
    ON gr33ncore.cost_transactions (crop_cycle_id)
    WHERE crop_cycle_id IS NOT NULL;

-- broaden input category so animal feed etc. cost correctly (Phase 20.8)
ALTER TYPE gr33nnaturalfarming.input_category_enum ADD VALUE IF NOT EXISTS 'animal_feed';
ALTER TYPE gr33nnaturalfarming.input_category_enum ADD VALUE IF NOT EXISTS 'bedding';
ALTER TYPE gr33nnaturalfarming.input_category_enum ADD VALUE IF NOT EXISTS 'veterinary_supply';
