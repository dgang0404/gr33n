-- ============================================================
-- Queries: gr33nnaturalfarming.application_recipes & recipe_input_components
-- ============================================================

-- name: CreateRecipe :one
INSERT INTO gr33nnaturalfarming.application_recipes (
    farm_id, name, input_definition_id, description,
    target_application_type, dilution_ratio, instructions,
    frequency_guidelines, notes
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: ListRecipesByFarm :many
SELECT * FROM gr33nnaturalfarming.application_recipes
WHERE farm_id = $1 AND deleted_at IS NULL
ORDER BY name ASC;

-- name: GetRecipeByID :one
SELECT * FROM gr33nnaturalfarming.application_recipes
WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateRecipe :one
UPDATE gr33nnaturalfarming.application_recipes SET
    name = $2,
    input_definition_id = $3,
    description = $4,
    target_application_type = $5,
    dilution_ratio = $6,
    instructions = $7,
    frequency_guidelines = $8,
    notes = $9,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteRecipe :exec
UPDATE gr33nnaturalfarming.application_recipes
SET deleted_at = NOW()
WHERE id = $1;

-- name: ListRecipeComponents :many
SELECT
    c.application_recipe_id,
    c.input_definition_id,
    c.part_value,
    c.part_unit_id,
    c.notes,
    d.name AS input_name
FROM gr33nnaturalfarming.recipe_input_components c
JOIN gr33nnaturalfarming.input_definitions d ON d.id = c.input_definition_id
WHERE c.application_recipe_id = $1
ORDER BY d.name ASC;

-- name: AddRecipeComponent :exec
INSERT INTO gr33nnaturalfarming.recipe_input_components (
    application_recipe_id, input_definition_id, part_value, part_unit_id, notes
) VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (application_recipe_id, input_definition_id) DO UPDATE SET
    part_value = EXCLUDED.part_value,
    part_unit_id = EXCLUDED.part_unit_id,
    notes = EXCLUDED.notes;

-- name: RemoveRecipeComponent :exec
DELETE FROM gr33nnaturalfarming.recipe_input_components
WHERE application_recipe_id = $1 AND input_definition_id = $2;

-- Phase 211.02 — recipe revision history

-- name: GetLatestRecipeRevision :one
SELECT * FROM gr33nnaturalfarming.application_recipe_revisions
WHERE application_recipe_id = $1
ORDER BY revision_number DESC
LIMIT 1;

-- name: ListRecipeRevisions :many
SELECT * FROM gr33nnaturalfarming.application_recipe_revisions
WHERE application_recipe_id = $1
ORDER BY revision_number DESC;

-- name: GetRecipeRevisionByID :one
SELECT * FROM gr33nnaturalfarming.application_recipe_revisions
WHERE id = $1;

-- name: CreateRecipeRevision :one
WITH next AS (
    SELECT COALESCE(MAX(revision_number), 0) + 1 AS n
    FROM gr33nnaturalfarming.application_recipe_revisions
    WHERE application_recipe_id = sqlc.arg('application_recipe_id')
)
INSERT INTO gr33nnaturalfarming.application_recipe_revisions (
    application_recipe_id, revision_number, snapshot, change_summary, created_by_user_id
)
SELECT
    sqlc.arg('application_recipe_id'),
    next.n,
    sqlc.arg('snapshot'),
    sqlc.arg('change_summary'),
    sqlc.arg('created_by_user_id')
FROM next
RETURNING *;
