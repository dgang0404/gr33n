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
 * Farm recipes whose primary input matches a batch input definition.
 * @param {number|string|null|undefined} inputDefinitionId
 * @param {object[]} recipes
 */
export function recipesForInput(inputDefinitionId, recipes) {
  const iid = Number(inputDefinitionId)
  if (!Number.isFinite(iid)) return []
  return (recipes || []).filter((r) => Number(r.input_definition_id) === iid)
}

/**
 * Best single recipe to open from an on-hand batch row.
 * @param {number|string|null|undefined} inputDefinitionId
 * @param {object[]} recipes
 */
export function primaryRecipeForInput(inputDefinitionId, recipes) {
  const matches = recipesForInput(inputDefinitionId, recipes)
  if (matches.length === 1) return matches[0]
  if (matches.length > 1) {
    return [...matches].sort((a, b) => String(a.name || '').localeCompare(String(b.name || '')))[0]
  }
  return null
}

/**
 * Match a farm recipe row to canon seed_name (exact, then prefix/contains).
 * @param {string|null|undefined} canonName
 * @param {object[]} recipes
 */
export function findFarmRecipeByName(canonName, recipes) {
  const needle = String(canonName || '').trim().toLowerCase()
  if (!needle) return null
  const list = recipes || []
  const exact = list.find((r) => String(r.name || '').trim().toLowerCase() === needle)
  if (exact) return exact
  return list.find((r) => {
    const name = String(r.name || '').trim().toLowerCase()
    return name.startsWith(needle) || needle.startsWith(name) || name.includes(needle)
  }) ?? null
}

/**
 * Deep link into Recipes & apply for a ready batch (input-linked, then name).
 * @param {{ inputDefinitionId?: number|string|null, inputName?: string|null }} row
 * @param {object[]} recipes
 */
export function recipeApplyRouteForStockRow(row, recipes) {
  const byInput = primaryRecipeForInput(row?.inputDefinitionId, recipes)
  const hit = byInput || findFarmRecipeByName(row?.inputName, recipes)
  return hit?.id ?? null
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
