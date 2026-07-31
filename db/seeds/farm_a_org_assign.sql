-- Phase 212 — Install A: Organization A + rename farm 1 → Farm A.
-- One-shot against Install A DATABASE_URL. Idempotent.
--
--   psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f db/seeds/farm_a_org_assign.sql

BEGIN;

INSERT INTO gr33ncore.organizations (name, plan_tier, billing_status)
SELECT 'Organization A', 'pilot', 'none'
WHERE NOT EXISTS (
    SELECT 1 FROM gr33ncore.organizations WHERE name = 'Organization A'
);

INSERT INTO gr33ncore.organization_memberships (organization_id, user_id, role_in_org)
SELECT o.id, '00000000-0000-0000-0000-000000000001'::uuid, 'owner'
FROM gr33ncore.organizations o
WHERE o.name = 'Organization A'
ON CONFLICT (organization_id, user_id) DO NOTHING;

UPDATE gr33ncore.farms f
SET
    name = 'Farm A',
    description = 'Phase 212 Install A federation test farm (Organization A).',
    timezone = 'America/New_York',
    currency = 'USD',
    organization_id = o.id,
    meta_data = COALESCE(f.meta_data, '{}'::jsonb) || '{"phase212":"install_a"}'::jsonb
FROM gr33ncore.organizations o
WHERE f.id = 1
  AND o.name = 'Organization A';

COMMIT;
