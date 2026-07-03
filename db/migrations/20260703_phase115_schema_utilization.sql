-- Phase 115 — seed farm_active_modules for existing farms; drop unused validation_rules.

INSERT INTO gr33ncore.farm_active_modules (farm_id, module_schema_name, is_enabled, configuration)
SELECT f.id, m.module_schema_name, m.is_enabled, '{}'::jsonb
FROM gr33ncore.farms f
CROSS JOIN (VALUES
  ('gr33ncrops', TRUE),
  ('gr33nnaturalfarming', TRUE),
  ('gr33nanimals', FALSE),
  ('gr33naquaponics', FALSE)
) AS m(module_schema_name, is_enabled)
WHERE f.deleted_at IS NULL
ON CONFLICT (farm_id, module_schema_name) DO NOTHING;

-- WS8: validation_rules had no reader; per-field validation lives in handlers.
DROP TABLE IF EXISTS gr33ncore.validation_rules CASCADE;
