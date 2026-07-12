/**
 * Phase 174 — Today farm health rollup + greeting helpers.
 */

import { zoneNeedsAttention } from './zoneQuickActions.js'

/**
 * @param {object} params
 * @param {object[]} params.zones
 * @param {(zone: object) => object} params.getStatus
 * @param {number} [params.tasksTodayCount]
 * @param {number} [params.unreadAlerts]
 * @param {number} [params.overdueTaskCount]
 */
export function buildFarmTodayRollup({
  zones = [],
  getStatus,
  tasksTodayCount = 0,
  unreadAlerts = 0,
  overdueTaskCount = 0,
}) {
  let healthy = 0
  let attention = 0
  for (const zone of zones) {
    const status = getStatus?.(zone)
    if (zoneNeedsAttention(status)) attention += 1
    else if (status?.health === 'ok') healthy += 1
  }
  return {
    healthy,
    attention,
    tasksTodayCount,
    unreadAlerts,
    overdueTaskCount,
  }
}

/**
 * @param {Date} [now]
 */
export function todayTimeGreeting(now = new Date()) {
  const hour = now.getHours()
  if (hour < 12) return 'Good morning'
  if (hour < 17) return 'Good afternoon'
  return 'Good evening'
}

/**
 * @param {object|null|undefined} siteWeather
 * @param {Date} [now]
 */
export function todayHeaderSubtitle(siteWeather, now = new Date()) {
  const greeting = todayTimeGreeting(now)
  if (siteWeather?.solar?.sunrise_at && siteWeather?.solar?.sunset_at) {
    return `${greeting} — daylight on your farm today`
  }
  return greeting
}
