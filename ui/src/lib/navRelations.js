/**
 * Phase 49 WS2 — related sidebar routes for hover affordance.
 * @see docs/plans/phase_49_sidebar_nav_polish.plan.md
 */

/** @type {Record<string, string[]>} */
export const NAV_RELATIONS = {
  '/zones': ['/feeding', '/comfort-targets', '/plants'],
  '/plants': ['/zones', '/comfort-targets'],
  '/feeding': ['/zones', '/comfort-targets', '/plants'],
  '/comfort-targets': ['/zones', '/feeding', '/plants', '/automation', '/schedules'],
  '/actuators': ['/sensors', '/fertigation'],
  '/sensors': ['/actuators'],
  '/lighting': ['/fertigation'],
  '/pi-setup': ['/sensors', '/actuators'],
  '/operator-guide': ['/pi-setup'],
  '/alerts': ['/tasks', '/zones'],
  '/tasks': ['/alerts', '/schedules', '/zones'],
  '/operations/supplies': ['/tasks', '/operations/feeding'],
  '/operations/feeding': ['/operations/supplies', '/feeding'],
  '/operations/money': ['/operations/supplies', '/plants'],
  '/fertigation': ['/plants', '/zones', '/feeding', '/operations/feeding'],
}

/**
 * @param {string | null | undefined} route
 * @returns {string[]}
 */
export function relatedTo(route) {
  if (!route) return []
  return NAV_RELATIONS[route] ?? []
}
