/**
 * Phase 211.01/211.02 — Natural farming workspace defaults (no API).
 */

/** Default tab when route has no ?tab= */
export const NATURAL_FARMING_DEFAULT_TAB = 'batch'

/**
 * @param {number} [_recipeCount] — ignored; kept for test compat
 */
export function defaultNaturalFarmingTab(_recipeCount) {
  return NATURAL_FARMING_DEFAULT_TAB
}
