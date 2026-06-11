/**
 * Phase 49 WS2 + Phase 68 WS5 + Phase 78 zone-first — related sidebar routes for hover affordance.
 */

import { WORKSPACE_RELATIONS } from './workspaces.js'

/** @type {Record<string, string[]>} */
export const NAV_RELATIONS = {
  ...WORKSPACE_RELATIONS,
  '/comfort-targets': ['/zones'],
  '/operator-guide': ['/zones', '/money'],
  '/chat': ['/zones', '/operator-guide'],
  '/tasks': ['/zones'],
  '/alerts': ['/zones'],
  '/plants': ['/zones'],
  '/schedules': ['/comfort-targets'],
  '/automation': ['/comfort-targets'],
  '/setpoints': ['/comfort-targets'],
  '/feeding': ['/zones', '/comfort-targets'],
  '/fertigation': ['/zones'],
  '/operations/feeding': ['/money', '/zones'],
  '/operations/supplies': ['/money', '/zones'],
  '/operations/money': ['/money', '/zones'],
  '/sensors': ['/zones'],
  '/actuators': ['/zones'],
  '/lighting': ['/zones'],
  '/pi-setup': ['/zones', '/operator-guide'],
  '/costs': ['/money'],
  '/inventory': ['/money'],
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
