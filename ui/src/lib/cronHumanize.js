/**
 * Farmer-facing cron text (Phase 40). Avoid showing raw cron on grow routes.
 */

function formatClock(hour, minute = 0) {
  const h = Number(hour)
  const m = Number(minute) || 0
  if (Number.isNaN(h)) return null
  const period = h >= 12 ? 'PM' : 'AM'
  const h12 = h % 12 === 0 ? 12 : h % 12
  return m ? `${h12}:${String(m).padStart(2, '0')} ${period}` : `${h12} ${period}`
}

/**
 * @param {string|undefined|null} cron
 * @returns {string|null}
 */
export function humanizeCron(cron) {
  const expr = String(cron || '').trim()
  if (!expr) return null
  const fields = expr.split(/\s+/)
  if (fields.length < 5) return null

  const [minField, hourField, dom, , dow] = fields
  const hourList = hourField.includes(',')
    ? hourField.split(',').map((h) => Number(h)).filter((h) => !Number.isNaN(h) && h >= 0 && h <= 23)
    : hourField === '*'
      ? []
      : [Number(hourField)]

  const min = minField === '*' ? 0 : Number(minField)
  const minute = Number.isNaN(min) ? 0 : min

  if (hourList.length > 1 && dom === '*' && (dow === '*' || dow === '?')) {
    const times = hourList.map((h) => formatClock(h, minute)).filter(Boolean)
    if (times.length) return `Today at ${times.join(' and ')}`
  }

  if (hourList.length === 1 && dom === '*' && (dow === '*' || dow === '?') && (minField === '0' || minField === '00' || minField === '*')) {
    const t = formatClock(hourList[0], minute)
    if (t) return `Every day at ${t}`
  }

  if (hourField === '*' && dom === '*' && (dow === '*' || dow === '?')) {
    return 'Runs throughout the day'
  }

  return null
}

/**
 * @param {{ name?: string, cron_expression?: string }} schedule
 * @returns {string}
 */
export function scheduleRunsLabel(schedule) {
  if (!schedule) return 'No run scheduled'
  const human = humanizeCron(schedule.cron_expression)
  if (human) return human
  return schedule.name || 'Scheduled'
}

/**
 * Build a daily cron expression from local clock time (Phase 42).
 * @param {number} hour 0–23
 * @param {number} [minute=0]
 * @returns {string}
 */
export function buildDailyCron(hour, minute = 0) {
  const h = Math.min(23, Math.max(0, Number(hour) || 0))
  const m = Math.min(59, Math.max(0, Number(minute) || 0))
  return `${m} ${h} * * *`
}
