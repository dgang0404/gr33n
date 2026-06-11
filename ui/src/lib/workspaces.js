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
    subtitle: 'My zones, farm-wide hardware, and plant strains',
    tabs: [
      { id: 'rooms', label: 'My zones' },
      { id: 'fleet', label: 'Hardware & devices' },
      { id: 'strains', label: 'Plants & strains' },
    ],
    absorbs: {
      '/sensors': { tab: 'fleet', fleet: 'sensors' },
      '/actuators': { tab: 'fleet', fleet: 'controls' },
      '/lighting': { tab: 'fleet', fleet: 'lighting' },
      '/plants': { tab: 'strains' },
      '/feeding': { zoneTab: 'water' },
      '/operations/feeding': { zoneTab: 'water' },
      '/fertigation': { zoneTab: 'water' },
      '/pi-setup': { tab: 'fleet', fleet: 'sensors' },
    },
  },
  money: {
    label: 'Money',
    icon: '💰',
    route: '/money',
    subtitle: 'Spend, ledger, supply costs, and grow economics',
    tabs: [
      { id: 'summary', label: 'This month' },
      { id: 'ledger', label: 'Ledger' },
      { id: 'supplies', label: 'Supplies on hand' },
      { id: 'inventory', label: 'Inventory & recipes' },
      { id: 'grows', label: 'Grows' },
    ],
    absorbs: {
      '/operations/money': { tab: 'summary' },
      '/costs': { tab: 'ledger' },
      '/operations/supplies': { tab: 'supplies' },
      '/inventory': { tab: 'inventory' },
    },
  },
  help: {
    label: 'Help',
    icon: '📖',
    route: '/operator-guide',
    subtitle: 'Operator guide, knowledge search, and commons catalog',
    tabs: [
      { id: 'guide', label: 'Guide' },
      { id: 'knowledge', label: 'Knowledge' },
      { id: 'catalog', label: 'Catalog' },
    ],
    absorbs: {
      '/farm-knowledge': { tab: 'knowledge' },
      '/catalog': { tab: 'catalog' },
    },
  },
  comfort: {
    label: 'Comfort & automation',
    icon: '🎯',
    route: '/comfort-targets',
    subtitle: 'Comfort bands, what runs when, and automation toggles',
    tabs: [
      { id: 'comfort', label: 'Comfort' },
      { id: 'schedules', label: 'What runs when' },
      { id: 'automations', label: 'Automations' },
      { id: 'raw', label: 'Raw setpoints' },
    ],
    absorbs: {
      '/schedules': { tab: 'schedules' },
      '/automation': { tab: 'automations' },
      '/setpoints': { tab: 'raw' },
    },
  },
}

/** Hardware sub-views inside Zones → Hardware & devices tab. */
export const FLEET_SUB_TABS = [
  { id: 'sensors', label: 'Sensors' },
  { id: 'controls', label: 'Controls' },
  { id: 'lighting', label: 'Lighting' },
]

/** Cross-workspace jump targets (Phase 68 WS5, Phase 78 zone-first). */
export const WORKSPACE_RELATIONS = {
  '/zones': ['/money', '/comfort-targets', '/operator-guide'],
  '/money': ['/zones', '/operator-guide'],
  '/comfort-targets': ['/zones'],
  '/operator-guide': ['/zones', '/money'],
  '/chat': ['/zones', '/operator-guide'],
}

const LEGACY_ABSORB_INDEX = buildLegacyAbsorbIndex()

function buildLegacyAbsorbIndex() {
  /** @type {Record<string, { workspaceId: string, route: string, tab: string, fleet?: string, zoneTab?: string }>} */
  const index = {}
  for (const [workspaceId, ws] of Object.entries(WORKSPACES)) {
    for (const [legacyPath, target] of Object.entries(ws.absorbs ?? {})) {
      index[legacyPath] = {
        workspaceId,
        route: ws.route,
        tab: target.tab ?? ws.tabs[0]?.id ?? 'rooms',
        fleet: target.fleet,
        zoneTab: target.zoneTab,
      }
    }
  }
  return index
}

function parseZoneIdFromQuery(query) {
  const raw = query?.zone_id
  if (raw == null) return ''
  return String(Array.isArray(raw) ? raw[0] : raw).trim()
}

/**
 * Phase 78 — retired workspace routes → zones hub or zone detail.
 * @param {import('vue-router').RouteLocationNormalized} to
 */
export function redirectSunsetWorkspace(to) {
  const zoneId = parseZoneIdFromQuery(to.query)
  const query = { ...to.query }
  delete query.zone_id

  if (zoneId) {
    const tab = to.path === '/feed-water' ? 'water' : 'overview'
    return { path: `/zones/${zoneId}`, query: { ...query, tab } }
  }

  if (to.path === '/hardware') {
    return { path: '/zones', query: { ...query, tab: 'fleet', fleet: 'sensors' } }
  }

  return { path: '/zones', query }
}

/** @returns {Array<{ path: string, redirect: (to: import('vue-router').RouteLocationNormalized) => object }>} */
export function buildSunsetWorkspaceRedirects() {
  return [
    { path: '/feed-water', redirect: redirectSunsetWorkspace },
    { path: '/hardware', redirect: redirectSunsetWorkspace },
  ]
}

/**
 * @param {string | null | undefined} path
 * @returns {{ workspaceId: string, route: string, tab: string, fleet?: string, zoneTab?: string } | null}
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

/** Legacy comfort hub tab ids → workspace tab ids (Phase 75). */
const COMFORT_TAB_ALIASES = {
  bands: 'comfort',
  comfort: 'comfort',
  schedules: 'schedules',
  rules: 'automations',
  automations: 'automations',
  raw: 'raw',
}

/**
 * @param {string} workspaceId
 * @param {string | undefined | null} tabId
 * @returns {string}
 */
export function resolveWorkspaceTab(workspaceId, tabId) {
  const tabs = tabsFor(workspaceId)
  let resolved = tabId
  if (workspaceId === 'comfort' && tabId) {
    resolved = COMFORT_TAB_ALIASES[tabId] ?? tabId
  }
  if (resolved && tabs.some((t) => t.id === resolved)) return resolved
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
export function buildZoneOpsRedirectRoutes() {
  return [
    {
      path: '/tasks',
      redirect: (to) => redirectToZoneOps(to, 'tasks'),
    },
    {
      path: '/alerts',
      redirect: (to) => redirectToZoneOps(to, 'alerts'),
    },
  ]
}

/**
 * @param {import('vue-router').RouteLocationNormalized} to
 * @param {'tasks' | 'alerts'} ops
 */
function redirectToZoneOps(to, ops) {
  const raw = to.query.zone_id
  const zoneId = raw != null ? String(Array.isArray(raw) ? raw[0] : raw).trim() : ''
  const query = { ...to.query, tab: 'ops', ops }
  delete query.zone_id
  if (zoneId) {
    return { path: `/zones/${zoneId}`, query }
  }
  return { path: '/', query: {} }
}

/** @returns {Array<{ path: string, redirect: (to: import('vue-router').RouteLocationNormalized) => object }>} */
export function buildLegacyRedirectRoutes() {
  return Object.entries(LEGACY_ABSORB_INDEX).map(([legacyPath, hit]) => ({
    path: legacyPath,
    redirect: (to) => {
      const zoneId = parseZoneIdFromQuery(to.query)

      if (hit.zoneTab && zoneId) {
        const query = { ...to.query }
        delete query.zone_id
        return { path: `/zones/${zoneId}`, query: { ...query, tab: hit.zoneTab } }
      }

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
