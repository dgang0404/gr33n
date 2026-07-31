-- Phase 212 — Install B distinguishable tenant (Organization B + Farm B).
-- Apply on Install B AFTER master_seed (or instead of renaming demo farm).
-- Idempotent: safe to re-run.
--
--   psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f db/seeds/farm_b_seed.sql

BEGIN;

-- Ensure demo user exists (no-op if master_seed already ran).
INSERT INTO auth.users (id, email)
VALUES ('00000000-0000-0000-0000-000000000001', 'dev@gr33n.local')
ON CONFLICT (id) DO NOTHING;

INSERT INTO gr33ncore.profiles (user_id, full_name, email, role)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'Dev Farmer', 'dev@gr33n.local', 'farm_manager'
) ON CONFLICT (user_id) DO NOTHING;

-- Organization B (distinct from Install A's Organization A / smoke orgs).
INSERT INTO gr33ncore.organizations (name, plan_tier, billing_status)
SELECT 'Organization B', 'pilot', 'none'
WHERE NOT EXISTS (
    SELECT 1 FROM gr33ncore.organizations WHERE name = 'Organization B'
);

INSERT INTO gr33ncore.organization_memberships (organization_id, user_id, role_in_org)
SELECT o.id, '00000000-0000-0000-0000-000000000001'::uuid, 'owner'
FROM gr33ncore.organizations o
WHERE o.name = 'Organization B'
ON CONFLICT (organization_id, user_id) DO NOTHING;

-- Rename demo farm 1 → Farm B and attach Organization B.
-- Different timezone/currency from Install A Farm A (America/New_York + USD).
UPDATE gr33ncore.farms f
SET
    name = 'Farm B',
    description = 'Phase 212 Install B federation test farm (Organization B).',
    timezone = 'America/Chicago',
    currency = 'CAD',
    organization_id = o.id,
    meta_data = COALESCE(f.meta_data, '{}'::jsonb) || '{"phase212":"install_b","dev_seed_profile":"demo_showcase"}'::jsonb
FROM gr33ncore.organizations o
WHERE f.id = 1
  AND o.name = 'Organization B';

COMMIT;
