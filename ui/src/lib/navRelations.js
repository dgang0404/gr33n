/**
 * Phase 49 WS2 — related sidebar routes for hover affordance.
 * @see docs/plans/phase_49_sidebar_nav_polish.plan.md
 */

/** @type {Record<string, string[]>} */
export const NAV_RELATIONS = {
  '/zones': ['/feeding', '/comfort-targets'],
  '/feeding': ['/zones', '/comfort-targets'],
  '/comfort-targets': ['/zones', '/feeding'],
  '/actuators': ['/sensors', '/fertigation'],
  '/sensors': ['/actuators'],
  '/lighting': ['/fertigation'],
  '/pi-setup': ['/sensors', '/actuators'],
}

/**
 * @param {string | null | undefined} route
 * @returns {string[]}
 */
export function relatedTo(route) {
  if (!route) return []
  return NAV_RELATIONS[route] ?? []
}
