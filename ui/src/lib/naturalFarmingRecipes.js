/**
 * Phase 209 WS4 — farm recipe CRUD + apply-to-zone helpers.
 */
import { enumLabel } from './domainEnums.js'
import { filterProgramsForZone } from './zoneContext.js'
import { feedWaterFertigationRoute } from './workspaceRoutes.js'

export function emptyRecipeForm() {
  return {
    name: '',
    target_application_type: 'soil_drench',
    input_definition_id: null,
    description: '',
    dilution_ratio: '',
    instructions: '',
    frequency_guidelines: '',
    notes: '',
  }
}

/**
 * Deep link into Feed & water advanced (Fertigation programs tab).
 * @param {number|string|null|undefined} recipeId
 * @param {{ zoneId?: number|string|null, tab?: string }} [opts]
 */
export function feedWaterProgramLink(recipeId, { zoneId = null, tab = 'programs' } = {}) {
  return feedWaterFertigationRoute(tab, { recipe: recipeId, zoneId })
}

/**
 * @param {number|null|undefined} zoneId
 * @param {object[]} programs
 * @param {object[]} cropCycles
 */
export function programsForZone(zoneId, programs, cropCycles) {
  if (!zoneId) return programs || []
  return filterProgramsForZone(programs, zoneId, cropCycles)
}

/**
 * @param {number|string|null|undefined} recipeId
 * @param {object[]} programs
 */
export function programsUsingRecipe(recipeId, programs) {
  const rid = Number(recipeId)
  if (!Number.isFinite(rid)) return []
  return (programs || []).filter((p) => Number(p.application_recipe_id) === rid)
}

/**
 * @param {string[]|null|undefined} stages
 * @param {object|null|undefined} domainEnums
 */
export function formatGrowthStages(stages, domainEnums) {
  const list = Array.isArray(stages) ? stages.filter(Boolean) : []
  if (!list.length) return '—'
  return list.map((s) => enumLabel('growth_stages', s, domainEnums)).join(', ')
}

const LIVESTOCK_APPLICATION_TYPES = new Set(['livestock_water_supplement'])
const LIVESTOCK_INPUT_CATEGORIES = new Set(['animal_feed', 'livestock_water_supplement'])

/**
 * @param {object|null|undefined} recipe
 * @param {Record<number, object>} inputsById
 */
export function isLivestockRecipe(recipe, inputsById = {}) {
  if (!recipe) return false
  if (LIVESTOCK_APPLICATION_TYPES.has(String(recipe.target_application_type || ''))) return true
  const inputId = recipe.input_definition_id
  if (inputId == null) return false
  const inp = inputsById[inputId] ?? inputsById[Number(inputId)]
  return LIVESTOCK_INPUT_CATEGORIES.has(String(inp?.category || ''))
}

/**
 * @param {object[]} inputs
 */
export function inputsByIdMap(inputs) {
  /** @type {Record<number, object>} */
  const map = {}
  for (const i of inputs || []) {
    if (i?.id != null) map[i.id] = i
  }
  return map
}
