/**
 * Phase 40 WS7b — zone conversation starter chips (send known-good prompts, not PRs).
 */

import { buildZoneGuardianPrompt, buildZoneGuardianContextRef } from './guardianContextPrompts.js'

/**
 * @param {'zone_overview'|'zone_water'|'zone_light'|'zone_climate'} surface
 * @param {object} ctx — zone snapshot (see guardianContextPrompts)
 * @returns {Array<{ id: string, label: string, message: string, contextRef?: object }>}
 */
export function buildZoneStarters(surface, ctx) {
  const zoneName = ctx.zone?.name || 'this room'
  const starters = []
  const baseRef = () => buildZoneGuardianContextRef({ ...ctx, activeTab: tabForSurface(surface) })

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
      label: "Today's schedule",
      message: `What runs today in ${zoneName} for "${ctx.nextSchedule.schedule.name}"?`,
      contextRef: baseRef(),
    })
  }

  if (surface === 'zone_water' || surface === 'zone_overview') {
    if (ctx.queueDepth > 0) {
      starters.push({
        id: 'queue-safety',
        label: 'Check command queue',
        message: `What's queued for devices in ${zoneName} and is it safe to run another pulse?`,
        contextRef: baseRef(),
      })
    } else if (ctx.activeProgramName) {
      starters.push({
        id: 'feeding-plan',
        label: 'Review feeding plan',
        message: `Review the active feeding program "${ctx.activeProgramName}" for ${zoneName} — last feed, next run, and anything off target.`,
        contextRef: baseRef(),
      })
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

  return starters.slice(0, 5)
}

function tabForSurface(surface) {
  if (surface === 'zone_water') return 'water'
  if (surface === 'zone_light') return 'light'
  if (surface === 'zone_climate') return 'climate'
  return 'overview'
}
