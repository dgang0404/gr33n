/**
 * Phase 166 WS1 — per-zone visual status for the Today farm canvas.
 */

import { scheduleRunsLabel } from './cronHumanize.js'
import { filterZoneAlerts, alertMatchesZone } from './zoneGrowSummary.js'
import { buildZoneFeedingPlan } from './zoneFeedingPlan.js'
import { lastZoneFeedEvent } from './zoneWaterGrowStory.js'

const SENSOR_DEFAULTS = {
  temperature: { low: 15, high: 30 },
  humidity: { low: 40, high: 80 },
  co2: { low: 400, high: 1500 },
  ph: { low: 5.5, high: 7.0 },
  ec: { low: 1.0, high: 3.5 },
  par: { low: 100, high: 900 },
  soil_moisture: { low: 30, high: 80 },
}

const STAGE_LABELS = {
  clone: 'Clone stage',
  seedling: 'Seedling stage',
  early_veg: 'Early veg',
  late_veg: 'Veg stage',
  transition: 'Transition',
  early_flower: 'Bloom stage',
  mid_flower: 'Bloom stage',
  late_flower: 'Late bloom',
  harvest: 'Harvest',
  dry: 'Drying',
  cure: 'Cure',
}

/** Recommended default positions for the demo farm (normalized 0–1). */
export const DEFAULT_ZONE_LAYOUTS_BY_NAME = {
  'Veg Room': { x: 0.04, y: 0.06, w: 0.20, h: 0.18 },
  'Flower Room': { x: 0.28, y: 0.06, w: 0.20, h: 0.18 },
  'Propagation Room': { x: 0.52, y: 0.06, w: 0.20, h: 0.18 },
  'Herb & Greens Room': { x: 0.76, y: 0.06, w: 0.20, h: 0.18 },
  'Outdoor Garden': { x: 0.10, y: 0.32, w: 0.24, h: 0.20 },
  'Outdoor Pepper Bed': { x: 0.38, y: 0.34, w: 0.20, h: 0.18 },
  'Outdoor Berry Patch': { x: 0.62, y: 0.34, w: 0.20, h: 0.18 },
}

export const DEFAULT_TILE_W = 0.20
export const DEFAULT_TILE_H = 0.18

function parseMeta(raw) {
  if (!raw) return {}
  if (typeof raw === 'string') {
    try { return JSON.parse(raw) } catch { return {} }
  }
  return { ...raw }
}

function thresholdLow(sensor) {
  if (sensor?.alert_threshold_low != null) return Number(sensor.alert_threshold_low)
  return SENSOR_DEFAULTS[sensor?.sensor_type]?.low ?? 0
}

function thresholdHigh(sensor) {
  if (sensor?.alert_threshold_high != null) return Number(sensor.alert_threshold_high)
  return SENSOR_DEFAULTS[sensor?.sensor_type]?.high ?? 100
}

/**
 * @param {object} sensor
 * @param {object|null} reading
 * @param {boolean} alertLinked
 */
export function classifySensorHardwareState(sensor, reading, alertLinked = false) {
  if (!reading || reading.value_raw == null) return 'not_set_up'
  if (reading.is_valid === false || alertLinked) return 'attention'
  const val = Number(reading.value_raw)
  if (!Number.isFinite(val)) return 'not_set_up'
  if (val < thresholdLow(sensor) || val > thresholdHigh(sensor)) return 'attention'
  if (reading.reading_time) {
    const ageMs = Date.now() - new Date(reading.reading_time).getTime()
    if (ageMs > 48 * 3600 * 1000) return 'not_set_up'
  }
  return 'healthy'
}

function sensorAlertLinked(alert, sensor) {
  return alert?.triggering_event_source_type === 'sensor'
    && Number(alert.triggering_event_source_id) === Number(sensor.id)
}

function formatStage(stage) {
  if (!stage) return ''
  return STAGE_LABELS[stage] || String(stage).replace(/_/g, ' ')
}

