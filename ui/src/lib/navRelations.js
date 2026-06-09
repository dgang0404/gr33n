/**
 * Phase 49 WS2 + Phase 68 WS5 — related sidebar routes for hover affordance.
 * @see docs/plans/phase_49_sidebar_nav_polish.plan.md
 * @see docs/plans/phase_68_workspace_shell_spa_nav.plan.md
 */

import { WORKSPACE_RELATIONS } from './workspaces.js'

/** @type {Record<string, string[]>} */
export const NAV_RELATIONS = {
  ...WORKSPACE_RELATIONS,
  '/comfort-targets': ['/zones', '/feed-water', '/automation', '/schedules'],
  '/operator-guide': ['/hardware'],
  '/tasks': ['/zones'],
  '/alerts': ['/zones'],
  '/plants': ['/zones'],
  // Legacy paths still linked from in-app copy — targets must be sidebar routes (Phase 68)
  '/feeding': ['/zones', '/comfort-targets'],
  '/fertigation': ['/feed-water', '/zones'],
  '/operations/feeding': ['/feed-water', '/money'],
  '/operations/supplies': ['/money', '/feed-water', '/zones'],
  '/operations/money': ['/money', '/feed-water'],
  '/sensors': ['/zones', '/hardware'],
  '/actuators': ['/zones', '/hardware'],
  '/lighting': ['/zones'],
  '/pi-setup': ['/hardware', '/zones'],
  '/costs': ['/money', '/feed-water'],
  '/inventory': ['/money', '/feed-water'],
}

/**
 * @param {string | null | undefined} route
 * @returns {string[]}
 */
export function relatedTo(route) {
  if (!route) return []
  const normalized = route.split('?')[0]
  return NAV_RELATIONS[normalized] ?? []
}
