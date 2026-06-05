/**
 * Phase 40 WS1 — client-side zone "Today" snapshot from data ZoneDetail already loads.
 */

import { humanizeCron, scheduleRunsLabel } from './cronHumanize.js'
import { PLANT_NEEDS, sensorPlantNeed } from './plantNeeds.js'

function parseTriggerConfig(rule) {
  try {
    const tc = rule?.trigger_configuration
    return typeof tc === 'string' ? JSON.parse(tc) : (tc || {})
  } catch {
    return {}
  }
}

/** Normalize legacy array or canonical `{ logic, predicates }` rule conditions. */
export function ruleConditionPredicates(rule) {
  try {
    const raw = rule?.conditions_jsonb
    const parsed = typeof raw === 'string'
      ? (() => { try { return JSON.parse(raw) } catch { return null } })()
      : raw
    if (Array.isArray(parsed)) return parsed
    if (parsed && Array.isArray(parsed.predicates)) return parsed.predicates
    return []
  } catch {
    return []
  }
}

/**
 * @param {object} rule
 * @param {number} zoneId
 * @param {string} zoneName
 * @param {Set<number>} zoneSensorIds
 */
export function ruleAppliesToZone(rule, zoneId, zoneName, zoneSensorIds) {
  const tc = parseTriggerConfig(rule)
  if (tc.zone_id != null && Number(tc.zone_id) === zoneId) return true
  if (zoneName && tc.target_zone && String(tc.target_zone).trim() === String(zoneName).trim()) {
    return true
  }
  for (const p of ruleConditionPredicates(rule)) {
    if (p?.sensor_id != null && zoneSensorIds.has(Number(p.sensor_id))) return true
  }
  return false
}

/**
 * @param {Array} alerts
 * @param {Array} zoneSensors
 * @param {string} [zoneName]
 */
export function alertMatchesZone(a, zoneSensors, zoneName = '') {
  const sensorIds = new Set(zoneSensors.map((s) => s.id))
  const name = String(zoneName || '').trim()
  if (a.triggering_event_source_type === 'sensor' && a.triggering_event_source_id != null) {
    return sensorIds.has(Number(a.triggering_event_source_id))
  }
  if (name && a.message_text_rendered && String(a.message_text_rendered).includes(name)) return true
  if (name && a.subject_rendered && String(a.subject_rendered).includes(name)) return true
  return false
}

/** Unread, unacknowledged alerts for a zone (Today strip). */
export function filterZoneAlerts(alerts, zoneSensors, zoneName = '') {
  return (alerts || []).filter(
    (a) => !a.is_read && !a.is_acknowledged && alertMatchesZone(a, zoneSensors, zoneName),
  )
}

/** All alerts tied to this room (Overview panel). */
export function filterZoneAlertsForRoom(alerts, zoneSensors, zoneName = '') {
  return (alerts || []).filter((a) => alertMatchesZone(a, zoneSensors, zoneName))
}

export function countZoneUnreadAlerts(alerts, zoneSensors, zoneName = '') {
  return filterZoneAlerts(alerts, zoneSensors, zoneName).length
}

/**
 * @param {object} params
 * @returns {number[]}
 */
export function collectZoneScheduleIds({
  zoneId,
  zoneName,
  schedules = [],
  activeProgram,
  lightingPrograms = [],
}) {
  const ids = new Set()
  if (activeProgram?.schedule_id) ids.add(Number(activeProgram.schedule_id))
  for (const lp of lightingPrograms) {
    if (lp.zone_id !== zoneId) continue
    if (lp.schedule_on_id) ids.add(Number(lp.schedule_on_id))
    if (lp.schedule_off_id) ids.add(Number(lp.schedule_off_id))
  }
  const name = String(zoneName || '').trim()
  for (const s of schedules) {
    if (!s?.is_active) continue
    if (ids.has(s.id)) continue
    const desc = `${s.description || ''} ${s.name || ''}`
    if (name && desc.includes(name)) ids.add(s.id)
  }
  return [...ids]
}

/**
 * @param {object} params
 */