function activeCropCycle(cropCycles, zoneId) {
  return (cropCycles || []).find(
    (c) => Number(c.zone_id) === Number(zoneId) && c.is_active !== false,
  ) || null
}

function activeProgramForZone(programs, zoneId) {
  return (programs || []).find(
    (p) => Number(p.target_zone_id) === Number(zoneId) && p.is_active !== false,
  ) || null
}

function zoneActuators(actuators, zoneId) {
  return (actuators || []).filter((a) => Number(a.zone_id) === Number(zoneId))
}

function resolveWaterKind(program, actuators, zoneId) {
  if (!program) {
    return { kind: 'none', label: 'No watering set up', nextRun: null, lastEvent: null }
  }
  const zoneActs = zoneActuators(actuators, zoneId)
  const hasDrip = zoneActs.some(
    (a) => String(a.actuator_type || '').toLowerCase() === 'drip'
      || String(a.name || '').toLowerCase().includes('drip'),
  )
  const name = String(program.name || '').toLowerCase()
  const isGravity = hasDrip || name.includes('gravity') || name.includes('drip')
  const plain = Boolean(program.irrigation_only)

  let kind = 'pump'
  let label = plain ? 'Plain water' : 'Pump feed'
  if (isGravity) {
    kind = 'gravity_drip'
    label = plain ? 'Gravity drip · plain water' : 'Gravity drip'
  } else if (plain) {
    kind = 'manual'
    label = 'Plain water'
  }

  return { kind, label, plain }
}

function resolveLight({ zone, schedules, actuators, programs }) {
  const meta = parseMeta(zone?.meta_data)
  const zoneActs = zoneActuators(actuators, zone?.id)
  const lightActs = zoneActs.filter(
    (a) => ['light', 'dimmer', 'relay'].includes(String(a.actuator_type || '').toLowerCase())
      || String(a.name || '').toLowerCase().includes('light'),
  )

  if (meta.lighting_program_id) {
    const linked = (schedules || []).filter((s) => {
      const md = parseMeta(s.meta_data)
      return Number(md.lighting_program_id) === Number(meta.lighting_program_id)
        || String(s.description || s.name || '').includes(zone?.name || '')
    })
    const active = linked.find((s) => s.is_active !== false)
    return {
      state: active ? 'scheduled' : 'none',
      scheduleLabel: active ? scheduleRunsLabel(active) : null,
    }
  }

  const program = activeProgramForZone(programs, zone?.id)
  const sched = program?.schedule_id
    ? (schedules || []).find((s) => Number(s.id) === Number(program.schedule_id))
    : null
  const schedName = `${sched?.name || ''} ${sched?.description || ''}`.toLowerCase()
  if (sched && sched.is_active !== false && (schedName.includes('light') || schedName.includes('18/6') || schedName.includes('12/12'))) {
    return { state: 'scheduled', scheduleLabel: scheduleRunsLabel(sched) }
  }

  if (lightActs.length) {
    const on = lightActs.some((a) => String(a.last_command || a.desired_state || '').toLowerCase() === 'on')
    return { state: on ? 'on' : 'off', scheduleLabel: null }
  }

  return { state: 'none', scheduleLabel: null }
}

function rollupSensorStates(states) {
  if (!states.length) {
    return { state: 'not_set_up', summary: 'Not set up yet', worst: 'not_set_up', healthy: 0, attention: 0, unset: 0 }
  }
  const healthy = states.filter((s) => s === 'healthy').length
  const attention = states.filter((s) => s === 'attention').length
  const unset = states.filter((s) => s === 'not_set_up').length

  if (attention > 0 && healthy === 0 && unset === 0) {
    return { state: 'attention', summary: 'Needs attention', worst: 'attention', healthy, attention, unset }
  }
  if (unset > 0 && healthy === 0 && attention === 0) {
    return { state: 'not_set_up', summary: 'Not set up yet', worst: 'not_set_up', healthy, attention, unset }
  }
  if (attention > 0) {
    const label = attention === 1 ? '1 sensor needs attention' : `${attention} sensors need attention`
    return { state: 'mixed', summary: label, worst: 'attention', healthy, attention, unset }
  }
  if (unset > 0 && healthy > 0) {
    return {
      state: 'mixed',
      summary: unset === 1 ? '1 sensor not set up yet' : `${unset} sensors not set up yet`,
      worst: 'not_set_up',
      healthy,
      attention,
      unset,
    }
  }
  const label = healthy === 1 ? '1 sensor healthy' : `${healthy} sensors healthy`
  return { state: 'healthy', summary: label, worst: 'healthy', healthy, attention, unset }
}

