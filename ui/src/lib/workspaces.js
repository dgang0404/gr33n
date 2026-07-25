import { redirectLegacyInventory } from './workspaceRoutes.js'
import { NF_WORKSPACE_TAB_LABELS } from './naturalFarmingVocabulary.js'

/**
 * Phase 68 WS1 — workspace model (SPA shells with internal tabs).
 * @see docs/plans/archive/phase_68_workspace_shell_spa_nav.plan.md
 */

/** @typedef {{ id: string, label: string, conceptId?: string, description?: string }} WorkspaceTab */
/** @typedef {{ tab: string, fleet?: string, section?: string }} AbsorbTarget */

/** @type {Record<string, { label: string, icon: string, route: string, subtitle: string, tabs: WorkspaceTab[], absorbs?: Record<string, AbsorbTarget> }>} */
export const WORKSPACES = {
  zones: {
    label: 'Zones',
    icon: '🗂️',
    route: '/zones',
    subtitle: 'Rooms, farm-wide hardware, and your plant catalog — organized by where you grow.',
    tabs: [
      {
        id: 'rooms',
        label: 'My zones',
        description: 'Open a zone to manage water, light, climate, and wiring in one place. Add zones as your layout grows.',
      },
      {
        id: 'fleet',
        label: 'Hardware & devices',
        description: 'Sensors, controls, and lighting programs across every zone — grouped so you can spot gaps before they bite.',
      },
      {
        id: 'plants',
        label: 'Plants',
        description: 'Catalog crop types and track active grows farm-wide. Start a batch in a zone or compare harvests later.',
      },
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
    subtitle: 'Track spending, receipts, supply costs, and what each grow actually earned.',
    tabs: [
      {
        id: 'summary',
        label: 'This month',
        description: 'What you spent and received this month — save receipts, tag grows, and spot trends without ledger jargon on the first screen.',
      },
      {
        id: 'ledger',
        label: 'Ledger',
        description: 'Full cost ledger with categories, energy rates, and exports — for when you need every line item.',
      },
      {
        id: 'supplies',
        label: 'Supplies on hand',
        conceptId: 'input_batch',
        description: 'Inputs and batches on hand with quantities and restock costs. Link supplies to recipes and feeding programs.',
      },
      {
        id: 'grows',
        label: 'Grows',
        description: 'Compare harvests and open grow summaries for cost-per-gram and yield context across rooms.',
      },
    ],
    absorbs: {
      '/operations/money': { tab: 'summary' },
      '/costs': { tab: 'ledger' },
      '/operations/supplies': { tab: 'supplies' },
    },
  },
  help: {
    label: 'Help',
    icon: '📖',
    route: '/operator-guide',
    subtitle: 'Operator guides, Pi setup, farm search, symptoms, and Commons import packs.',
    tabs: [
      {
        id: 'library',
        label: 'Library',
        description: 'Step-by-step operator guides — start here when you are learning a workflow end to end.',
      },
      {
        id: 'pi-setup',
        label: 'Pi + HAT setup',
        description: 'Wire Sequent Microsystems HATs, register Pis, and validate edge config before you trust automation.',
      },
      {
        id: 'knowledge',
        label: 'Search',
        description: 'Search farm knowledge, docs, and saved notes across everything you have imported.',
      },
      {
        id: 'symptoms',
        label: 'Symptom guide',
        description: 'Look up crop symptoms and tie them back to zones, recipes, and Guardian suggestions.',
      },
      {
        id: 'catalog',
        label: 'Import',
        description: 'Import Commons packs — crop catalogs, natural-farming recipes, and shared operator content.',
      },
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
    subtitle: 'Comfort bands per zone, schedules, automation rules, and raw setpoints when you need them.',
    tabs: [
      {
        id: 'comfort',
        label: 'Comfort',
        conceptId: 'comfort_band',
        description: 'Set temperature and humidity bands per zone so automation knows when to heat, cool, or dehumidify.',
      },
      {
        id: 'schedules',
        label: 'What runs when',
        conceptId: 'schedule',
        description: 'Define what runs when — feeds, lights, and routines — without touching cron unless you want to.',
      },
      {
        id: 'automations',
        label: 'Automations',
        conceptId: 'rule',
        description: 'Toggle and tune automation rules that react to sensor readings and comfort targets.',
      },
      {
        id: 'raw',
        label: 'Raw setpoints',
        conceptId: 'setpoint',
        description: 'Direct setpoint values for power users — use when comfort bands and schedules are not enough.',
      },
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
    subtitle: 'GPIO board layout, registered Pi devices, and wiring reference for edge hardware.',
    tabs: [
      {
        id: 'board',
        label: 'GPIO board',
        description: 'Visual GPIO map for your HAT — see which pins are power, ground, assigned, or free to wire.',
      },
      {
        id: 'devices',
        label: 'Pi devices',
        description: 'Edge devices registered on this farm — API keys, last seen, and links to zone wiring.',
      },
      {
        id: 'reference',
        label: 'Wiring guide',
        description: 'Pi + HAT setup steps, DIP switches, and safe wiring practices before you energize relays.',
      },
    ],
    absorbs: {
      '/pi-setup': { tab: 'reference' },
    },
  },
  feedwater: {
    label: 'Feed & water',
    icon: '💧',
    route: '/feed-water',
    subtitle: 'Daily watering, fertigation programs, nutrient mixing, and the full console when you need it.',
    tabs: [
      {
        id: 'daily',
        label: 'Daily',
        conceptId: 'feeding_plan',
        description: 'One card per zone — next feed, last run, and plan status. Open a zone to edit water on the Water tab.',
      },
      {
        id: 'programs',
        label: 'Programs & tanks',
        conceptId: 'fertigation_program',
        description: 'Feeding programs, tank assignments, and schedules that drive automatic watering per zone.',
      },
      {
        id: 'nutrients',
        label: 'Nutrients & mix',
        conceptId: 'mixing_event',
        description: 'Mix nutrients, log EC/pH targets, and record what went into each reservoir batch.',
      },
      {
        id: 'advanced',
        label: 'Advanced',
        conceptId: 'fertigation_console',
        description: 'Full fertigation console — reservoirs, programs, crop cycles, and event history for power users.',
      },
    ],
    absorbs: {
      '/feeding': { tab: 'daily' },
      '/operations/feeding': { tab: 'programs' },
      '/fertigation': { tab: 'advanced' },
    },
  },
  naturalfarming: {
    label: 'Natural farming',
    icon: '🌱',
    route: '/natural-farming',
    subtitle: 'Ferment inputs, read the canon field guide, and wire apply recipes to your zones.',
    tabs: [
      {
        id: 'batch',
        label: NF_WORKSPACE_TAB_LABELS.batch,
        conceptId: 'input_batch',
        description: 'Pick an input type, follow the field guide steps, and record a batch on this farm with cost and prep tasks.',
      },
      {
        id: 'library',
        label: NF_WORKSPACE_TAB_LABELS.library,
        conceptId: 'nf_field_guide',
        description: 'Read-only canon — ingredients and prep for JADAM/KNF inputs, apply recipes, and livestock programs. Not your on-hand inventory.',
      },
      {
        id: 'recipes',
        label: NF_WORKSPACE_TAB_LABELS.recipes,
        conceptId: 'application_recipe',
        description: 'Your farm apply recipes — create, edit, and link them to Feed & water programs so Guardian and automation know what to run.',
      },
    ],
    absorbs: {
      '/inventory': { tab: 'recipes' },
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
  '/zones': ['/feed-water', '/natural-farming', '/hardware', '/money', '/comfort-targets', '/operator-guide'],
  '/hardware': ['/zones', '/feed-water', '/operator-guide'],
  '/feed-water': ['/zones', '/natural-farming', '/money', '/operator-guide'],
  '/natural-farming': ['/feed-water', '/money', '/zones', '/operator-guide'],
  '/money': ['/zones', '/feed-water', '/natural-farming', '/operator-guide'],
  '/comfort-targets': ['/zones', '/feed-water', '/natural-farming'],
  '/operator-guide': ['/zones', '/money', '/feed-water', '/natural-farming'],
  '/chat': ['/zones', '/feed-water', '/natural-farming', '/operator-guide'],
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
  if (normalized === '/hardware') return '/virtual-pi'
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

/** Legacy natural-farming tab ids (Phase 211.02). */
const NATURAL_FARMING_TAB_ALIASES = {
  start: 'batch',
  guide: 'batch',
  stock: 'batch',
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
  if (workspaceId === 'naturalfarming' && tabId) {
    resolved = NATURAL_FARMING_TAB_ALIASES[tabId] ?? tabId
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

/** Named legacy paths so `{ name: 'automation' }` resolves (Comfort workspace absorb). */
const LEGACY_ROUTE_NAMES = {
  '/automation': 'automation',
  '/schedules': 'schedules',
  '/setpoints': 'setpoints',
}

/** @returns {Array<{ path: string, redirect: (to: import('vue-router').RouteLocationNormalized) => object }>} */
export function buildLegacyRedirectRoutes() {
  return Object.entries(LEGACY_ABSORB_INDEX).map(([legacyPath, hit]) => ({
    path: legacyPath,
    ...(LEGACY_ROUTE_NAMES[legacyPath] ? { name: LEGACY_ROUTE_NAMES[legacyPath] } : {}),
    redirect: (to) => {
      if (legacyPath === '/inventory') {
        return redirectLegacyInventory(to)
      }

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

      if (legacyPath === '/fertigation') {
        const query = { ...to.query, tab: hit.tab }
        const rawTab = to.query.tab
        const sub = typeof rawTab === 'string' ? rawTab : Array.isArray(rawTab) ? rawTab[0] : ''
        const fertSubTabs = new Set(['reservoirs', 'ec-targets', 'programs', 'mixing', 'crop-cycles', 'events'])
        if (sub && fertSubTabs.has(sub)) query.fert_tab = sub
        return { path: hit.route, query }
      }

      if (legacyPath === '/operations/feeding') {
        const query = { ...to.query, tab: hit.tab }
        const rawTab = to.query.tab
        const sub = typeof rawTab === 'string' ? rawTab : Array.isArray(rawTab) ? rawTab[0] : ''
        const adminSubTabs = new Set(['programs', 'reservoirs', 'ec-targets'])
        if (sub && adminSubTabs.has(sub) && sub !== 'programs') query.admin_tab = sub
        return { path: hit.route, query }
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
  const normalized = route.split('?')[0]
  if (WORKSPACE_RELATIONS[normalized]) return WORKSPACE_RELATIONS[normalized]
  const canonical = canonicalSidebarPath(normalized)
  return WORKSPACE_RELATIONS[canonical] ?? []
}
