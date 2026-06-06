/**
 * Phase 40 WS7b + Phase 47 WS6 — zone and feeding hub Guardian starter chips.
 * Phase 44 WS4 — setup-mode starters for wizards and onboarding drawer.
 */

const SETUP_SURFACE_MAX = {
  first_run_dashboard: 3,
  farm_setup_wizard: 2,
  zone_wizard: 3,
  device_wizard: 2,
  empty_zone_grow: 3,
  setup_mode_chat: 4,
}

function setupRouteRef(farmId, kind, surface) {
  const paths = {
    farm: `/farms/${farmId}/setup`,
    zone: `/farms/${farmId}/zones/new`,
    device: `/farms/${farmId}/devices/new`,
  }
  const names = {
    farm: 'Farm setup',
    zone: 'Add grow room',
    device: 'Connect edge device',
  }
  return { type: 'route', path: paths[kind], name: names[kind], surface }
}

/**
 * Phase 44 WS4 — setup / first-run Guardian starter chips (priority order per spec §4.2).
 * @param {object} params
 */
export function buildSetupStarters({
  surface = 'setup_mode_chat',
  farmId = null,
  zoneCount = 0,
  zones = [],
  activeCycles = [],
  unreadAlerts = [],
  deviceOffline = false,
  zoneName = '',
  deviceWizardStep = false,
} = {}) {
  const max = SETUP_SURFACE_MAX[surface] ?? 4
  const candidates = []

  if (zoneCount === 0) {
    candidates.push({
      id: 'first-grow-room',
      label: 'Add my first grow room',
      message: "I'm setting up a new farm — what should I do first after creating a zone?",
      contextRef: farmId
        ? setupRouteRef(farmId, 'zone', surface)
        : { type: 'route', path: '/zones', name: 'Zones', surface },
      setupMode: true,
    })
  }

  const firstZone = zones[0]
  const zoneForGrow = zoneName || firstZone?.name || ''
  const zoneId = firstZone?.id
  const hasCycleInZone = zoneId != null
    && (activeCycles || []).some((c) => Number(c.zone_id) === Number(zoneId))
  const growSurfaces = new Set(['setup_mode_chat', 'empty_zone_grow', 'zone_wizard'])
  if (zoneCount > 0 && zoneForGrow && !hasCycleInZone && growSurfaces.has(surface)) {
    candidates.push({
      id: 'start-grow',
      label: `Start a grow in ${zoneForGrow}`,
      message: `Add my philodendron to ${zoneForGrow} with a light fertigation program`,
      contextRef: zoneId
        ? { type: 'zone', id: zoneId, name: zoneForGrow }
        : (farmId ? setupRouteRef(farmId, 'zone', surface) : null),
      setupMode: true,
    })
  }

  const alert = (unreadAlerts || []).find((a) => !a.is_read) || unreadAlerts[0]
  if (alert?.id) {
    candidates.push({
      id: 'handle-alert',
      label: 'Handle this alert',
      message: `Acknowledge alert #${alert.id}: ${alert.subject_rendered || alert.subject || 'alert'}`,
      contextRef: { type: 'alert', id: alert.id },
      setupMode: true,
    })
  }

  if (surface === 'device_wizard' || deviceWizardStep) {
    candidates.push({
      id: 'wire-pi',
      label: 'Wire Pi checklist',
      message: 'start procedure wire-pi-relay-light',
      contextRef: farmId ? setupRouteRef(farmId, 'device', surface) : null,
      setupMode: true,
    })
  }

  if (surface === 'farm_setup_wizard') {
    candidates.push({
      id: 'compare-templates',
      label: 'Compare templates',
      message: "What's the difference between indoor veg and greenhouse climate bootstrap templates?",
      contextRef: farmId ? setupRouteRef(farmId, 'farm', surface) : null,
      setupMode: true,
    })
  }

  if (deviceOffline) {
    candidates.push({
      id: 'pi-offline',
      label: 'Why is my Pi offline?',
      message: 'start procedure diagnose-pi-offline',
      contextRef: farmId ? setupRouteRef(farmId, 'device', surface) : null,
      setupMode: true,
    })
  }

  candidates.push({
    id: 'setup-walkthrough',
    label: 'What does setup mode do?',
    message: "I'm in farm setup — walk me through zones, device, and comfort targets in order",
    contextRef: { type: 'route', path: '/chat', name: 'Farm Guardian chat', surface: 'setup_mode_chat' },
    setupMode: true,
  })

  return dedupeStarters(candidates).slice(0, max)
}

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

/**
 * Parse "Inventory low: OHN at …" alert subjects (mirrors Go lowStockInputFromSubject).
 * @param {object} alert
 */
export function lowStockInputFromAlert(alert) {
  const subject = String(alert?.subject_rendered || alert?.subject || '')
  const prefix = 'Inventory low:'
  if (!subject.includes(prefix)) return ''
  const rest = subject.split(prefix)[1]?.trim() || ''
  const at = rest.indexOf(' at ')
  return (at > 0 ? rest.slice(0, at) : rest).trim()
}

