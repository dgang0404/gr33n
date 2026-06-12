/**
 * Phase 88 — bundled fallback when GET /platform/domain-enums is unavailable.
 * Keep in sync with internal/platform/domainenums/enums.go order.
 */

/** @param {string} value */
function label(value) {
  return String(value).replace(/_/g, ' ')
}

/** @param {string[]} values */
function opts(values) {
  return values.map((value) => ({ value, label: label(value) }))
}

export const FALLBACK_DOMAIN_ENUMS = {
  growth_stages: opts([
    'clone', 'seedling', 'early_veg', 'late_veg', 'transition',
    'early_flower', 'mid_flower', 'late_flower', 'flush', 'harvest', 'dry_cure',
  ]),
  reservoir_statuses: opts([
    'ready', 'mixing', 'needs_top_up', 'needs_flush', 'flushing', 'offline', 'empty',
  ]),
  cost_categories: opts([
    'seeds_plants', 'fertilizers_soil_amendments', 'pest_disease_control', 'water_irrigation',
    'labor_wages', 'equipment_purchase_rental', 'equipment_maintenance_fuel',
    'utilities_electricity_gas', 'land_rent_mortgage', 'insurance', 'licenses_permits',
    'feed_livestock', 'veterinary_services', 'packaging_supplies', 'transportation_logistics',
    'marketing_sales', 'training_consultancy', 'miscellaneous',
  ]),
  application_targets: opts([
    'soil_drench', 'foliar_spray', 'seed_treatment', 'compost_pile_inoculant',
    'livestock_water_supplement', 'other',
  ]),
  input_definition_categories: opts([
    'microbial_inoculant', 'fermented_plant_juice', 'water_soluble_nutrient',
    'oriental_herbal_nutrient', 'fish_amino_acid', 'insect_attractant_repellent',
    'soil_conditioner', 'compost_tea_extract', 'biochar_preparation',
    'other_ferment', 'other_extract', 'animal_feed', 'bedding', 'veterinary_supply',
  ]),
  batch_statuses: opts([
    'planning', 'ingredients_gathered', 'mixing_in_progress', 'fermenting_brewing',
    'maturing_aging', 'ready_for_use', 'partially_used', 'fully_used',
    'expired_discarded', 'failed_production',
  ]),
}

export const FALLBACK_GROWTH_STAGE_VALUES = FALLBACK_DOMAIN_ENUMS.growth_stages.map((r) => r.value)
