/**
 * Phase 76 — Today dashboard deep links into workspaces and zone Ops.
 */

import { isOpenTask, todayDateIso } from './zoneTasks.js'
import { alertMatchesZone } from './zoneGrowSummary.js'

/** @typedef {'tasks' | 'alerts'} ZoneOpsSubTab */

/**
 * @param {number|string} zoneId
 * @param {ZoneOpsSubTab} [ops]
 */
export function zoneOpsRoute(zoneId, ops = 'tasks') {
  return { path: `/zones/${zoneId}`, query: { tab: 'ops', ops } }
}

/**
 * @param {object[]} [zones]
 * @param {number|string|null} [zoneId]
 */
export function feedWaterRoute(zones = [], zoneId = null) {
  const zid = zoneId ?? zones[0]?.id
  if (zid) return { path: `/zones/${zid}`, query: { tab: 'water' } }
  return { path: '/zones' }
}

/**
 * @param {string} [tab]
 */
export function comfortRoute(tab = 'schedules') {
  return { path: '/comfort-targets', query: { tab } }
}

/**
 * @param {string} [tab]
 */
export function moneyRoute(tab = 'summary') {
  return { path: '/money', query: { tab } }
}

/**
 * @param {object[]} tasks
 */
export function firstOpenTaskZoneId(tasks) {
  const today = todayDateIso()
  const open = (tasks || []).filter(isOpenTask)
  const due = open.find((t) => t.zone_id && String(t.due_date || '').slice(0, 10) <= today)
  if (due?.zone_id) return due.zone_id
  const any = open.find((t) => t.zone_id)
  return any?.zone_id ?? null
}

/**
 * @param {object[]} alerts
 * @param {object[]} zones
 * @param {object[]} sensors
 */
export function firstUnreadAlertZoneId(alerts, zones = [], sensors = []) {
  const unread = (alerts || []).filter((a) => !a.is_read && !a.is_acknowledged)
  for (const zone of zones) {
    const zoneSensors = sensors.filter((s) => Number(s.zone_id) === Number(zone.id))
    if (unread.some((a) => alertMatchesZone(a, zoneSensors, zone.name))) {
      return zone.id
    }
  }
  return null
}

/**
 * @param {object[]} tasks
 * @param {object[]} zones
 */
export function tasksViewAllRoute(tasks, zones = []) {
  const zid = firstOpenTaskZoneId(tasks)
  if (zid) return zoneOpsRoute(zid, 'tasks')
  if (zones.length === 1) return zoneOpsRoute(zones[0].id, 'tasks')
  return { path: '/zones' }
}

/**
 * @param {object[]} alerts
 * @param {object[]} zones
 * @param {object[]} sensors
 */
export function alertsViewAllRoute(alerts, zones = [], sensors = []) {
  const zid = firstUnreadAlertZoneId(alerts, zones, sensors)
  if (zid) return zoneOpsRoute(zid, 'alerts')
  if (zones.length === 1) return zoneOpsRoute(zones[0].id, 'alerts')
  return { path: '/zones' }
}

/**
 * @param {object[]} tasks
 * @param {object[]} zones
 */
export function newTaskRoute(tasks, zones = []) {
  const zid = firstOpenTaskZoneId(tasks) || zones[0]?.id
  if (zid) {
    return { path: `/zones/${zid}`, query: { tab: 'ops', ops: 'tasks', create: '1' } }
  }
  return { path: '/zones' }
}
