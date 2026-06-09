/**
 * Phase 49 WS2 + Phase 68 WS5 — related sidebar routes for hover affordance.
 * @see docs/plans/phase_49_sidebar_nav_polish.plan.md
 * @see docs/plans/phase_68_workspace_shell_spa_nav.plan.md
 */

import { WORKSPACE_RELATIONS } from './workspaces.js'

/** @type {Record<string, string[]>} */
export const NAV_RELATIONS = {
  ...WORKSPACE_RELATIONS,
  '/plants': ['/zones', '/comfort-targets', '/money'],
  '/comfort-targets': ['/zones', '/feed-water', '/plants', '/automation', '/schedules'],
  '/operator-guide': ['/hardware'],
  '/alerts': ['/tasks', '/zones'],
  '/tasks': ['/alerts', '/schedules', '/zones'],
  // Legacy paths still linked from in-app copy — targets must be sidebar routes (Phase 68)
  '/feeding': ['/zones', '/comfort-targets', '/plants'],
  '/fertigation': ['/feed-water', '/zones', '/plants'],
  '/operations/feeding': ['/feed-water', '/money'],
  '/operations/supplies': ['/money', '/feed-water', '/tasks'],
  '/operations/money': ['/money', '/feed-water', '/plants'],
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