export function pickNextZoneSchedule(params) {
  const ids = collectZoneScheduleIds(params)
  const linked = (params.schedules || []).filter((s) => ids.includes(s.id) && s.is_active !== false)
  if (!linked.length) return null
  const withHuman = linked.map((s) => ({
    schedule: s,
    label: scheduleRunsLabel(s),
    sortKey: humanizeCron(s.cron_expression) ? 0 : 1,
  }))
  withHuman.sort((a, b) => a.sortKey - b.sortKey || String(a.schedule.name).localeCompare(b.schedule.name))
  return withHuman[0]
}

/**
 * @param {object} params
 * @returns {{ chips: Array<{ id: string, icon: string, label: string, value: string, tone?: string }>, unreadAlerts: object[], activeRulesCount: number, queueDepth: number }}
 */
export function computeZoneTodaySnapshot(params) {
  const {
    zone,
    sensors = [],
    devices = [],
    alerts = [],
    rules = [],
    schedules = [],
    activeProgram,
    lightingPrograms = [],
    queueDepth = 0,
    zoneTasks = [],
  } = params

  const zoneId = zone?.id
  const zoneName = zone?.name || ''
  const zoneSensorIds = new Set(sensors.map((s) => s.id))
  const unreadAlerts = filterZoneAlerts(alerts, sensors, zoneName)

  const activeRules = (rules || []).filter(
    (r) => r.is_active && ruleAppliesToZone(r, zoneId, zoneName, zoneSensorIds),
  )

  const online = devices.filter((d) => d.status === 'online').length
  const offline = devices.length - online

  const nextSched = pickNextZoneSchedule({
    zoneId,
    zoneName,
    schedules,
    activeProgram,
    lightingPrograms,
  })

  const chips = []

  if (nextSched) {
    chips.push({
      id: 'next-schedule',
      icon: '📅',
      label: 'Next run',
      value: nextSched.label,
      detail: nextSched.schedule.name,
    })
  } else {
    chips.push({
      id: 'next-schedule',
      icon: '📅',
      label: 'Next run',
      value: 'Nothing scheduled',
      tone: 'muted',
    })
  }

  chips.push({
    id: 'active-rules',
    icon: '⚙️',
    label: 'Automations on',
    value: activeRules.length ? String(activeRules.length) : 'None',
    tone: activeRules.length ? 'ok' : 'muted',
  })

  chips.push({
    id: 'open-alerts',
    icon: '🔔',
    label: 'Open alerts',
    value: unreadAlerts.length ? String(unreadAlerts.length) : 'None',
    tone: unreadAlerts.length ? 'warn' : 'ok',
  })

  const deviceValue = devices.length
    ? (offline ? `${online} online · ${offline} offline` : `${online} online`)
    : 'No devices'
  chips.push({
    id: 'devices',
    icon: '📡',
    label: 'Devices',
    value: deviceValue,
    tone: offline ? 'warn' : 'ok',
  })

  if (queueDepth > 0) {
    chips.push({
      id: 'queue',
      icon: '⏳',
      label: 'Queued commands',
      value: String(queueDepth),
      tone: 'warn',
    })
  }

  if (zoneTasks.length) {
    chips.push({
      id: 'tasks',
      icon: '✅',
      label: 'Due today',
      value: String(zoneTasks.length),
    })
  }

  return {
    chips,
    unreadAlerts,
    activeRulesCount: activeRules.length,
    queueDepth,
    missingComfortTargets: countMissingComfortTargets(sensors, params.setpoints || []),
  }
}

function countMissingComfortTargets(sensors, setpoints) {
  const types = new Set(
    sensors
      .filter((s) => sensorPlantNeed(s.sensor_type) === PLANT_NEEDS.air)
      .map((s) => s.sensor_type),
  )
  let missing = 0
  for (const t of types) {
    const sp = setpoints.find((x) => x.sensor_type === t)
    if (!sp || (sp.min_value == null && sp.ideal_value == null && sp.max_value == null)) missing += 1
  }
  return missing
}
