/**
 * Canonical SPA routes — use these instead of legacy /schedules, /feeding, etc.
 */

export const COMFORT_ADVANCED_SCHEDULES_HASH = '#comfort-advanced-schedules'
export const ZONE_WATER_PLAN_HASH = '#zone-water-plan'
export const ZONE_FEED_HISTORY_HASH = '#zone-feed-history'
export const ZONE_HARDWARE_HASH = '#zone-hardware'

/**
 * @param {string} tab comfort | schedules | automations | raw
 * @param {{ zoneId?: number|string, hash?: string }} [opts]
 */
export function comfortTabRoute(tab, { zoneId, hash } = {}) {
  /** @type {Record<string, string>} */
  const query = { tab }
  if (zoneId != null && String(zoneId).trim() !== '') {
    query.zone_id = String(zoneId)
  }
  return {
    path: '/comfort-targets',
    query,
    ...(hash ? { hash } : {}),
  }
}

/** Scroll target for cron / precondition schedule editor (same workspace tab). */
export function comfortAdvancedSchedulesRoute(opts = {}) {
  return comfortTabRoute('schedules', { ...opts, hash: COMFORT_ADVANCED_SCHEDULES_HASH })
}

/**
 * @param {number|string} zoneId
 * @param {string} [tab] overview | water | light | air | ops | plants
 * @param {string} [hash]
 */
export function zoneTabRoute(zoneId, tab = 'overview', hash) {
  return {
    path: `/zones/${zoneId}`,
    query: tab && tab !== 'overview' ? { tab } : {},
    ...(hash ? { hash } : {}),
  }
}

export function zoneWaterPlanRoute(zoneId) {
  return zoneTabRoute(zoneId, 'water', ZONE_WATER_PLAN_HASH)
}

export function zoneHardwareRoute(zoneId) {
  return zoneTabRoute(zoneId, 'overview', ZONE_HARDWARE_HASH)
}

/**
 * @param {string} tab summary | ledger | supplies | inventory | grows
 * @param {{ inv?: string, zoneId?: number|string }} [opts]
 */
export function moneyTabRoute(tab, { inv, zoneId } = {}) {
  /** @type {Record<string, string>} */
  const query = { tab }
  if (inv) query.inv = inv
  if (zoneId != null) query.zone_id = String(zoneId)
  return { path: '/money', query }
}

/**
 * @param {{ batchId?: number|string, zoneId?: number|string }} [opts]
 */
export function moneySuppliesRoute({ batchId, zoneId } = {}) {
  /** @type {Record<string, string>} */
  const query = { tab: 'supplies' }
  if (batchId != null && batchId !== '') query.batch_id = String(batchId)
  if (zoneId != null && zoneId !== '') query.zone_id = String(zoneId)
  return { path: '/money', query }
}

/**
 * @param {string} tab start | library | batch | recipes | stock
 * @param {{ recipe?: number|string, batchId?: number|string, process?: string, zoneId?: number|string }} [opts]
 */
export function naturalFarmingTabRoute(tab, { recipe, batchId, process, zoneId } = {}) {
  if (tab === 'stock' || tab === 'manage') {
    return moneySuppliesRoute({ batchId, zoneId })
  }
  /** @type {Record<string, string>} */
  const query = { tab }
  if (recipe != null && recipe !== '') query.recipe = String(recipe)
  if (batchId != null && batchId !== '') query.batch_id = String(batchId)
  if (process) query.process = process
  if (zoneId != null && zoneId !== '') query.zone_id = String(zoneId)
  return { path: '/natural-farming', query }
}

/**
 * Legacy manage-tab deep links → Money supplies (Phase 211.03 WS8).
 * @param {{ inv?: string, batchId?: number|string, zoneId?: number|string }} [opts]
 */
export function naturalFarmingManageRoute({ inv, batchId, zoneId } = {}) {
  if (inv === 'recipes') {
    return naturalFarmingTabRoute('recipes', { zoneId })
  }
  return moneySuppliesRoute({ batchId, zoneId })
}

/**
 * @param {import('vue-router').RouteLocationNormalized} to
 */
