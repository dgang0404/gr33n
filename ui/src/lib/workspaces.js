/**
 * Phase 68 WS1 — workspace model (SPA shells with internal tabs).
 * @see docs/plans/phase_68_workspace_shell_spa_nav.plan.md
 */

/** @typedef {{ id: string, label: string }} WorkspaceTab */
/** @typedef {{ tab: string, fleet?: string, section?: string }} AbsorbTarget */

/** @type {Record<string, { label: string, icon: string, route: string, subtitle: string, tabs: WorkspaceTab[], absorbs?: Record<string, AbsorbTarget> }>} */
export const WORKSPACES = {
  zones: {
    label: 'Zones',
    icon: '🗂️',
    route: '/zones',
    subtitle: 'My zones, farm-wide hardware, and plants',
    tabs: [
      { id: 'rooms', label: 'My zones' },
      { id: 'fleet', label: 'Hardware & devices' },
      { id: 'plants', label: 'Plants' },
    ],
    absorbs: {
      '/sensors': { tab: 'fleet', fleet: 'sensors' },
      '/actuators': { tab: 'fleet', fleet: 'controls' },
      '/lighting': { tab: 'fleet', fleet: 'lighting' },
      '/plants': { tab: 'plants' },
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
    subtitle: 'How-to, search, symptoms, and import packs',
    tabs: [
      { id: 'library', label: 'Library' },
      { id: 'pi-setup', label: 'Pi + HAT setup' },
      { id: 'knowledge', label: 'Search' },
      { id: 'symptoms', label: 'Symptom guide' },
      { id: 'catalog', label: 'Import' },
    ],
    absorbs: {
      '/farm-knowledge': { tab: 'knowledge' },
      '/catalog': { tab: 'catalog' },
      '/symptom-guide': { tab: 'symptoms' },
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
  hardware: {
    label: 'Hardware',
    icon: '🖥️',
    route: '/hardware',
    subtitle: 'GPIO board, Pi devices, and wiring reference',
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
    label: 'Feed & water',
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
}

/** Hardware sub-views inside Zones → Hardware & devices tab. */
export const FLEET_SUB_TABS = [
  { id: 'sensors', label: 'Sensors' },
  { id: 'controls', label: 'Controls' },
  { id: 'lighting', label: 'Lighting' },
]

/** Cross-workspace jump targets (Phase 68 WS5, Phase 78 zone-first). */
export const WORKSPACE_RELATIONS = {
  '/zones': ['/feed-water', '/hardware', '/money', '/comfort-targets', '/operator-guide'],
  '/hardware': ['/zones', '/feed-water', '/operator-guide'],
  '/feed-water': ['/zones', '/money', '/operator-guide'],
  '/money': ['/zones', '/feed-water', '/operator-guide'],
  '/comfort-targets': ['/zones', '/feed-water'],
  '/operator-guide': ['/zones', '/money', '/feed-water'],
  '/chat': ['/zones', '/feed-water', '/operator-guide'],
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
        section: target.section,
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
 * Phase 78 — retired workspace routes with zone_id → zone detail (feed-water/hardware are live again in 70/71).
 * @param {import('vue-router').RouteLocationNormalized} to
 */
export function redirectSunsetWorkspace(to) {
  const zoneId = parseZoneIdFromQuery(to.query)
  const query = { ...to.query }
  delete query.zone_id

  if (zoneId && to.path === '/feed-water') {
    return { path: `/zones/${zoneId}`, query: { ...query, tab: 'water' } }
  }

  return { path: '/zones', query }
}

/** @returns {Array<{ path: string, redirect: (to: import('vue-router').RouteLocationNormalized) => object }>} */
export function buildSunsetWorkspaceRedirects() {
  return []
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

/** Legacy zones tab ids → workspace tab ids (Phase 93). */
const ZONES_TAB_ALIASES = {
  strains: 'plants',
  plants: 'plants',
  rooms: 'rooms',
  fleet: 'fleet',
}

/** Legacy feed-water tab ids → workspace tab ids (Phase 71). */
const FEEDWATER_TAB_ALIASES = {
  daily: 'daily',
  programs: 'programs',
  nutrients: 'nutrients',
  advanced: 'advanced',
  water: 'daily',
}

/** Phase 183 — legacy Help tab ids map to workspace tabs. */
const HELP_TAB_ALIASES = {
  guide: 'library',
  knowledge: 'knowledge',
  symptoms: 'symptoms',
  catalog: 'catalog',
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
  if (workspaceId === 'zones' && tabId) {
    resolved = ZONES_TAB_ALIASES[tabId] ?? tabId
  }
  if (workspaceId === 'feedwater' && tabId) {
    resolved = FEEDWATER_TAB_ALIASES[tabId] ?? tabId
  }
  if (workspaceId === 'help' && tabId) {
    resolved = HELP_TAB_ALIASES[tabId] ?? tabId
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

      if (zoneId && (legacyPath === '/feeding' || legacyPath === '/operations/feeding' || legacyPath === '/fertigation')) {
        const query = { ...to.query }
        delete query.zone_id
        return { path: `/zones/${zoneId}`, query: { ...query, tab: 'water' } }
      }

      const query = { ...to.query, tab: hit.tab }
      if (hit.fleet) query.fleet = hit.fleet
      if (hit.section) query.section = hit.section
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