function operationsRouteRef(path, name, surface) {
  return { type: 'route', path, name, surface }
}

/**
 * Phase 43 WS8 — Supplies hub Guardian starters.
 */
export function buildSuppliesHubStarters({
  lowStockRows = [],
  lowStockAlerts = [],
  recipes = [],
  zones = [],
  zoneContextId = null,
  zoneName = '',
  programs = [],
  surface = 'supplies_hub',
}) {
  const max = surface === 'supplies_hub_zone' ? 3 : 4
  const path = zoneContextId ? `/operations/supplies?zone_id=${zoneContextId}` : '/operations/supplies'
  const routeRef = operationsRouteRef(path, 'Supplies', surface)
  const name = focusZoneName(zones, zoneContextId, zoneName)
  const starters = []

  if (lowStockRows.length) {
    starters.push({
      id: 'whats-running-low',
      label: "What's running low?",
      message: 'What supplies are below their low-stock threshold on this farm?',
      contextRef: routeRef,
    })
  }

  const lowAlert = lowStockAlerts[0]
  if (lowAlert) {
    const inputName = lowStockInputFromAlert(lowAlert)
    starters.push({
      id: 'refill-from-alert',
      label: 'Turn alert into refill task',
      message: `Create a refill task from alert #${lowAlert.id}${inputName ? ` for ${inputName}` : ''}`,
      contextRef: { ...routeRef, alert_id: lowAlert.id },
    })
  }

  if (zoneContextId && programs.some((p) => p.is_active && Number(p.target_zone_id) === Number(zoneContextId))) {
    starters.push({
      id: 'feeding-setup-zone',
      label: 'Feeding setup for this room',
      message: `Summarize feeding programs and reservoirs for ${name} — what should I check before the next run?`,
      contextRef: routeRef,
    })
  }

  const firstLow = lowStockRows[0]
  if (firstLow?.inputName && recipes.length) {
    starters.push({
      id: 'recipe-reorder',
      label: 'Which recipe uses this input?',
      message: `Which mixing recipes use ${firstLow.inputName} and what should I reorder?`,
      contextRef: routeRef,
    })
  }

  if (!starters.some((s) => s.id === 'whats-running-low')) {
    starters.push({
      id: 'log-mix-how',
      label: 'How do I log a mix?',
      message: 'How do I log a nutrient mix and tie it to inventory on this farm?',
      contextRef: routeRef,
    })
  }

  return dedupeStarters(starters).slice(0, max)
}

/**
 * Phase 43 WS8 — Feeding (details) admin hub starters.
 */
export function buildFeedingAdminStarters({
  zones = [],
  zoneContextId = null,
  programs = [],
}) {
  const path = zoneContextId ? `/operations/feeding?zone_id=${zoneContextId}` : '/operations/feeding'
  const routeRef = operationsRouteRef(path, 'Feeding (details)', 'feeding_admin')
  const name = focusZoneName(zones, zoneContextId)
  const starters = []

  const activeProgram = programs.find(
    (p) => p.is_active && (zoneContextId == null || Number(p.target_zone_id) === Number(zoneContextId)),
  )
  if (activeProgram) {
    starters.push({
      id: 'next-feed-schedule',
      label: 'When does feeding run next?',
      message: `When does the fertigation schedule for ${name} run next, in plain language?`,
      contextRef: routeRef,
    })
  }

  starters.push({
    id: 'feeding-admin-overview',
    label: 'Explain feeding setup',
    message: zoneContextId
      ? `Summarize active feeding programs, tanks, and strength targets for ${name}.`
      : 'Which rooms have active feeding programs I should review before the next run?',
    contextRef: routeRef,
  })

  return dedupeStarters(starters).slice(0, 3)
}

/**
 * Phase 43 WS8 — Money hub Guardian starters.
 */
export function buildMoneyHubStarters() {
  return [{
    id: 'month-spend',
    label: "Explain this month's spend",
    message: 'Summarize what I spent this month in plain language — no accounting jargon',
    contextRef: operationsRouteRef('/operations/money', 'Money', 'money_hub'),
  }]
}

/**
 * Phase 43 WS8 — Dashboard starters when supplies are low.
 */
export function buildDashboardOpsStarters({ lowStockCount = 0, lowStockAlerts = [] }) {
  if (!lowStockCount) return []
  const routeRef = operationsRouteRef('/', 'Dashboard', 'dashboard_ops')
  const starters = [{
    id: 'whats-running-low',
    label: "What's running low?",
    message: 'What supplies are below their low-stock threshold on this farm?',
    contextRef: routeRef,
  }]
  const lowAlert = lowStockAlerts[0]
  if (lowAlert) {
    const inputName = lowStockInputFromAlert(lowAlert)
    starters.push({
      id: 'refill-from-alert',
      label: 'Turn alert into refill task',
      message: `Create a refill task from alert #${lowAlert.id}${inputName ? ` for ${inputName}` : ''}`,
      contextRef: { ...routeRef, alert_id: lowAlert.id },
    })
  }
  return dedupeStarters(starters).slice(0, 2)
}

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
