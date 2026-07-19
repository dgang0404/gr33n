-- Dev hygiene — strip Go smoke-test pollution from farm 1 (and optional extra farms).
-- Safe to re-run. Does NOT touch auth users or Docker volumes.
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

-- Extra test farms (id > 1) are harmless if you stay on farm 1; deleting them hits FK
-- check constraints on executable_actions. Drop manually only when needed:
--   SELECT id, name FROM gr33ncore.farms WHERE id > 1;

COMMIT;