export function redirectNaturalFarmingManageTab(to) {
  if (to.query.tab !== 'manage') return null
  const q = { ...to.query }
  delete q.tab
  const inv = typeof q.inv === 'string' ? q.inv : Array.isArray(q.inv) ? q.inv[0] : undefined
  delete q.inv
  if (inv === 'recipes') {
    return { path: '/natural-farming', query: { ...q, tab: 'recipes' } }
  }
  return moneySuppliesRoute({
    batchId: q.batch_id,
    zoneId: q.zone_id,
  })
}

/**
 * Legacy Money inventory tab → Natural farming manage / apply recipes.
 * @param {import('vue-router').RouteLocationNormalized} to
 */
export function redirectMoneyInventoryTab(to) {
  if (to.query.tab !== 'inventory') return null
  const q = { ...to.query }
  delete q.tab
  const inv = typeof q.inv === 'string' ? q.inv : Array.isArray(q.inv) ? q.inv[0] : undefined
  delete q.inv
  if (inv === 'recipes') {
    return { path: '/natural-farming', query: { ...q, tab: 'recipes' } }
  }
  return moneySuppliesRoute({
    batchId: q.batch_id,
    zoneId: q.zone_id,
  })
}

/**
 * Legacy ?tab=stock on Natural farming → Money supplies (ready batches live there).
 * @param {import('vue-router').RouteLocationNormalized} to
 */
export function redirectNaturalFarmingStockTab(to) {
  if (to.query.tab !== 'stock') return null
  const q = { ...to.query }
  delete q.tab
  return moneySuppliesRoute({ batchId: q.batch_id, zoneId: q.zone_id })
}

/**
 * Phase 209 WS6 — legacy /inventory → natural-farming studio (definitions stay on Money).
 * @param {import('vue-router').RouteLocationNormalized} to
 */
export function redirectLegacyInventory(to) {
  const q = { ...to.query }
  const inv = q.inv ?? q.tab
  delete q.inv
  if (inv === 'definitions') {
    return moneySuppliesRoute({ zoneId: q.zone_id })
  }
  if (inv === 'batches' || q.batch_id) {
    return moneySuppliesRoute({ batchId: q.batch_id, zoneId: q.zone_id })
  }
  return { path: '/natural-farming', query: { ...q, tab: 'recipes' } }
}

export function zonesWorkspaceTabRoute(tab, { fleet } = {}) {
  /** @type {Record<string, string>} */
  const query = { tab }
  if (fleet) query.fleet = fleet
  return { path: '/zones', query }
}

/** Fertigation sub-tabs inside Feed & water → Advanced (workspace tab stays `advanced`). */
export const FERTIGATION_SUB_TABS = [
  'reservoirs',
  'ec-targets',
  'programs',
  'mixing',
  'crop-cycles',
  'events',
]

function queryStringParam(query, key) {
  const raw = query?.[key]
  if (raw == null) return undefined
  const s = Array.isArray(raw) ? raw[0] : raw
  return typeof s === 'string' ? s : undefined
}

/** @param {Record<string, unknown>} [query] */
export function resolveFertigationSubTab(query) {
  const fromFert = queryStringParam(query, 'fert_tab')
  if (fromFert && FERTIGATION_SUB_TABS.includes(fromFert)) return fromFert
  const fromTab = queryStringParam(query, 'tab')
  if (fromTab && FERTIGATION_SUB_TABS.includes(fromTab)) return fromTab
  return undefined
}

/**
 * @param {string} [fertTab]
 * @param {{ recipe?: number|string, zoneId?: number|string, query?: Record<string, string> }} [opts]
 */
export function feedWaterFertigationRoute(fertTab = 'reservoirs', { recipe, zoneId, query: extra = {} } = {}) {
  /** @type {Record<string, string>} */
  const query = { ...extra, tab: 'advanced' }
  if (fertTab) query.fert_tab = fertTab
  if (recipe != null && recipe !== '') query.recipe = String(recipe)
  if (zoneId != null && zoneId !== '') query.zone_id = String(zoneId)
  return { path: '/feed-water', query }
}
