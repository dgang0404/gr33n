/**
 * Phase 40 WS7b — snapshot-aware Ask Guardian prefills (replaces generic zone status).
 */

/**
 * @param {object} ctx
 * @param {object} ctx.zone — { id, name }
 * @param {string} [ctx.activeTab]
 * @param {object[]} [ctx.unreadAlerts]
 * @param {number} [ctx.queueDepth]
 * @param {number} [ctx.missingComfortTargets]
 * @param {number} [ctx.offlineDevices]
 * @param {object|null} [ctx.nextSchedule] — { schedule: { name } }
 */
export function buildZoneGuardianPrompt(ctx) {
  const zoneName = ctx.zone?.name || 'this room'
  const tab = ctx.activeTab && ctx.activeTab !== 'overview' ? ctx.activeTab : null
  const tabHint = tab ? ` (I'm on the ${tab} tab)` : ''

  const alerts = ctx.unreadAlerts || []
  if (alerts.length) {
    const top = alerts[0]
    return `Explain alert #${top.id} (${top.subject_rendered || 'alert'}) for ${zoneName} and what I should do in the next 10 minutes${tabHint}.`
  }

  if (ctx.queueDepth > 0) {
    return `What's queued for devices in ${zoneName} (queue depth ${ctx.queueDepth}) and is it safe to run another pulse or feed now?${tabHint}`
  }

  if (ctx.missingComfortTargets > 0) {
    return `What humidity and temperature comfort targets should I set for ${zoneName} at my current crop stage?${tabHint}`
  }

  if (ctx.offlineDevices > 0) {
    return `${ctx.offlineDevices} device(s) are offline in ${zoneName}. What should I check first on-site?${tabHint}`
  }

  if (ctx.nextSchedule?.schedule?.name) {
    return `Walk me through what runs today in ${zoneName} — starting with "${ctx.nextSchedule.schedule.name}" — and anything I should verify before lights or feeding fire.${tabHint}`
  }

  return `Summarize what matters today in ${zoneName}: comfort, feeding, lights, and any risks I should handle before end of shift.${tabHint}`
}

/**
 * @param {object} ctx — same shape as buildZoneGuardianPrompt
 * @returns {{ type: 'zone', id: number, name: string, tab?: string }}
 */
export function buildZoneGuardianContextRef(ctx) {
  const ref = {
    type: 'zone',
    id: ctx.zone?.id,
    name: ctx.zone?.name,
  }
  if (ctx.activeTab && ctx.activeTab !== 'overview') ref.tab = ctx.activeTab
  if (ctx.unreadAlerts?.[0]?.id) ref.alert_id = ctx.unreadAlerts[0].id
  return ref
}
