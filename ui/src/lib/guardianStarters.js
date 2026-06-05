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

function targetsRouteRef(zoneContextId, tab = 'bands', surface = 'comfort_hub') {
  const path = zoneContextId
    ? `/comfort-targets?zone_id=${zoneContextId}${tab !== 'bands' ? `&tab=${tab}` : ''}`
    : `/comfort-targets${tab !== 'bands' ? `?tab=${tab}` : ''}`
  return {
    type: 'route',
    path,
    name: 'Targets & schedules',
    surface,
  }
}

function focusZoneName(zones, zoneContextId, zoneName = '') {
  const focusZone = zoneContextId
    ? zones.find((z) => Number(z.id) === Number(zoneContextId))
    : zones[0]
  return focusZone?.name || zoneName || 'this farm'
}

function stageLabelForZone(activeCycles, zoneId) {
  const cycle = (activeCycles || []).find((c) => Number(c.zone_id) === Number(zoneId))
  if (!cycle?.current_stage) return ''
  return String(cycle.current_stage).replace(/_/g, ' ')
}

function isGreenhouseRule(rule) {
  const n = String(rule?.name || '').toLowerCase()
  return rule?.is_active && (n.includes('gh') || n.includes('shade') || n.includes('greenhouse'))
}

function alertMentionsComfort(alert) {
  const text = `${alert?.subject_rendered || ''} ${alert?.message_text_rendered || ''}`.toLowerCase()
  return text.includes('setpoint') || text.includes('threshold') || text.includes('comfort') || text.includes('humidity') || text.includes('temperature')
}

/**
 * Phase 42 WS8 — comfort targets hub starter chips.
 * @param {object} params
 */
export function buildComfortHubStarters({
  zones = [],
  zoneContextId = null,
  zoneName = '',
  cards = [],
  rules = [],
  programs = [],
  schedules = [],
  alerts = [],
  activeCycles = [],
  surface = 'comfort_hub',
}) {
  const name = focusZoneName(zones, zoneContextId, zoneName)
  const zoneId = zoneContextId || zones[0]?.id
  const card = cards.find((c) => Number(c.zone.id) === Number(zoneId)) || cards[0]
  const routeRef = targetsRouteRef(zoneContextId, 'bands', surface)
  const stage = stageLabelForZone(activeCycles, zoneId)
  const stageBit = stage ? ` at ${stage} stage` : ''
  const starters = []

  if (card?.bands?.some((b) => (b.sensorType === 'humidity' || b.sensorType === 'rh') && b.status === 'missing')) {
    starters.push({
      id: 'set-humidity-band',
      label: 'Set humidity band',
      message: `Help me set a humidity comfort band for ${name}${stageBit}.`,
      contextRef: routeRef,
    })
  }

  const ghRule = rules.find((r) => isGreenhouseRule(r))
  if (ghRule) {
    starters.push({
      id: 'pause-shade-rule',
      label: 'Pause shade automation',
      message: `Disable the greenhouse shade rule for ${name} until I turn it back on`,
      contextRef: routeRef,
    })
  }

  const comfortAlert = (alerts || []).find((a) => !a.is_read && !a.is_acknowledged && alertMentionsComfort(a))
  if (comfortAlert) {
    starters.push({
      id: 'explain-comfort-alert',
      label: 'Explain this alert',
      message: `Explain alert #${comfortAlert.id} and whether I should change my comfort targets`,
      contextRef: { ...routeRef, alert_id: comfortAlert.id },
    })
  }

  const linkedProgram = programs.find((p) => p.is_active && p.schedule_id && (
    zoneId == null || Number(p.target_zone_id) === Number(zoneId)
  ))
  if (linkedProgram) {
    starters.push({
      id: 'next-feed-schedule',
      label: 'When does feeding run?',
      message: `When does the fertigation schedule for ${name} run next, in plain language?`,
      contextRef: routeRef,
    })
  } else if (schedules.some((s) => s.is_active)) {
    starters.push({
      id: 'next-schedule-run',
      label: 'What runs when?',
      message: `What runs next on the active schedules for ${name}?`,
      contextRef: targetsRouteRef(zoneContextId, 'schedules', 'schedules_farmer'),
    })
  }

  if (card?.bands?.some((b) => b.status === 'missing')) {
    starters.push({
      id: 'missing-targets',
      label: 'What should I fix?',
      message: `What comfort targets am I missing for ${name}?`,
      contextRef: routeRef,
    })
  }

  if (card?.bands?.some((b) => b.status === 'out_of_range')) {
    starters.push({
      id: 'out-of-range',
      label: 'Fix out-of-range',
      message: `Which comfort bands are out of range in ${name}, and what should I change?`,
      contextRef: routeRef,
    })
  }

  if (!starters.length) {
    starters.push({
      id: 'targets-overview',
      label: 'Explain my targets',
      message: zoneContextId
        ? `Summarize comfort bands and scheduled runs for ${name} in plain language.`
        : 'Which rooms have missing comfort bands or paused schedules I should fix today?',
      contextRef: routeRef,
    })
  }

  return dedupeStarters(starters).slice(0, surface === 'comfort_hub_zone' ? 3 : 4)
}

/**
 * Phase 42 WS8 — farmer schedules tab starters.
 */
export function buildSchedulesFarmerStarters({
  zones = [],
  zoneContextId = null,
  zoneName = '',
  schedules = [],
}) {
  const name = focusZoneName(zones, zoneContextId, zoneName)
  const routeRef = targetsRouteRef(zoneContextId, 'schedules', 'schedules_farmer')
  const starters = [
    {
      id: 'next-run-plain',
      label: 'Next run',
      message: `When does the next scheduled run for ${name} happen, in plain language?`,
      contextRef: routeRef,
    },
  ]
  if (schedules.some((s) => s.is_active)) {
    starters.push({
      id: 'pause-schedule-chat',
      label: 'Pause a schedule',
      message: `Help me pause the right schedule for ${name} without deleting it`,
      contextRef: routeRef,
    })
  }
  starters.push({
    id: 'schedules-overview',
    label: 'Explain schedules',
    message: `Which schedules affect ${name} and what does each one do?`,
    contextRef: routeRef,
  })
  return dedupeStarters(starters).slice(0, 3)
}

/**
 * Phase 42 WS8 — farmer automation rules tab starters.
 */
export function buildRulesFarmerStarters({
  zones = [],
  zoneContextId = null,
  zoneName = '',
  rules = [],
}) {
  const name = focusZoneName(zones, zoneContextId, zoneName)
  const routeRef = targetsRouteRef(zoneContextId, 'rules', 'rules_farmer')
  const starters = []
  const active = rules.filter((r) => r.is_active)
  if (active.length) {
    starters.push({
      id: 'explain-rules',
      label: 'Explain automations',
      message: `Explain the ${active.length} active automation rule(s) affecting ${name} and when they run.`,
      contextRef: routeRef,
    })
    const gh = active.find((r) => isGreenhouseRule(r))
    if (gh) {
      starters.push({
        id: 'pause-shade-chat',
        label: 'Pause shade rule',
        message: `Disable the greenhouse shade rule for ${name} until I turn it back on`,
        contextRef: routeRef,
      })
    }
  }
  starters.push({
    id: 'rules-safety',
    label: 'Safe to change?',
    message: `If I pause automations in ${name}, what climate control stops running?`,
    contextRef: routeRef,
  })
  return dedupeStarters(starters).slice(0, 3)
}
