/**
 * Phase 40 WS3 — zone-scoped rules and schedules filtered by plant need.
 */

import { PLANT_NEEDS, sensorPlantNeed } from './plantNeeds.js'
import { ruleAppliesToZone, collectZoneScheduleIds } from './zoneGrowSummary.js'
import { scheduleRunsLabel } from './cronHumanize.js'

function parseTriggerConfig(rule) {
  try {
    const tc = rule?.trigger_configuration
    return typeof tc === 'string' ? JSON.parse(tc) : (tc || {})
  } catch {
    return {}
  }
}

function parseConditions(rule) {
  try {
    const c = rule?.conditions_jsonb
    return typeof c === 'string' ? JSON.parse(c) : (c || [])
  } catch {
    return []
  }
}

export function isGreenhouseRule(rule) {
  const name = String(rule?.name || '')
  return name.startsWith('GH —') || name.startsWith('GH -')
}

/**
 * @param {object} rule
 * @param {string} need
 * @param {object} ctx
 */
export function ruleAppliesToNeed(rule, need, ctx) {
  const {
    zoneId,
    zoneName,
    sensors = [],
  } = ctx
  const zoneSensorIds = new Set(sensors.map((s) => s.id))
  if (!ruleAppliesToZone(rule, zoneId, zoneName, zoneSensorIds)) return false

  if (isGreenhouseRule(rule)) return need === PLANT_NEEDS.air

  const tc = parseTriggerConfig(rule)
  const name = String(rule.name || '').toLowerCase()

  if (need === PLANT_NEEDS.light) {
    if (name.includes('light')) return true
    if (tc.target_zone && (tc.action || '').includes('actuator')) return true
  }

  if (need === PLANT_NEEDS.water) {
    if (name.includes('water') || name.includes('irrig') || name.includes('feed')) return true
  }

  for (const p of parseConditions(rule)) {
    const sid = p?.sensor_id
    if (sid == null) continue
    const sensor = sensors.find((s) => s.id === Number(sid))
    if (sensor && sensorPlantNeed(sensor.sensor_type) === need) return true
  }

  if (tc.sensor_id != null) {
    const sensor = sensors.find((s) => s.id === Number(tc.sensor_id))
    if (sensor && sensorPlantNeed(sensor.sensor_type) === need) return true
  }

  if (need === PLANT_NEEDS.air) {
    for (const p of parseConditions(rule)) {
      if (p?.type === 'setpoint' && p?.sensor_type) {
        if (sensorPlantNeed(p.sensor_type) === PLANT_NEEDS.air) return true
      }
    }
    if (tc.zone_id != null && Number(tc.zone_id) === zoneId) return true
  }

  return false
}

/**
 * @param {object} schedule
 * @param {string} need
 * @param {object} ctx
 */
export function scheduleAppliesToNeed(schedule, need, ctx) {
  const zoneScheduleIds = collectZoneScheduleIds(ctx)
  if (!zoneScheduleIds.includes(schedule.id)) return false

  const st = String(schedule.schedule_type || '').toLowerCase()
  const { activeProgram, lightingPrograms = [], zoneId } = ctx

  if (need === PLANT_NEEDS.water) {
    return st === 'irrigation' || activeProgram?.schedule_id === schedule.id
  }
  if (need === PLANT_NEEDS.light) {
    if (st === 'lighting') return true
    return lightingPrograms.some(
      (lp) => lp.zone_id === zoneId
        && (lp.schedule_on_id === schedule.id || lp.schedule_off_id === schedule.id),
    )
  }
  if (need === PLANT_NEEDS.air) {
    return st !== 'lighting' && st !== 'irrigation'
  }
  return false
}

export function linkedNameForSchedule(schedule, ctx) {
  const { activeProgram, lightingPrograms = [], zoneId } = ctx
  if (activeProgram?.schedule_id === schedule.id) return activeProgram.name
  const lp = lightingPrograms.find(
    (p) => p.zone_id === zoneId
      && (p.schedule_on_id === schedule.id || p.schedule_off_id === schedule.id),
  )
  return lp?.name || null
}

/**
 * @param {object} ctx
 * @returns {{ rules: object[], schedules: object[] }}
 */
export function zoneAutomationForNeed(need, ctx) {
  const rules = (ctx.rules || []).filter((r) => ruleAppliesToNeed(r, need, ctx))
  const schedules = (ctx.schedules || [])
    .filter((s) => scheduleAppliesToNeed(s, need, ctx))
    .map((s) => ({
      schedule: s,
      runsLabel: scheduleRunsLabel(s),
      linkedName: linkedNameForSchedule(s, ctx),
    }))
  return { rules, schedules }
}

/**
 * Interlock badges for GH climate rules (Phase 36).
 * @param {object} rule
 * @param {object[]} sensors
 */
export function greenhouseRuleBadges(rule, sensors) {
  if (!isGreenhouseRule(rule)) return []
  const badges = []
  const name = String(rule.name || '').toLowerCase()
  const types = new Set(sensors.map((s) => String(s.sensor_type || '').toLowerCase()))
  const hasLux = ['lux', 'par', 'par_umol', 'ppfd', 'light_level'].some((t) => types.has(t))
  const hasTemp = ['temperature', 'temp', 'air_temp'].some((t) => types.has(t))

  if ((name.includes('lux') || name.includes('shade') || name.includes('par')) && !hasLux) {
    badges.push({ id: 'no-lux', label: 'No lux sensor', tone: 'warn' })
  }
  if ((name.includes('temp') || name.includes('fan') || name.includes('vent') || name.includes('heat')) && !hasTemp) {
    badges.push({ id: 'no-temp', label: 'No temp sensor', tone: 'warn' })
  }
  return badges
}

export function formatRuleLastFired(iso) {
  if (!iso) return 'Never fired'
  try {
    return new Date(iso).toLocaleString(undefined, {
      month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit',
    })
  } catch {
    return iso
  }
}
