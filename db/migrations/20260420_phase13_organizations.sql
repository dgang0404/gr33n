-- Phase 13 WS5: Optional organization grouping for multi-farm tenants + billing hooks (reversible).

CREATE TABLE IF NOT EXISTS gr33ncore.organizations (
    id              BIGSERIAL PRIMARY KEY,
    name            TEXT        NOT NULL,
    plan_tier       TEXT        NOT NULL DEFAULT 'pilot',
    billing_status  TEXT        NOT NULL DEFAULT 'none',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

DO $$ BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_trigger WHERE tgname = 'trg_organizations_updated_at'
  ) THEN
    CREATE TRIGGER trg_organizations_updated_at
      BEFORE UPDATE ON gr33ncore.organizations
      FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();
  END IF;
END $$;

CREATE TABLE IF NOT EXISTS gr33ncore.organization_memberships (
    organization_id BIGINT NOT NULL REFERENCES gr33ncore.organizations(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES gr33ncore.profiles(user_id) ON DELETE CASCADE,
    role_in_org     TEXT   NOT NULL CHECK (role_in_org IN ('owner', 'admin', 'member')),
    joined_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (organization_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_org_memberships_user
    ON gr33ncore.organization_memberships (user_id);

ALTER TABLE gr33ncore.farms ADD COLUMN IF NOT EXISTS organization_id BIGINT
    REFERENCES gr33ncore.organizations(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_farms_organization_id
    ON gr33ncore.farms (organization_id)
    WHERE deleted_at IS NULL AND organization_id IS NOT NULL;
