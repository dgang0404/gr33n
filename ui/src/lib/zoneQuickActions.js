/**
 * Phase 167 — quick-action sheet helpers and mobile stack ordering.
 */

import { filterZoneAlerts } from './zoneGrowSummary.js'
import { actuatorGHRole, supportsPulseCommand } from './plantNeeds.js'

const LIGHT_TYPES = new Set(['light', 'dimmer', 'relay'])
const WATER_TYPES = new Set(['pump', 'drip', 'valve', 'relay'])

function norm(s) {
  return String(s || '').toLowerCase()
}

function isLightActuator(a) {
  const t = norm(a.actuator_type)
  return LIGHT_TYPES.has(t) || norm(a.name).includes('light')
}

function isWaterActuator(a) {
  const t = norm(a.actuator_type)
  return WATER_TYPES.has(t) || t.includes('pump') || t.includes('drip') || norm(a.name).includes('drip')
}

/**
 * @param {object} status from computeZoneVisualStatus
 * @param {boolean} hasTasksDue
 */
export function zoneStackSortKey(status, hasTasksDue = false) {
  if (status?.health === 'alert') return 0
  if (status?.health === 'warn' || (status?.attention?.length > 0)) return 1
  if (hasTasksDue) return 2
  if (status?.plants?.state === 'empty') return 4
  return 3
}

/**
 * @param {object[]} zones
 * @param {(zone: object) => object} getStatus
 * @param {(zoneId: number) => boolean} [hasTasksDue]
 */
export function sortZonesForStack(zones, getStatus, hasTasksDue = () => false) {
  return [...zones].sort((a, b) => {
    const ka = zoneStackSortKey(getStatus(a), hasTasksDue(a.id))
    const kb = zoneStackSortKey(getStatus(b), hasTasksDue(b.id))
    if (ka !== kb) return ka - kb
    return String(a.name).localeCompare(String(b.name))
  })
}

/**
 * @param {object} params
 */
export function activeProgramForZone(programs, zoneId) {
  return (programs || []).find(
    (p) => Number(p.target_zone_id) === Number(zoneId) && p.is_active !== false,
  ) || null
}

/**
 * @param {object} params
 */
export function resolveWaterNowAction({ zone, programs = [], actuators = [] }) {
  const program = activeProgramForZone(programs, zone?.id)
  if (program) {
    const mins = program.run_duration_seconds
      ? Math.max(1, Math.round(program.run_duration_seconds / 60))
      : null
    const durationLabel = mins ? `${mins} min` : 'now'
    return {
      mode: 'program',
      program,
      label: `Water now — ${program.name}`,
      confirm: `Run ${program.name}${mins ? ` — ${durationLabel}` : ''}?`,
      durationSeconds: program.run_duration_seconds ?? 180,
    }
  }

  const zoneActs = (actuators || []).filter((a) => Number(a.zone_id) === Number(zone?.id))
  const waterAct = zoneActs.find(isWaterActuator)
  if (waterAct && supportsPulseCommand(waterAct.actuator_type)) {
    return {
      mode: 'pulse',
      actuator: waterAct,
      label: 'Water now — timed pulse',
      confirm: `Run ${waterAct.name} for 60 seconds?`,
      defaultSeconds: 60,
    }
  }
  if (waterAct) {
    return {
      mode: 'toggle',
      actuator: waterAct,
      label: `Water now — ${waterAct.name}`,
      confirm: `Turn on ${waterAct.name}?`,
    }
  }

  return {
    mode: 'setup',
    label: 'Set up watering',
    link: { path: `/zones/${zone.id}`, query: { tab: 'water' } },
  }
}

/**
 * @param {object[]} actuators
 * @param {number} zoneId
 */
export function lightActuatorsForZone(actuators, zoneId) {
  return (actuators || []).filter(
    (a) => Number(a.zone_id) === Number(zoneId) && isLightActuator(a),
  )
}

/**
 * @param {object} zone
 * @param {object[]} actuators
 */
export function greenhouseActuatorsForZone(zone, actuators) {
  if (norm(zone?.zone_type) !== 'greenhouse') return []
  let meta = zone.meta_data
  if (typeof meta === 'string') {
    try { meta = JSON.parse(meta) } catch { meta = {} }
  }
  const gc = meta?.greenhouse_climate || {}
  const zoneActs = (actuators || []).filter((a) => Number(a.zone_id) === Number(zone?.id))
  const out = []
  if (gc.shade_actuator_id) {
    const a = zoneActs.find((x) => Number(x.id) === Number(gc.shade_actuator_id))
    if (a) out.push({ actuator: a, role: 'shade', commands: ['deploy', 'retract'] })
  }
  if (gc.vent_actuator_id) {
    const a = zoneActs.find((x) => Number(x.id) === Number(gc.vent_actuator_id))
    if (a) out.push({ actuator: a, role: 'vent', commands: ['open', 'close'] })
  }
  for (const a of zoneActs) {
    const role = actuatorGHRole(a.actuator_type)
    if (role === 'fan' && !out.some((x) => x.actuator.id === a.id)) {
      out.push({ actuator: a, role: 'fan', commands: ['on', 'off'] })
    }
  }
  return out
}

/**
 * @param {object[]} tasks
 * @param {number} zoneId
 */
export function zoneTasksForSheet(tasks, zoneId, limit = 3) {
  const today = new Date().toISOString().slice(0, 10)
  return (tasks || [])
    .filter((t) => {
      if (Number(t.zone_id) !== Number(zoneId)) return false
      if (t.status === 'completed' || t.status === 'cancelled') return false
      if (!t.due_date) return false
      return String(t.due_date).slice(0, 10) <= today
    })
    .slice(0, limit)
}

/**
 * @param {object[]} alerts
 * @param {object[]} sensors
 * @param {object} zone
 */
export function zoneAlertsForSheet(alerts, sensors, zone, limit = 3) {
  const zoneSensors = (sensors || []).filter((s) => Number(s.zone_id) === Number(zone?.id))
  return filterZoneAlerts(alerts, zoneSensors, zone?.name).slice(0, limit)
}

/**
 * @param {object[]} tasks
 * @param {number} zoneId
 */
export function zoneHasTasksDueToday(tasks, zoneId) {
  return zoneTasksForSheet(tasks, zoneId, 99).length > 0
}

/**
 * @param {object|null|undefined} status from computeZoneVisualStatus
 */
export function zoneNeedsAttention(status) {
  if (!status) return false
  return status.health === 'alert'
    || status.health === 'warn'
    || (status.attention?.length > 0)
}

/**
 * One-line farmer summary for attention strip chips.
 * @param {object|null|undefined} status
 */
export function zoneAttentionSummary(status) {
  if (!status) return 'Needs attention'
  const item = status.attention?.[0]
  if (item?.label) return item.label
  if (status.sensors?.worst === 'attention' && status.sensors?.summary) {
    return status.sensors.summary
  }
  if (status.health === 'alert') return 'Needs urgent attention'
  if (status.health === 'warn') return 'Worth a look'
  return 'Needs attention'
}

/**
 * @param {object[]} zones
 * @param {(zone: object) => object} getStatus
 */
export function listAttentionZones(zones, getStatus) {
  return (zones || [])
    .map((zone) => ({ zone, status: getStatus(zone) }))
    .filter(({ status }) => zoneNeedsAttention(status))
}
