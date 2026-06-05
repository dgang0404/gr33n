/**
 * Phase 41 WS3 — zone query-param context for farm-wide pages.
 */

import { alertMatchesZone, ruleAppliesToZone, collectZoneScheduleIds } from './zoneGrowSummary.js'

/** @param {string|string[]|undefined|null} raw */
export function parseZoneIdQuery(raw) {
  if (raw == null) return null
  const s = Array.isArray(raw) ? raw[0] : raw
  const n = Number(s)
  return Number.isFinite(n) && n > 0 ? n : null
}

/**
 * @param {object} schedule
 * @param {object} ctx
 */
export function scheduleAppliesToZone(schedule, ctx) {
  const {
    zoneId,
    zoneName = '',
    programs = [],
    lightingPrograms = [],
    tasks = [],
  } = ctx

  const ids = collectZoneScheduleIds({
    zoneId,
    zoneName,
    schedules: [schedule],
    activeProgram: programs.find((p) => p.is_active && Number(p.target_zone_id) === zoneId),
    lightingPrograms,
  })
  if (ids.includes(schedule.id)) return true

  const name = String(zoneName || '').trim()
  const desc = `${schedule.description || ''} ${schedule.name || ''}`
  if (name && desc.includes(name)) return true

  if (programs.some((p) => Number(p.schedule_id) === schedule.id && Number(p.target_zone_id) === zoneId)) {
    return true
  }
  if (lightingPrograms.some(
    (lp) => lp.zone_id === zoneId
      && (Number(lp.schedule_on_id) === schedule.id || Number(lp.schedule_off_id) === schedule.id),
  )) {
    return true
  }
  if (tasks.some((t) => Number(t.schedule_id) === schedule.id && Number(t.zone_id) === zoneId)) {
    return true
  }
  return false
}

export function filterSchedulesForZone(schedules, zoneId, zoneName, programs, lightingPrograms, tasks) {
  const ctx = { zoneId, zoneName, programs, lightingPrograms, tasks }
  return (schedules || []).filter((s) => scheduleAppliesToZone(s, ctx))
}

export function filterAlertsForZone(alerts, zoneId, zoneName, sensors) {
  return (alerts || []).filter((a) => alertMatchesZone(a, sensors || [], zoneName))
}

export function filterRulesForZone(rules, zoneId, zoneName, sensors) {
  const sensorIds = new Set((sensors || []).filter((s) => Number(s.zone_id) === zoneId).map((s) => s.id))
  return (rules || []).filter((r) => ruleAppliesToZone(r, zoneId, zoneName, sensorIds))
}

export function programAppliesToZone(program, zoneId, cropCycles = []) {
  if (Number(program?.target_zone_id) === zoneId) return true
  return (cropCycles || []).some(
    (c) => Number(c.zone_id) === zoneId && Number(c.program_id) === Number(program?.id),
  )
}

export function filterProgramsForZone(programs, zoneId, cropCycles = []) {
  return (programs || []).filter((p) => programAppliesToZone(p, zoneId, cropCycles))
}
