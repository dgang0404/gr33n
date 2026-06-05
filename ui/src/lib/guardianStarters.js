/**
 * Phase 40 WS7b + Phase 47 WS6 — zone and feeding hub Guardian starter chips.
 */

import { buildZoneGuardianPrompt, buildZoneGuardianContextRef } from './guardianContextPrompts.js'

const FEEDING_STARTER_IDS = ['next-feed', 'run-feed-safe', 'water-only']

/**
 * @param {'zone_overview'|'zone_water'|'zone_light'|'zone_climate'} surface
 * @param {object} ctx — zone snapshot (see guardianContextPrompts)
 * @returns {Array<{ id: string, label: string, message: string, contextRef?: object }>}
 */
export function buildZoneStarters(surface, ctx) {
  const zoneName = ctx.zone?.name || 'this room'
  const starters = []
  const baseRef = () => buildZoneGuardianContextRef({ ...ctx, activeTab: tabForSurface(surface) })

  if (surface === 'zone_water') {
    return buildWaterTabStarters(ctx, baseRef).slice(0, 5)
  }

  const alerts = ctx.unreadAlerts || []
  if (alerts.length) {
    const a = alerts[0]
    starters.push({
      id: 'explain-alert',
      label: 'Explain latest alert',
      message: `Explain alert #${a.id} (${a.subject_rendered || 'alert'}) for ${zoneName} and what I should do next.`,
      contextRef: { ...baseRef(), alert_id: a.id },
    })
  }

  if (ctx.nextSchedule?.schedule?.name) {
    starters.push({
      id: 'today-schedule',
      label: 'What runs when',
      message: `What runs today in ${zoneName} for "${ctx.nextSchedule.schedule.name}"?`,
      contextRef: baseRef(),
    })
  }

  if (surface === 'zone_overview') {
    if (ctx.queueDepth > 0) {
      starters.push({
        id: 'queue-safety',
        label: 'Check command queue',
        message: `What's queued for devices in ${zoneName} and is it safe to run another pulse?`,
        contextRef: baseRef(),
      })
    } else if (ctx.activeProgramName) {
      starters.push(...buildWaterTabStarters(ctx, baseRef).filter((s) => s.id !== 'next-feed').slice(0, 2))
    }
  }

  if ((ctx.missingComfortTargets || 0) > 0 && (surface === 'zone_climate' || surface === 'zone_overview')) {
    starters.push({
      id: 'comfort-targets',
      label: 'Set comfort targets',
      message: `What humidity and temperature comfort targets should I set for ${zoneName} at my current crop stage?`,
      contextRef: baseRef(),
    })
  }

  if (surface === 'zone_climate' && ctx.activeRulesCount > 0) {
    starters.push({
      id: 'climate-rules',
      label: 'Explain automations',
      message: `Explain the ${ctx.activeRulesCount} active automation(s) affecting climate in ${zoneName} and when they run.`,
      contextRef: baseRef(),
    })
  }

  if (!starters.length) {
    starters.push({
      id: 'summarize-room',
      label: 'Summarize this room',
      message: buildZoneGuardianPrompt(ctx),
      contextRef: baseRef(),
    })
  }

  return dedupeStarters(starters).slice(0, 5)
}

/**
 * Phase 47 WS6 — farm Feed & water hub starters.
 * @param {object} params
 * @param {object[]} params.zones
 * @param {number|null} params.zoneContextId
 * @param {string} [params.zoneName]
 */
export function buildFeedingHubStarters({ zones = [], zoneContextId = null, zoneName = '' }) {
  const focusZone = zoneContextId
    ? zones.find((z) => Number(z.id) === Number(zoneContextId))
    : zones[0]
  const name = focusZone?.name || zoneName || 'this farm'
  const routeRef = { type: 'route', path: '/feeding', name: 'Feed & water' }

  if (focusZone) {
    const ctx = { zone: focusZone, activeTab: 'water' }
    return buildWaterTabStarters(ctx, () => buildZoneGuardianContextRef(ctx))
  }

  return [{
    id: 'farm-feeding-overview',
    label: 'Feeding overview',
    message: 'Which rooms have feeding plans today, and which need reservoir top-up or attention?',
    contextRef: routeRef,
  }]
}

function buildWaterTabStarters(ctx, baseRef) {
  const zoneName = ctx.zone?.name || 'this room'
  const ref = baseRef()
  const programHint = ctx.activeProgramName ? ` (plan: "${ctx.activeProgramName}")` : ''

  return [
    {
      id: 'next-feed',
      label: 'Next feed',
      message: `When is the next feed for ${zoneName}${programHint}? Include last feed and reservoir status.`,
      contextRef: ref,
    },
    {
      id: 'run-feed-safe',
      label: 'Run feed now?',
      message: `Is it safe to run a feed now in ${zoneName}? Check the Pi queue and whether mixing is still required.`,
      contextRef: ref,
    },
    {
      id: 'water-only',
      label: 'Water only',
      message: `Switch ${zoneName} to plain water-only irrigation (no nutrients) and explain what would change in the feeding plan.`,
      contextRef: ref,
    },
  ]
}

function dedupeStarters(starters) {
  const seen = new Set()
  return starters.filter((s) => {
    if (seen.has(s.id)) return false
    seen.add(s.id)
    return true
  })
}

function tabForSurface(surface) {
  if (surface === 'zone_water') return 'water'
  if (surface === 'zone_light') return 'light'
  if (surface === 'zone_climate') return 'climate'
  return 'overview'
}

export { FEEDING_STARTER_IDS }
