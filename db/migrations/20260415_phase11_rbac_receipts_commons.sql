-- Phase 11: RBAC enum values, Insert Commons farm flags (idempotent-ish for dev DBs)

DO $$ BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_enum e
    JOIN pg_type t ON e.enumtypid = t.oid
    JOIN pg_namespace n ON n.oid = t.typnamespace
    WHERE n.nspname = 'gr33ncore' AND t.typname = 'farm_member_role_enum' AND e.enumlabel = 'operator'
  ) THEN
    ALTER TYPE gr33ncore.farm_member_role_enum ADD VALUE 'operator';
  END IF;
END $$;
DO $$ BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_enum e
    JOIN pg_type t ON e.enumtypid = t.oid
    JOIN pg_namespace n ON n.oid = t.typnamespace
    WHERE n.nspname = 'gr33ncore' AND t.typname = 'farm_member_role_enum' AND e.enumlabel = 'finance'
  ) THEN
    ALTER TYPE gr33ncore.farm_member_role_enum ADD VALUE 'finance';
  END IF;
END $$;

ALTER TABLE gr33ncore.farms
    ADD COLUMN IF NOT EXISTS insert_commons_opt_in BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE gr33ncore.farms
    ADD COLUMN IF NOT EXISTS insert_commons_last_sync_at TIMESTAMPTZ;
