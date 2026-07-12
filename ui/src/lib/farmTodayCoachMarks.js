/**
 * Phase 177 — first-visit Today coach marks (session dismiss).
 */

export const TODAY_COACH_DONE_KEY = 'gr33n_today_coach_done'

const NARROW_VIEWPORT_MAX = 480

function safeSessionGet(key) {
  try {
    return sessionStorage.getItem(key)
  } catch {
    return null
  }
}

function safeSessionSet(key, value) {
  try {
    sessionStorage.setItem(key, value)
  } catch {
    /* private browsing / SSR */
  }
}

export function isTodayCoachDone() {
  return safeSessionGet(TODAY_COACH_DONE_KEY) === '1'
}

export function markTodayCoachDone() {
  safeSessionSet(TODAY_COACH_DONE_KEY, '1')
}

/**
 * @param {number} [width]
 */
export function isNarrowTodayViewport(width = typeof window !== 'undefined' ? window.innerWidth : 1280) {
  return Number(width || 0) <= NARROW_VIEWPORT_MAX
}

/**
 * @param {{ hasAttention?: boolean, narrowViewport?: boolean }} opts
 */
export function buildTodayCoachSteps({ hasAttention = false, narrowViewport = false } = {}) {
  if (narrowViewport) {
    return [{
      id: 'tap_zone',
      title: 'Tap a zone',
      body: 'Open quick actions — water, lights, tasks, and alerts — without leaving Today.',
      target: 'today-farm-hero',
    }]
  }

  const steps = [
    {
      id: 'farm_map',
      title: 'This is your farm',
      body: 'Your grow areas on one map. Bigger farms can filter or switch to a list.',
      target: 'today-farm-hero',
    },
    {
      id: 'tap_zone',
      title: 'Tap a zone',
      body: 'Quick actions open here — feed, lights, tasks, and what needs attention.',
      target: 'today-farm-hero',
    },
  ]

  if (hasAttention) {
    steps.push({
      id: 'attention',
      title: 'Needs attention',
      body: 'Flagged zones show up here first. Tap a chip to triage without hunting the map.',
      target: 'farm-today-attention',
    })
  } else {
    steps.push({
      id: 'pulse',
      title: 'Farm pulse',
      body: 'Next water, lights, crops, and devices — operational context at a glance.',
      target: 'farm-site-strip',
    })
  }

  return steps
}

/**
 * @param {boolean} [reduceMotion]
 */
export function todayCoachTransitionClass(reduceMotion = false) {
  if (reduceMotion) return ''
  return 'transition-opacity duration-200 motion-safe:transition-transform'
}