function attentionItems({ zone, zoneSensors, zoneAlerts, zoneTasks }) {
  const items = []
  for (const a of zoneAlerts.slice(0, 3)) {
    items.push({
      kind: 'alert',
      label: a.title || a.message_text_rendered || a.subject_rendered || 'Alert',
      severity: a.severity || 'warning',
      link: { path: `/zones/${zone.id}`, query: { tab: 'alerts' } },
    })
  }
  const today = new Date().toISOString().slice(0, 10)
  for (const t of (zoneTasks || []).filter((task) => {
    if (task.status === 'completed' || task.status === 'cancelled') return false
    if (!task.due_date) return false
    return String(task.due_date).slice(0, 10) <= today
  }).slice(0, 2)) {
    items.push({
      kind: 'task',
      label: t.title || 'Task due',
      severity: 'info',
      link: { path: `/zones/${zone.id}`, query: { tab: 'ops', ops: 'tasks' } },
    })
  }
  return items
}

function resolveHealth({ sensorsRollup, zoneAlerts, attention }) {
  const sev = (zoneAlerts[0]?.severity || '').toLowerCase()
  if (sev === 'critical' || sev === 'high') return 'alert'
  if (sensorsRollup.worst === 'attention' || sev === 'warning' || sev === 'medium') return 'warn'
  if (sensorsRollup.state === 'not_set_up' && sensorsRollup.unset === sensorsRollup.healthy + sensorsRollup.attention + sensorsRollup.unset) {
    return 'unconfigured'
  }
  if (attention.length) return 'warn'
  return 'ok'
}

function resolveGreenhouse({ zone, actuators, readings, sensors }) {
  if (String(zone?.zone_type || '').toLowerCase() !== 'greenhouse') return null
  const meta = parseMeta(zone.meta_data)
  const gc = meta.greenhouse_climate || {}
  const zoneActs = zoneActuators(actuators, zone.id)
  const vent = zoneActs.find((a) => Number(a.id) === Number(gc.vent_actuator_id))
  const shade = zoneActs.find((a) => Number(a.id) === Number(gc.shade_actuator_id))
  const tempSensor = (sensors || []).find((s) => s.sensor_type === 'temperature')
  const tempReading = tempSensor ? readings?.[tempSensor.id] : null
  const insideTemp = tempReading?.value_raw != null
    ? `${Number(tempReading.value_raw).toFixed(1)}°C inside`
    : null
  return {
    policy: gc.automation_policy || 'manual',
    ventState: vent ? String(vent.last_command || vent.desired_state || 'unknown') : null,
    shadeState: shade ? String(shade.last_command || shade.desired_state || 'unknown') : null,
    insideTemp,
  }
}

/**
 * @param {object} params
 */
