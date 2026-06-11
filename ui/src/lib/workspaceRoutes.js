/**
 * Canonical SPA routes — use these instead of legacy /schedules, /feeding, etc.
 */

export const COMFORT_ADVANCED_SCHEDULES_HASH = '#comfort-advanced-schedules'
export const ZONE_WATER_PLAN_HASH = '#zone-water-plan'

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

export function zonesWorkspaceTabRoute(tab, { fleet } = {}) {
  /** @type {Record<string, string>} */
  const query = { tab }
  if (fleet) query.fleet = fleet
  return { path: '/zones', query }
}
