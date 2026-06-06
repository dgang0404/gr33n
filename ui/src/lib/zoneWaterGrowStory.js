/**
 * Phase 40 WS5 — compose last/next feed + queue narrative for zone Water tab.
 */

import { scheduleRunsLabel } from './cronHumanize.js'

/**
 * @param {object[]} events
 * @param {number} zoneId
 */
export function lastZoneFeedEvent(events, zoneId) {
  return (events || [])
    .filter((e) => Number(e.zone_id) === Number(zoneId))
    .sort((a, b) => new Date(b.applied_at) - new Date(a.applied_at))[0] || null
}

/**
 * @param {object|null} event
 * @param {string} [programName]
 */
export function formatLastFeedLine(event, programName = '') {
  if (!event) return 'No feed logged yet for this zone'
  const when = formatWhen(event.applied_at)
  const vol = event.volume_applied_liters != null ? `${event.volume_applied_liters}L` : null
  const ec =
    event.ec_before_mscm != null && event.ec_after_mscm != null
      ? `EC ${event.ec_before_mscm} → ${event.ec_after_mscm}`
      : null
  const parts = [when, vol, ec, programName].filter(Boolean)
  return parts.join(' · ')
}

/**
 * @param {object|null} schedule
 * @param {object|null} activeProgram
 */
export function formatNextFeedLine(schedule, activeProgram) {
  if (!activeProgram) return 'No active feeding program'
  if (!schedule) return `Program runs on demand — ${activeProgram.name}`
  const when = scheduleRunsLabel(schedule)
  return `${when} · ${activeProgram.name}`
}

/**
 * @param {object|null} command
 */
export function formatQueueHeadLabel(command) {
  if (!command) return null
  const t = String(command.command_type || '').toLowerCase()
  if (t === 'mix_batch') return 'Mix batch'
  if (t === 'pulse') {
    const sec = command.payload?.duration_seconds
    return sec ? `Pump pulse (${sec}s)` : 'Pump pulse'
  }
  if (t === 'actuator') {
    const cmd = command.payload?.command || command.command
    return cmd ? `Pump ${cmd}` : 'Pump command'
  }
  return t || 'Command'
}

/**
 * @param {number} depth
 * @param {object|null} headCommand
 */
export function formatEdgeQueueLine(depth, headCommand) {
  const n = Number(depth) || 0
  if (n <= 0) return 'Nothing waiting on the Pi'
  const head = formatQueueHeadLabel(headCommand)
  return head ? `${n} queued · next: ${head}` : `${n} command${n === 1 ? '' : 's'} queued`
}

function formatWhen(iso) {
  if (!iso) return '—'
  try {
    return new Date(iso).toLocaleString(undefined, {
      month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit',
    })
  } catch {
    return iso
  }
}

/**
 * @param {object} ctx
 */
export function buildWaterGrowStory(ctx) {
  const {
    zoneId,
    events = [],
    programs = [],
    schedules = [],
    activeProgram,
    waterStatus = null,
    queueHead = null,
  } = ctx

  const lastEvent = lastZoneFeedEvent(events, zoneId)
  const programName = activeProgram?.name
    || (lastEvent?.program_id
      ? programs.find((p) => p.id === lastEvent.program_id)?.name
      : null)
    || ''

  const schedule = activeProgram?.schedule_id
    ? schedules.find((s) => s.id === activeProgram.schedule_id)
    : null

  return {
    lastFeed: {
      line: formatLastFeedLine(lastEvent, programName),
      event: lastEvent,
    },
    nextFeed: {
      line: formatNextFeedLine(schedule, activeProgram),
      schedule,
    },
    edge: {
      line: formatEdgeQueueLine(waterStatus?.queue_depth, queueHead),
      depth: waterStatus?.queue_depth ?? 0,
      headType: queueHead?.command_type || null,
    },
    irrigationOnly: Boolean(activeProgram?.irrigation_only),
  }
}
