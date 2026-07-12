/**
 * Phase 175 — curated Today Ask gr33n starters (≤2 chips on populated farms).
 */

const MORNING_CHECK_DAY_KEY = 'gr33n_today_morning_check_day'

/**
 * Morning check chip before noon local, or once per calendar day on first Today visit.
 * @param {Date} [now]
 */
export function shouldOfferMorningCheckOnToday(now = new Date()) {
  if (now.getHours() < 12) return true
  const day = now.toISOString().slice(0, 10)
  try {
    if (sessionStorage.getItem(MORNING_CHECK_DAY_KEY) !== day) {
      sessionStorage.setItem(MORNING_CHECK_DAY_KEY, day)
      return true
    }
  } catch {
    return true
  }
  return false
}

/**
 * @param {object} params
 * @param {object[]} [params.morningStarters]
 * @param {boolean} [params.showMorningCheck]
 * @param {string} [params.farmName]
 */
export function buildCuratedTodayAskStarters({
  morningStarters = [],
  showMorningCheck = false,
  farmName = '',
} = {}) {
  const starters = []
  if (showMorningCheck && morningStarters.length) {
    starters.push(morningStarters[0])
  }
  const farmBit = farmName ? ` at ${farmName}` : ''
  starters.push({
    id: 'ask-about-farm',
    label: 'Ask about your farm',
    message: `What should I focus on at my farm${farmBit} today?`,
    contextRef: {
      type: 'route',
      path: '/',
      name: 'Today',
      surface: 'dashboard_ask',
    },
  })
  return starters.slice(0, 2)
}

/**
 * Power-user starters for the collapsed details subsection.
 * @param {object[][]} groups
 */
export function mergeTodayDetailsGuardianStarters(...groups) {
  const seen = new Set()
  const out = []
  for (const group of groups) {
    for (const s of group || []) {
      if (!s?.id || seen.has(s.id)) continue
      seen.add(s.id)
      out.push(s)
    }
  }
  return out
}
