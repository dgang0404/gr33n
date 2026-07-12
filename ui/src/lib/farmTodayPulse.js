/**
 * Phase 176 — operational "farm pulse" cells for FarmSiteStrip.
 */

import { cronSortHourFromSchedule, scheduleRunsLabel } from './cronHumanize.js'
import { computeZoneVisualStatus } from './farmVisualStatus.js'
import { comfortRoute, feedWaterRoute } from './dashboardWorkspaceLinks.js'

function zoneName(zones, zoneId) {
  return zones.find((z) => Number(z.id) === Number(zoneId))?.name || 'Zone'
}

function stageBucket(stage) {
  const s = String(stage || '').toLowerCase()
  if (s.includes('bloom') || s.includes('flower')) return 'bloom'
  if (s.includes('prop') || s.includes('clone')) return 'propagate'
  if (s.includes('veg') || s.includes('early')) return 'veg'
  return 'other'
}

/**
 * @param {object} params
 */
export function resolveNextWaterCell({ zones = [], programs = [], schedules = [] }) {
  const active = (programs || []).filter(
    (p) => p.is_active !== false && p.schedule_id && p.target_zone_id,
  )
  const entries = active.map((program) => {
    const sched = (schedules || []).find((s) => Number(s.id) === Number(program.schedule_id))
    if (!sched || sched.is_active === false) return null
    const when = scheduleRunsLabel(sched)
    if (!when || when === 'No run scheduled') return null
    const name = zoneName(zones, program.target_zone_id)
    const hour = cronSortHourFromSchedule(sched)
    return { name, when, zoneId: program.target_zone_id, hour }
  }).filter(Boolean)

  if (!entries.length) return null
  entries.sort((a, b) => {
    if (a.hour !== b.hour) return a.hour - b.hour
    return String(a.name).localeCompare(String(b.name))
  })
  const first = entries[0]
  return {
    id: 'next_water',
    label: 'Next water',
    value: `${first.name} · ${first.when}`,
    link: feedWaterRoute(zones, first.zoneId),
  }
}

/**
 * @param {object} params
 */
export function resolveLightsCell({
  zones = [],
  schedules = [],
  actuators = [],
  programs = [],
  sensors = [],
  readings = {},
  tasks = [],
  alerts = [],
  cropCycles = [],
  fertigationEvents = [],
}) {
  const statusCtx = {
    sensors,
    readings,
    actuators,
    tasks,
    alerts,
    schedules,
    programs,
    cropCycles,
    fertigationEvents,
  }

  const onCount = zones.filter((zone) => {
    const status = computeZoneVisualStatus({ zone, ...statusCtx })
    return status.light?.state === 'on'
  }).length

  if (onCount > 0) {
    return {
      id: 'next_light',
      label: 'Lights',
      value: `${onCount} zone${onCount === 1 ? '' : 's'} on`,
      link: comfortRoute('schedules'),
    }
  }

  const lightSched = (schedules || []).find((s) => {
    if (s.is_active === false) return false
    const n = `${s.name || ''} ${s.description || ''}`.toLowerCase()
    return n.includes('light') || n.includes('18/6') || n.includes('12/12')
  })

  if (lightSched) {
    return {
      id: 'next_light',
      label: 'Lights',
      value: scheduleRunsLabel(lightSched),
      link: comfortRoute('schedules'),
    }
  }

  return null
}

/**
 * @param {object} params
 */
export function resolveGrowingCell({ cropCycles = [] }) {
  const active = (cropCycles || []).filter((c) => c.is_active !== false)
  if (!active.length) return null

  const bloom = active.filter((c) => stageBucket(c.current_stage) === 'bloom').length
  const parts = [`${active.length} run${active.length === 1 ? '' : 's'}`]
  if (bloom) parts.push(`${bloom} in bloom`)

  return {
    id: 'crops',
    label: 'Growing',
    value: parts.join(' · '),
    link: { path: '/zones' },
  }
}

/**
 * @param {object} params
 */
export function resolveDevicesCell({ devices = [], queueDepth = 0 }) {
  if (!devices.length) return null
  const online = devices.filter((d) => d.status === 'online').length
  let value = `${online} of ${devices.length} online`
  if (Number(queueDepth) > 0) value += ` · queue ${queueDepth}`
  return {
    id: 'edge',
    label: 'Devices',
    value,
    link: { path: '/hardware' },
  }
}

/**
 * @param {object} params
 */
export function buildFarmTodayPulse(params) {
  const cells = [
    resolveNextWaterCell(params),
    resolveLightsCell(params),
    resolveGrowingCell(params),
    resolveDevicesCell(params),
  ].filter(Boolean)

  return { cells }
}