export function computeZoneVisualStatus(params) {
  const {
    zone,
    sensors = [],
    readings = {},
    actuators = [],
    tasks = [],
    alerts = [],
    schedules = [],
    programs = [],
    cropCycles = [],
    fertigationEvents = [],
  } = params

  const zoneSensors = sensors.filter((s) => Number(s.zone_id) === Number(zone.id))
  const zoneAlerts = filterZoneAlerts(alerts, zoneSensors, zone.name)
  const zoneTasks = tasks.filter((t) => Number(t.zone_id) === Number(zone.id))

  const sensorStates = zoneSensors.map((s) => {
    const reading = readings[s.id]
    const linked = (alerts || []).some(
      (a) => !a.is_read && !a.is_acknowledged && sensorAlertLinked(a, s),
    )
    return classifySensorHardwareState(s, reading, linked)
  })

  const sensorsRollup = rollupSensorStates(sensorStates)

  // Prefer a specific attention label from humidity etc.
  let sensorSummary = sensorsRollup.summary
  const attentionSensor = zoneSensors.find((s, i) => sensorStates[i] === 'attention')
  if (attentionSensor && zoneAlerts.length) {
    const reading = readings[attentionSensor.id]
    if (attentionSensor.sensor_type === 'humidity' && reading?.value_raw != null) {
      sensorSummary = 'Humidity high'
    } else if (attentionSensor.sensor_type === 'temperature' && reading?.value_raw != null) {
      sensorSummary = 'Temperature out of range'
    }
  }

  const cycle = activeCropCycle(cropCycles, zone.id)
  const plants = cycle
    ? {
        state: 'growing',
        cropName: cycle.batch_label || cycle.name || 'Crop',
        stage: formatStage(cycle.current_stage),
        batchLabel: cycle.batch_label || '',
      }
    : { state: 'empty', cropName: null, stage: null, batchLabel: null, label: 'Empty — ready to plant' }

  const program = activeProgramForZone(programs, zone.id)
  const feedPlan = buildZoneFeedingPlan({
    zoneId: zone.id,
    activeProgram: program,
    programs,
    schedules,
    events: fertigationEvents,
    reservoirs: [],
    ecTargets: [],
    waterStatus: null,
  })
  const waterBase = resolveWaterKind(program, actuators, zone.id)
  const water = {
    ...waterBase,
    nextRun: feedPlan.nextRunLabel,
    lastEvent: feedPlan.lastEventSummary,
    scheduleLabel: feedPlan.nextRunLabel,
  }

  const light = resolveLight({ zone, schedules, actuators, programs })
  const attention = attentionItems({ zone, zoneSensors, zoneAlerts, zoneTasks })
  const health = resolveHealth({ sensorsRollup, zoneAlerts, attention })
  const greenhouse = resolveGreenhouse({ zone, actuators, readings, sensors: zoneSensors })

  return {
    plants,
    light,
    water,
    sensors: { ...sensorsRollup, summary: sensorSummary },
    attention,
    health,
    greenhouse,
  }
}

/**
 * @param {object} zone
 * @param {number} index
 */
export function defaultLayoutForZone(zone, index = 0) {
  const named = DEFAULT_ZONE_LAYOUTS_BY_NAME[zone?.name]
  if (named) return { ...named }
  const col = index % 4
  const row = Math.floor(index / 4)
  return {
    x: 0.04 + col * 0.22,
    y: 0.06 + row * 0.22,
    w: DEFAULT_TILE_W,
    h: DEFAULT_TILE_H,
  }
}

/**
 * Farmer-facing label for zone_type (Today tiles — not raw DB values).
 * @param {string|undefined|null} zoneType
 */
export function formatZoneTypeLabel(zoneType) {
  const t = String(zoneType || '').toLowerCase()
  if (!t) return 'Grow area'
  if (t.includes('greenhouse')) return 'Greenhouse'
  if (t.includes('outdoor')) return 'Outdoor bed'
  if (t.includes('indoor')) return 'Indoor grow area'
  if (t.includes('nursery')) return 'Nursery'
  if (t.includes('propagation')) return 'Propagation'
  return t.replace(/_/g, ' ')
}

/**
 * @param {object} zone
 * @param {(id:number)=>object|null} getLayout
 * @param {number} index
 */
export function resolveZoneLayout(zone, getLayout, index = 0) {
  const saved = getLayout?.(zone.id)
  if (saved && saved.x != null && saved.y != null) {
    return {
      x: Number(saved.x),
      y: Number(saved.y),
      w: Number(saved.w ?? DEFAULT_TILE_W),
      h: Number(saved.h ?? DEFAULT_TILE_H),
    }
  }
  return defaultLayoutForZone(zone, index)
}
