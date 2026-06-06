/**
 * Phase 47 WS4 — farm-wide feeding hub cards (one per zone).
 */

import { buildZoneFeedingPlan } from './zoneFeedingPlan.js'

/**
 * @param {object[]} programs
 * @param {number} zoneId
 */
export function activeProgramForZone(programs, zoneId) {
  const forZone = (programs || []).filter((p) => Number(p.target_zone_id) === Number(zoneId))
  return forZone.find((p) => p.is_active) || forZone[0] || null
}

/**
 * @param {object} plan
 * @returns {{ level: string, label: string }|null}
 */
export function detectFeedingAttention(plan) {
  if (!plan?.hasPlan) {
    return { level: 'muted', label: 'No feeding plan' }
  }
  if (plan.reservoirTone === 'warn') {
    return { level: 'warn', label: 'Reservoir needs attention' }
  }
  if (plan.schedule && !plan.scheduleActive) {
    return { level: 'warn', label: 'Feeding paused' }
  }
  const ecAfter = plan.lastEvent?.ec_after_mscm
  if (ecAfter != null && plan.ecRange?.min != null && !plan.irrigationOnly) {
    if (Number(ecAfter) < Number(plan.ecRange.min)) {
      return { level: 'warn', label: 'Last feed below target' }
    }
  }
  return null
}

/**
 * @param {object} params
 * @returns {Array<{ zone: object, plan: object, attention: object|null }>}
 */
export function buildFarmFeedingCards({
  zones = [],
  programs = [],
  schedules = [],
  events = [],
  ecTargets = [],
  reservoirs = [],
}) {
  return (zones || []).map((zone) => {
    const activeProgram = activeProgramForZone(programs, zone.id)
    const plan = buildZoneFeedingPlan({
      zoneId: zone.id,
      activeProgram,
      programs,
      schedules,
      events,
      ecTargets,
      reservoirs,
    })
    return {
      zone,
      plan,
      attention: detectFeedingAttention(plan),
    }
  })
}

/**
 * @param {Array} cards
 * @param {number|null} zoneId
 */
export function filterFeedingCardsByZone(cards, zoneId) {
  if (zoneId == null) return cards
  return (cards || []).filter((c) => Number(c.zone.id) === Number(zoneId))
}

/**
 * @param {object[]} programs
 * @param {object[]} zones
 */
export function countZonesWithFeedingPlan(programs, zones) {
  return (zones || []).filter((z) => activeProgramForZone(programs, z.id)).length
}

/** @deprecated Use countZonesWithFeedingPlan */
export const countRoomsWithFeedingPlan = countZonesWithFeedingPlan
