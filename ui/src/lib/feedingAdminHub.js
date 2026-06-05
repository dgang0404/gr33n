/**
 * Phase 43 WS3 — farm-wide feeding admin hub (programs, reservoirs, EC targets).
 */

import { scheduleRunsLabel } from './cronHumanize.js'

/**
 * @param {string} stage
 */
export function formatGrowthStage(stage) {
  if (!stage) return '—'
  return String(stage).replace(/_/g, ' ')
}

/**
 * @param {object} reservoir
 */
export function reservoirFillPct(reservoir) {
  const cap = Number(reservoir?.capacity_liters)
  const vol = Number(reservoir?.current_volume_liters)
  if (!Number.isFinite(cap) || cap <= 0 || !Number.isFinite(vol)) return 0
  return Math.min(100, Math.round((vol / cap) * 100))
}

/**
 * Farmer-facing reservoir status (ready / needs top-up).
 * @param {object} reservoir
 */
export function reservoirStatusLabel(reservoir) {
  const s = String(reservoir?.status || '').toLowerCase().replace(/\s+/g, '_')
  if (s === 'ready') return { label: 'Ready', tone: 'ok' }
  if (s === 'needs_top_up' || s === 'empty') return { label: 'Needs top-up', tone: 'warn' }
  if (s === 'mixing') return { label: 'Mixing', tone: 'muted' }
  if (s === 'offline') return { label: 'Offline', tone: 'warn' }
  if (s) return { label: s.replace(/_/g, ' '), tone: 'muted' }
  return { label: 'Unknown', tone: 'muted' }
}

/**
 * @param {object} program
 * @param {object[]} zones
 * @param {object[]} schedules
 */
export function buildProgramAdminCard(program, zones, schedules) {
  const zone = (zones || []).find((z) => Number(z.id) === Number(program.target_zone_id))
  const schedule = program.schedule_id
    ? (schedules || []).find((s) => Number(s.id) === Number(program.schedule_id))
    : null
  return {
    id: program.id,
    name: program.name,
    zoneName: zone?.name || 'No room linked',
    zoneId: program.target_zone_id,
    nextRunLabel: schedule
      ? scheduleRunsLabel(schedule)
      : (program.is_active ? 'On demand' : 'No schedule'),
    irrigationOnly: Boolean(program.irrigation_only),
    isActive: Boolean(program.is_active),
    scheduleActive: schedule?.is_active !== false,
    volumeLiters: program.total_volume_liters,
  }
}

/**
 * @param {object[]} programs
 * @param {object[]} zones
 * @param {object[]} schedules
 */
export function buildProgramAdminCards(programs, zones, schedules) {
  return (programs || [])
    .map((p) => buildProgramAdminCard(p, zones, schedules))
    .sort((a, b) => {
      if (a.isActive !== b.isActive) return a.isActive ? -1 : 1
      return a.zoneName.localeCompare(b.zoneName)
    })
}

/**
 * @param {object} reservoir
 * @param {object[]} zones
 */
export function buildReservoirAdminCard(reservoir, zones) {
  const zone = (zones || []).find((z) => Number(z.id) === Number(reservoir.zone_id))
  const status = reservoirStatusLabel(reservoir)
  return {
    id: reservoir.id,
    name: reservoir.name,
    zoneName: zone?.name || 'Farm-wide',
    zoneId: reservoir.zone_id,
    currentLiters: reservoir.current_volume_liters,
    capacityLiters: reservoir.capacity_liters,
    fillPct: reservoirFillPct(reservoir),
    statusLabel: status.label,
    statusTone: status.tone,
    ec: reservoir.last_ec_mscm,
    ph: reservoir.last_ph,
  }
}

/**
 * @param {object[]} reservoirs
 * @param {object[]} zones
 */
export function buildReservoirAdminCards(reservoirs, zones) {
  return (reservoirs || [])
    .map((r) => buildReservoirAdminCard(r, zones))
    .sort((a, b) => a.name.localeCompare(b.name))
}

/**
 * @param {object} target
 * @param {object[]} zones
 */
export function buildEcTargetAdminCard(target, zones) {
  const zone = (zones || []).find((z) => Number(z.id) === Number(target.zone_id))
  return {
    id: target.id,
    stageLabel: formatGrowthStage(target.growth_stage),
    zoneName: zone?.name || 'All rooms',
    zoneId: target.zone_id,
    ecRange: `${target.ec_min_mscm}–${target.ec_max_mscm} mS/cm`,
    phRange: target.ph_min != null && target.ph_max != null
      ? `pH ${target.ph_min}–${target.ph_max}`
      : null,
    notes: target.notes || '',
  }
}

/**
 * @param {object[]} ecTargets
 * @param {object[]} zones
 */
export function buildEcTargetAdminCards(ecTargets, zones) {
  return (ecTargets || [])
    .map((t) => buildEcTargetAdminCard(t, zones))
    .sort((a, b) => a.stageLabel.localeCompare(b.stageLabel))
}

/**
 * @param {object[]} reservoirs
 * @param {number|null} zoneId
 */
export function filterReservoirsForZone(reservoirs, zoneId) {
  if (zoneId == null) return reservoirs || []
  return (reservoirs || []).filter(
    (r) => r.zone_id == null || Number(r.zone_id) === Number(zoneId),
  )
}

/**
 * @param {object[]} ecTargets
 * @param {number|null} zoneId
 */
export function filterEcTargetsForZone(ecTargets, zoneId) {
  if (zoneId == null) return ecTargets || []
  return (ecTargets || []).filter(
    (t) => t.zone_id == null || Number(t.zone_id) === Number(zoneId),
  )
}
