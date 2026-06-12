-- Phase 102 — Tag demo fertigation programs and recipes with crop_key + stage metadata.

UPDATE gr33nfertigation.programs p
SET metadata = COALESCE(p.metadata, '{}'::jsonb) || jsonb_build_object(
    'recommended_crop_keys', jsonb_build_array('cannabis', 'tomato', 'pepper', 'cucumber'),
    'recommended_stages', jsonb_build_array('early_veg', 'late_veg'),
    'profile_ec_source', jsonb_build_object('crop_key', 'cannabis', 'stage', 'late_veg'),
    'ec_band_mscm', jsonb_build_object('min', 1.4, 'max', 2.2)
)
WHERE p.deleted_at IS NULL
  AND p.name = 'Veg Daily JLF Program';

UPDATE gr33nfertigation.programs p
SET metadata = COALESCE(p.metadata, '{}'::jsonb) || jsonb_build_object(
    'recommended_crop_keys', jsonb_build_array('cannabis'),
    'recommended_stages', jsonb_build_array('early_flower', 'mid_flower', 'late_flower', 'flush'),
    'profile_ec_source', jsonb_build_object('crop_key', 'cannabis', 'stage', 'early_flower'),
    'ec_band_mscm', jsonb_build_object('min', 1.6, 'max', 2.4)
)
WHERE p.deleted_at IS NULL
  AND p.name = 'Flower Daily FFJ+WCA Program';

UPDATE gr33nnaturalfarming.application_recipes r
SET target_crop_categories = ARRAY['cannabis', 'tomato', 'pepper', 'cucumber'],
    target_growth_stages = ARRAY['early_veg', 'late_veg']
WHERE r.deleted_at IS NULL
  AND r.name LIKE 'JLF and JMS Combined Drench%';

UPDATE gr33nnaturalfarming.application_recipes r
SET target_crop_categories = ARRAY['cannabis'],
    target_growth_stages = ARRAY['early_flower', 'mid_flower', 'late_flower', 'flush']
WHERE r.deleted_at IS NULL
  AND r.name LIKE 'FFJ and WCA Flowering Boost%';
