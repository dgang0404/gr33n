-- Dev hygiene — strip Go smoke-test pollution from farm 1 (+ smoke auth/orgs).
-- Safe to re-run. Does NOT wipe Docker volumes. Keeps dev@gr33n.local + farm 1.
--
-- Run standalone:
--   psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f scripts/sql/dev_purge_smoke_pollution.sql
-- Or via: ./scripts/dev-reset-farm.sh --farm-id 1 --profile demo_showcase

BEGIN;

-- Farm 1: automation run log (smoke rules tick thousands of times per test run).
DELETE FROM gr33ncore.automation_runs WHERE farm_id = 1;

-- Farm 1: alert inbox bloat from threshold/smoke tests — demo seed re-inserts 3 rows.
DELETE FROM gr33ncore.alerts_notifications WHERE farm_id = 1;

-- Smoke automation rules + their actions (keep demo AUTO Light*, coop gate rules, etc.).
DELETE FROM gr33ncore.executable_actions
WHERE rule_id IN (
    SELECT id FROM gr33ncore.automation_rules
    WHERE farm_id = 1
      AND (
          name ~ '_[0-9]{9,}$'
          OR name ~* '^(rule_ws|ws[0-9]+_|smoke_|phase[0-9]+|rule_cooldown|rule_inactive|rule_all_any|GH —)'
      )
);
DELETE FROM gr33ncore.automation_rules
WHERE farm_id = 1
  AND (
      name ~ '_[0-9]{9,}$'
      OR name ~* '^(rule_ws|ws[0-9]+_|smoke_|phase[0-9]+|rule_cooldown|rule_inactive|rule_all_any|GH —)'
  );

-- Smoke schedules + actions (uniqueName suffix or smoke_ prefix).
DELETE FROM gr33ncore.executable_actions
WHERE schedule_id IN (
    SELECT id FROM gr33ncore.schedules
    WHERE farm_id = 1
      AND (name ~ '_[0-9]{9,}$' OR name ~* '^smoke_')
);
DELETE FROM gr33ncore.schedules
WHERE farm_id = 1
  AND (name ~ '_[0-9]{9,}$' OR name ~* '^smoke_');

-- Soft-deleted animal/aquaponics smoke rows (Timeline clutter).
DELETE FROM gr33nanimals.animal_lifecycle_events
WHERE animal_group_id IN (
    SELECT id FROM gr33nanimals.animal_groups WHERE farm_id = 1 AND deleted_at IS NOT NULL
);
DELETE FROM gr33nanimals.animal_groups
WHERE farm_id = 1 AND deleted_at IS NOT NULL;

-- Smoke orgs/users (Settings page bloat: org_bootstrap_default_*, phase117_org_*, Viewer Smoke).
DELETE FROM gr33ncore.organizations
WHERE name ~ '^(org_bootstrap_default_|phase117_org_)';

DELETE FROM gr33ncore.farm_memberships
WHERE user_id IN (
    SELECT p.user_id FROM gr33ncore.profiles p
    WHERE p.email LIKE '%@test.local'
       OR p.full_name IN ('Viewer Smoke', 'Finance Smoke')
);

DELETE FROM gr33ncore.organization_memberships
WHERE user_id IN (
    SELECT p.user_id FROM gr33ncore.profiles p
    WHERE p.email LIKE '%@test.local'
       OR p.full_name IN ('Viewer Smoke', 'Finance Smoke')
);

DELETE FROM gr33ncore.profiles
WHERE email LIKE '%@test.local'
   OR full_name IN ('Viewer Smoke', 'Finance Smoke');

DELETE FROM auth.users
WHERE email LIKE '%@test.local';

-- Extra test farms (id > 1) are harmless if you stay on farm 1; deleting them hits FK
-- check constraints on executable_actions. Drop manually only when needed:
--   SELECT id, name FROM gr33ncore.farms WHERE id > 1;

COMMIT;
