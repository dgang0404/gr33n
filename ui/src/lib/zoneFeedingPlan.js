/**
 * Phase 47 WS1 — one feeding plan object per zone (client-side compose).
 */

import { scheduleRunsLabel } from './cronHumanize.js'
import {
  lastZoneFeedEvent,
  formatLastFeedLine,
  formatEdgeQueueLine,
} from './zoneWaterGrowStory.js'

/**
 * @param {object|null} program
 * @param {object[]} ecTargets
 */
export function resolveEcRange(program, ecTargets = []) {
  if (!program) return null

  const linked = program.ec_target_id
    ? ecTargets.find((t) => Number(t.id) === Number(program.ec_target_id))
    : null
  if (linked?.ec_min_mscm != null && linked?.ec_max_mscm != null) {
    return {
      min: linked.ec_min_mscm,
      max: linked.ec_max_mscm,
      label: `${linked.ec_min_mscm}–${linked.ec_max_mscm} mS/cm`,
    }
  }

  const min = program.ec_trigger_low
  const max = program.ec_trigger_high
  if (min != null && max != null) {
    return { min, max, label: `${min}–${max} mS/cm` }
  }
  if (min != null) return { min, max: null, label: `≥ ${min} mS/cm` }
  if (max != null) return { min: null, max, label: `≤ ${max} mS/cm` }
  return null
}

/**
 * @param {object|null} program
 * @param {object[]} reservoirs
 */
export function resolveReservoir(program, reservoirs = []) {
  if (!program?.reservoir_id) return null
  return reservoirs.find((r) => Number(r.id) === Number(program.reservoir_id)) || null
}

/**
 * @param {object|null} reservoir
 */
export function mapReservoirStatus(reservoir) {
  if (!reservoir) {
    return { status: 'unknown', label: 'No reservoir linked', tone: 'muted' }
  }
  const s = String(reservoir.status || '').toLowerCase().replace(/\s+/g, '_')
  const name = reservoir.name || 'Reservoir'
  if (s === 'ready') return { status: 'ready', label: `${name} · ready`, tone: 'ok' }
  if (s === 'needs_top_up' || s === 'empty') {
    return { status: 'low', label: `${name} · needs top-up`, tone: 'warn' }
  }
  if (s === 'offline') return { status: 'offline', label: `${name} · offline`, tone: 'warn' }
  return { status: s || 'unknown', label: name, tone: 'muted' }
}

/**
 * @param {object|null} event
 */
export function formatLastEventSummary(event) {
  if (!event) return null
  const vol = event.volume_applied_liters != null ? `${event.volume_applied_liters}L` : null
  let ecPart = null
  if (event.ec_after_mscm != null) ecPart = `EC ${event.ec_after_mscm}`
  const ok = event.ec_after_mscm != null || event.volume_applied_liters != null ? 'OK' : null
  const parts = [vol, ecPart, ok].filter(Boolean)
  if (parts.length) return `Fed ${parts.join(' · ')}`
  return null
}

/**
 * @param {object} plan
 */
export function buildFeedingPlanStatusLine(plan) {
  if (!plan.hasPlan) return 'No feeding plan for this room yet'
  const parts = []
  if (plan.nextRunLabel) parts.push(`Next feed: ${plan.nextRunLabel}`)
  if (plan.volumeLiters != null) parts.push(`${plan.volumeLiters}L`)
  if (plan.ecRange?.label) {
    parts.push(`EC ${plan.ecRange.label.replace(/\s*mS\/cm$/, '')}`)
  }
  return parts.join(' · ') || plan.programName
}

/**
 * @param {object} ctx
 */
export function buildZoneFeedingPlan(ctx) {
  const {
    zoneId,
    activeProgram = null,
    programs = [],
    schedules = [],
    events = [],
    ecTargets = [],
    reservoirs = [],
    waterStatus = null,
    queueHead = null,
  } = ctx

  const lastEvent = lastZoneFeedEvent(events, zoneId)
  const programName = activeProgram?.name || ''
  const schedule = activeProgram?.schedule_id
    ? schedules.find((s) => Number(s.id) === Number(activeProgram.schedule_id))
    : null

  const nextRunLabel = activeProgram
    ? (schedule ? scheduleRunsLabel(schedule) : 'On demand')
    : null

  const ecRange = resolveEcRange(activeProgram, ecTargets)
  const reservoir = resolveReservoir(activeProgram, reservoirs)
  const reservoirInfo = mapReservoirStatus(reservoir)
  const queueDepth = waterStatus?.queue_depth ?? 0

  const plan = {
    hasPlan: Boolean(activeProgram),
    programId: activeProgram?.id ?? null,
    programName,
    irrigationOnly: Boolean(activeProgram?.irrigation_only),
    scheduleActive: schedule?.is_active !== false,
    nextRunLabel,
    lastEventAt: lastEvent?.applied_at ?? null,
    lastEventSummary:
      formatLastEventSummary(lastEvent)
      || formatLastFeedLine(lastEvent, programName),
    volumeLiters: activeProgram?.total_volume_liters ?? null,
    runDurationSeconds: activeProgram?.run_duration_seconds ?? null,
    ecRange,
    reservoir,
    reservoirStatus: reservoirInfo.status,
    reservoirLabel: reservoirInfo.label,
    reservoirTone: reservoirInfo.tone,
    queueDepth,
    queueLine: formatEdgeQueueLine(queueDepth, queueHead),
    mixRequired: Boolean(waterStatus?.mix_required),
    activeProgram,
    schedule,
    lastEvent,
  }

  plan.statusLine = buildFeedingPlanStatusLine(plan)
  return plan
}
