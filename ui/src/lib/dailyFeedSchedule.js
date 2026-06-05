/**
 * Phase 47 WS3 — plain daily feed time ↔ cron (no cron shown to operators).
 */

/**
 * Parse a simple daily cron (`M H * * *`) into HH:MM for a time input.
 * @param {string|undefined|null} cron
 * @returns {string|null} HH:MM or null if not a simple daily pattern
 */
export function parseDailyFeedTime(cron) {
  const expr = String(cron || '').trim()
  if (!expr) return null
  const fields = expr.split(/\s+/)
  if (fields.length < 5) return null

  const [minField, hourField, dom, , dow] = fields
  if (dom !== '*' || (dow !== '*' && dow !== '?')) return null
  if (hourField.includes(',') || hourField.includes('/') || minField.includes('/')) return null

  const hour = hourField === '*' ? 6 : Number(hourField)
  const minute = minField === '*' ? 0 : Number(minField)
  if (Number.isNaN(hour) || Number.isNaN(minute)) return null
  if (hour < 0 || hour > 23 || minute < 0 || minute > 59) return null

  return `${String(hour).padStart(2, '0')}:${String(minute).padStart(2, '0')}`
}

/**
 * @param {string|undefined|null} timeStr HH:MM
 * @returns {string} daily cron expression
 */
export function buildDailyFeedCron(timeStr) {
  const raw = String(timeStr || '06:00').trim()
  const [hPart, mPart] = raw.split(':')
  let hour = Number(hPart)
  let minute = Number(mPart)
  if (Number.isNaN(hour)) hour = 6
  if (Number.isNaN(minute)) minute = 0
  hour = Math.min(23, Math.max(0, hour))
  minute = Math.min(59, Math.max(0, minute))
  return `${minute} ${hour} * * *`
}

/**
 * @param {object|null} schedule
 * @returns {string}
 */
export function defaultDailyFeedTime(schedule) {
  return parseDailyFeedTime(schedule?.cron_expression) || '06:00'
}
