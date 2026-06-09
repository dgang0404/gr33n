/**
 * Phase 68 WS1 — workspace model (SPA shells with internal tabs).
 * @see docs/plans/phase_68_workspace_shell_spa_nav.plan.md
 */

/** @typedef {{ id: string, label: string }} WorkspaceTab */
/** @typedef {{ tab: string, fleet?: string }} AbsorbTarget */

/** @type {Record<string, { label: string, icon: string, route: string, subtitle: string, tabs: WorkspaceTab[], absorbs?: Record<string, AbsorbTarget> }>} */
export const WORKSPACES = {
  zones: {
    label: 'Zones',
    icon: '🗂️',
    route: '/zones',
    subtitle: 'Every room — grows, sensors, controls, and lighting',
    tabs: [
      { id: 'rooms', label: 'Rooms' },
      { id: 'fleet', label: 'Fleet' },
    ],
    absorbs: {
      '/sensors': { tab: 'fleet', fleet: 'sensors' },
      '/actuators': { tab: 'fleet', fleet: 'controls' },
      '/lighting': { tab: 'fleet', fleet: 'lighting' },
    },
  },
  hardware: {
    label: 'Hardware',
    icon: '🔌',
    route: '/hardware',
    subtitle: 'Pi devices, GPIO wiring, and relay channels',
    tabs: [
      { id: 'board', label: 'GPIO board' },
      { id: 'devices', label: 'Pi devices' },
      { id: 'reference', label: 'Wiring guide' },
    ],
    absorbs: {
      '/pi-setup': { tab: 'reference' },
    },
  },
  feedwater: {
    label: 'Feed & Water',
    icon: '💧',
    route: '/feed-water',
    subtitle: 'Daily watering, programs, nutrients, and advanced fertigation',
    tabs: [
      { id: 'daily', label: 'Daily' },
      { id: 'programs', label: 'Programs & tanks' },
      { id: 'nutrients', label: 'Nutrients & mix' },
      { id: 'advanced', label: 'Advanced' },
    ],
    absorbs: {
      '/feeding': { tab: 'daily' },
      '/operations/feeding': { tab: 'programs' },
      '/fertigation': { tab: 'advanced' },
    },
  },
  money: {
    label: 'Money',
    icon: '💰',
    route: '/money',
    subtitle: 'Spend, ledger, and supply costs',
    tabs: [
      { id: 'summary', label: 'This month' },
      { id: 'ledger', label: 'Ledger' },
      { id: 'supplies', label: 'Supplies & costs' },
    ],
    absorbs: {
      '/operations/money': { tab: 'summary' },
      '/costs': { tab: 'ledger' },
      '/operations/supplies': { tab: 'supplies' },
      '/inventory': { tab: 'supplies' },
    },
  },
}

/** Fleet sub-views inside Zones → Fleet tab (Phase 69 will deepen). */
export const FLEET_SUB_TABS = [
  { id: 'sensors', label: 'Sensors' },
  { id: 'controls', label: 'Controls' },
  { id: 'lighting', label: 'Lighting' },
]

/** Cross-workspace jump targets (Phase 68 WS5). */
export const WORKSPACE_RELATIONS = {
  '/zones': ['/hardware', '/feed-water', '/money', '/comfort-targets', '/plants'],
  '/hardware': ['/zones', '/feed-water'],
  '/feed-water': ['/zones', '/hardware', '/money'],
  '/money': ['/feed-water', '/zones', '/plants'],
}

const LEGACY_ABSORB_INDEX = buildLegacyAbsorbIndex()

function buildLegacyAbsorbIndex() {
  /** @type {Record<string, { workspaceId: string, route: string, tab: string, fleet?: string }>} */
  const index = {}
  for (const [workspaceId, ws] of Object.entries(WORKSPACES)) {
    for (const [legacyPath, target] of Object.entries(ws.absorbs ?? {})) {
      index[legacyPath] = {
        workspaceId,
        route: ws.route,
        tab: target.tab,
        fleet: target.fleet,
      }
    }
  }
  return index
}

/**
 * @param {string | null | undefined} path
 * @returns {{ workspaceId: string, route: string, tab: string, fleet?: string } | null}
 */
export function workspaceFor(path) {
  if (!path) return null
  const normalized = path.split('?')[0]
  const hit = LEGACY_ABSORB_INDEX[normalized]
  if (hit) return hit
  for (const [workspaceId, ws] of Object.entries(WORKSPACES)) {
    if (ws.route === normalized) {
      return { workspaceId, route: ws.route, tab: ws.tabs[0]?.id ?? 'rooms' }
    }
  }
  return null
}

/**
 * Sidebar highlight path for v-nav-hint (legacy paths → workspace route).
 * @param {string | null | undefined} path
 * @returns {string | null}
 */
export function canonicalSidebarPath(path) {
  if (!path) return null
  const normalized = path.split('?')[0]
  return workspaceFor(normalized)?.route ?? normalized
}

/**
 * @param {string} workspaceId
 * @returns {WorkspaceTab[]}
 */
export function tabsFor(workspaceId) {
  return WORKSPACES[workspaceId]?.tabs ?? []
}

/**
 * @param {string} routePath
 * @returns {typeof WORKSPACES[string] | null}
 */
export function workspaceByRoute(routePath) {
  const normalized = routePath.split('?')[0]
  for (const ws of Object.values(WORKSPACES)) {
    if (ws.route === normalized) return ws
  }
  return null
}

/**
 * @param {string} workspaceId
 * @returns {string}
 */
export function defaultTabFor(workspaceId) {
  return WORKSPACES[workspaceId]?.tabs[0]?.id ?? ''
}

/**
 * @param {string} workspaceId
 * @param {string | undefined | null} tabId
 * @returns {string}
 */
export function resolveWorkspaceTab(workspaceId, tabId) {
  const tabs = tabsFor(workspaceId)
  if (tabId && tabs.some((t) => t.id === tabId)) return tabId
  return defaultTabFor(workspaceId)
}

/**
 * @param {string | undefined | null} fleetId
 * @returns {string}
 */
export function resolveFleetSubTab(fleetId) {
  if (fleetId && FLEET_SUB_TABS.some((t) => t.id === fleetId)) return fleetId
  return FLEET_SUB_TABS[0].id
}

/** @returns {Array<{ path: string, redirect: (to: import('vue-router').RouteLocationNormalized) => object }>} */
export function buildLegacyRedirectRoutes() {
  return Object.entries(LEGACY_ABSORB_INDEX).map(([legacyPath, hit]) => ({
    path: legacyPath,
    redirect: (to) => {
      const query = { ...to.query, tab: hit.tab }
      if (hit.fleet) query.fleet = hit.fleet
      return { path: hit.route, query }
    },
  }))
}

/**
 * @param {string | null | undefined} route
 * @returns {string[]}
 */
export function relatedWorkspaces(route) {
  if (!route) return []
  const normalized = canonicalSidebarPath(route) ?? route.split('?')[0]
  return WORKSPACE_RELATIONS[normalized] ?? []
}
