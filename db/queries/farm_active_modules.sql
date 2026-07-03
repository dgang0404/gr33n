-- ============================================================
-- Queries: gr33ncore.farm_active_modules (Phase 115 WS1)
-- ============================================================

-- name: ListFarmActiveModules :many
SELECT * FROM gr33ncore.farm_active_modules
WHERE farm_id = $1
ORDER BY module_schema_name ASC;

-- name: UpsertFarmActiveModule :one
INSERT INTO gr33ncore.farm_active_modules (farm_id, module_schema_name, is_enabled, configuration)
VALUES ($1, $2, $3, coalesce($4::jsonb, '{}'::jsonb))
ON CONFLICT (farm_id, module_schema_name) DO UPDATE SET
  is_enabled = EXCLUDED.is_enabled,
  configuration = EXCLUDED.configuration,
  activated_at = CASE
    WHEN gr33ncore.farm_active_modules.is_enabled = FALSE AND EXCLUDED.is_enabled THEN NOW()
    ELSE gr33ncore.farm_active_modules.activated_at
  END
RETURNING *;

-- name: FarmModuleIsEnabled :one
SELECT COALESCE(
  (SELECT is_enabled FROM gr33ncore.farm_active_modules
   WHERE farm_id = $1 AND module_schema_name = $2),
  FALSE
)::boolean AS is_enabled;

-- name: SeedFarmActiveModule :exec
INSERT INTO gr33ncore.farm_active_modules (farm_id, module_schema_name, is_enabled, configuration)
VALUES ($1, $2, $3, '{}'::jsonb)
ON CONFLICT (farm_id, module_schema_name) DO NOTHING;
