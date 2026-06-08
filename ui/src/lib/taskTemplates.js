/**
 * Phase 58 WS2 — task templates (client-side payloads for POST /tasks).
 */

import { buildRefillTaskPayload } from './suppliesHub.js'

/**
 * @param {object} row — low-stock row from listLowStockBatches
 */
export function refillTaskFromLowStock(row) {
  const payload = buildRefillTaskPayload(row)
  return { ...payload, task_type: 'refill', template_id: 'refill' }
}

/**
 * @param {object} alert
 * @param {object[]} [sensors]
 */
export function buildCheckSensorTaskPayload(alert, sensors = []) {
  let sensorName = 'sensor'
  if (alert?.triggering_event_source_type === 'sensor' && alert.triggering_event_source_id) {
    const s = sensors.find((x) => x.id === Number(alert.triggering_event_source_id))
    sensorName = s?.name || `sensor #${alert.triggering_event_source_id}`
  } else if (alert?.subject_rendered) {
    sensorName = String(alert.subject_rendered).replace(/alert/i, '').trim() || sensorName
  }
  return {
    title: `Check ${sensorName} wiring`,
    description: `Follow up on alert #${alert?.id}: verify wiring and last reading.`,
    priority: 2,
    task_type: 'check_sensor',
    template_id: 'check_sensor',
  }
}

/**
 * @param {object} schedule
 */
export function buildReviewFeedingPlanPayload(schedule) {
  const name = schedule?.name || 'Scheduled feed'
  return {
    title: `Review feeding plan — ${name}`,
    description: `${name} was due but hasn't run. Confirm whether that was intentional.`,
    priority: 2,
    task_type: 'review_feeding',
    template_id: 'review_feeding',
    schedule_id: schedule?.id ?? null,
  }
}

/**
 * @param {string} recipeName
 */
export function buildLogMixTaskPayload(recipeName = 'mix') {
  return {
    title: `Log mix — ${recipeName}`,
    description: 'Record a nutrient mix and link stock drawdown when you complete the task.',
    priority: 1,
    task_type: 'log_mix',
    template_id: 'log_mix',
  }
}

/**
 * Pick a template payload from an alert context.
 * @param {object} alert
 * @param {object[]} [sensors]
 */
export function detectAlertTaskTemplate(alert, sensors = []) {
  if (!alert) return null
  const subject = String(alert.subject_rendered || alert.message_text_rendered || '').toLowerCase()
  const offline = subject.includes('offline') || subject.includes('no reading') || subject.includes('stale')
  if (offline || alert.triggering_event_source_type === 'sensor') {
    return buildCheckSensorTaskPayload(alert, sensors)
  }
  return null
}

/**
 * @param {object[]} schedules
 */
export function detectMissedFeedSchedule(schedules) {
  const now = Date.now()
  const graceMs = 30 * 60 * 1000
  for (const s of schedules || []) {
    if (!s?.is_active || !s.next_expected_trigger_time) continue
    const due = new Date(s.next_expected_trigger_time).getTime()
    if (!Number.isFinite(due) || now - due < graceMs) continue
    const last = s.last_triggered_time ? new Date(s.last_triggered_time).getTime() : 0
    if (!last || last < due) return s
  }
  return null
}
