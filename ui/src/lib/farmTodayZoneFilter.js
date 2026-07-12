/**
 * Phase 173 — Today zone filters + paging for large farms.
 */

import { zoneNeedsAttention } from './zoneQuickActions.js'

export const TODAY_ZONE_FILTERS = [
  { id: 'all', label: 'All zones' },
  { id: 'attention', label: 'Needs attention' },
  { id: 'indoor', label: 'Indoor' },
  { id: 'outdoor', label: 'Outdoor' },
  { id: 'greenhouse', label: 'Greenhouse' },
]

const FILTER_BAR_THRESHOLD = 9
const MOBILE_PAGE_SIZE = 8
const DESKTOP_LIST_THRESHOLD = 13

function zoneTypeMatches(zone, filterId) {
  const t = String(zone?.zone_type || '').toLowerCase()
  if (filterId === 'greenhouse') return t.includes('greenhouse')
  if (filterId === 'outdoor') return t.includes('outdoor')
  if (filterId === 'indoor') return !t.includes('outdoor') && !t.includes('greenhouse')
  return true
}

/**
 * @param {object[]} zones
 * @param {string} filterId
 * @param {(zone: object) => object} getStatus
 */
export function filterZonesForToday(zones, filterId, getStatus) {
  const list = zones || []
  if (!filterId || filterId === 'all') return list
  if (filterId === 'attention') {
    return list.filter((zone) => zoneNeedsAttention(getStatus?.(zone)))
  }
  return list.filter((zone) => zoneTypeMatches(zone, filterId))
}

/**
 * @param {object[]} zones
 * @param {(zone: object) => object} getStatus
 */
export function countZonesPerFilter(zones, getStatus) {
  const counts = {}
  for (const f of TODAY_ZONE_FILTERS) {
    counts[f.id] = filterZonesForToday(zones, f.id, getStatus).length
  }
  return counts
}

/**
 * @param {number} zoneCount
 */
export function shouldShowTodayZoneFilterBar(zoneCount) {
  return Number(zoneCount || 0) >= FILTER_BAR_THRESHOLD
}

/**
 * @param {number} zoneCount
 * @param {number} [pageSize]
 */
export function shouldPageZoneStack(zoneCount, pageSize = MOBILE_PAGE_SIZE) {
  return Number(zoneCount || 0) > pageSize
}

/**
 * @param {number} zoneCount
 */
export function shouldOfferDesktopListView(zoneCount) {
  return Number(zoneCount || 0) > DESKTOP_LIST_THRESHOLD
}

/**
 * @param {object[]} zones
 * @param {number} page zero-indexed
 * @param {number} [pageSize]
 */
export function paginateZones(zones, page, pageSize = MOBILE_PAGE_SIZE) {
  const list = zones || []
  const start = Math.max(0, page) * pageSize
  return list.slice(start, start + pageSize)
}

/**
 * @param {number} zoneCount
 * @param {number} [pageSize]
 */
export function totalZonePages(zoneCount, pageSize = MOBILE_PAGE_SIZE) {
  return Math.max(1, Math.ceil(Number(zoneCount || 0) / pageSize))
}

export { MOBILE_PAGE_SIZE }

const SESSION_FILTER_KEY = 'gr33n_today_zone_filter'
const SESSION_VIEW_KEY = 'gr33n_today_desktop_view'

function safeSessionGet(key) {
  try {
    return sessionStorage.getItem(key)
  } catch {
    return null
  }
}

function safeSessionSet(key, value) {
  try {
    sessionStorage.setItem(key, value)
  } catch {
    /* best effort — private browsing or SSR */
  }
}

export function readTodayZoneFilter() {
  const v = safeSessionGet(SESSION_FILTER_KEY)
  return TODAY_ZONE_FILTERS.some((f) => f.id === v) ? v : 'all'
}

export function writeTodayZoneFilter(filterId) {
  safeSessionSet(SESSION_FILTER_KEY, filterId)
}

export function readTodayDesktopView() {
  const v = safeSessionGet(SESSION_VIEW_KEY)
  return v === 'list' ? 'list' : 'map'
}

export function writeTodayDesktopView(view) {
  safeSessionSet(SESSION_VIEW_KEY, view === 'list' ? 'list' : 'map')
}
