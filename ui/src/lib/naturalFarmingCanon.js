/**
 * Phase 209 WS2 — fetch canonical natural farming YAML via read API.
 */
import api from '../api'

let cachedCanon = null

/** @returns {Promise<Record<string, unknown>>} */
export async function loadRecipeCanon({ force = false } = {}) {
  if (cachedCanon && !force) return cachedCanon
  const { data } = await api.get('/v1/field-guides/recipe-canon')
  cachedCanon = data ?? {}
  return cachedCanon
}

/** Test-only reset. */
export function resetRecipeCanonCache() {
  cachedCanon = null
}

/**
 * @param {string} recipeName
 * @param {Record<string, unknown>} canon
 */
export function findApplicationRecipe(recipeName, canon) {
  const recipes = /** @type {Array<Record<string, unknown>>} */ (canon?.application_recipes ?? [])
  return recipes.find((r) => r.seed_name === recipeName) ?? null
}

/**
 * @param {string} seedName
 * @param {Record<string, unknown>} canon
 */
export function findCanonInput(seedName, canon) {
  const inputs = /** @type {Array<Record<string, unknown>>} */ (canon?.inputs ?? [])
  return inputs.find((i) => i.seed_name === seedName) ?? null
}
