-- Phase 12: Farm-configurable finance COA mapping overrides for GL exports

CREATE TABLE IF NOT EXISTS gr33ncore.farm_finance_account_mappings (
    id            BIGSERIAL PRIMARY KEY,
    farm_id       BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    cost_category gr33ncore.cost_category_enum NOT NULL,
    account_code  TEXT   NOT NULL,
    account_name  TEXT   NOT NULL,
    is_active     BOOLEAN NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at    TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    UNIQUE (farm_id, cost_category)
);

DO $$ BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_trigger
    WHERE tgname = 'trg_farm_finance_account_mappings_updated_at'
  ) THEN
    CREATE TRIGGER trg_farm_finance_account_mappings_updated_at
      BEFORE UPDATE ON gr33ncore.farm_finance_account_mappings
      FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();
  END IF;
END $$;
