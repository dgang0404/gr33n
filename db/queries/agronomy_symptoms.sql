-- name: ListAgronomySymptomEntries :many
SELECT *
FROM gr33ncrops.agronomy_symptom_entries
WHERE published = TRUE
ORDER BY sort_order, symptom_key;

-- name: ListAgronomySymptomsForCrop :many
-- Returns entries matching crop_key, category, or universal (empty crop_keys).
SELECT *
FROM gr33ncrops.agronomy_symptom_entries
WHERE published = TRUE
  AND (
    cardinality(crop_keys) = 0
    OR sqlc.arg('crop_key')::text = ANY(crop_keys)
    OR (
      sqlc.narg('category')::text IS NOT NULL
      AND sqlc.narg('category')::text = ANY(categories)
    )
  )
ORDER BY sort_order, symptom_key;
