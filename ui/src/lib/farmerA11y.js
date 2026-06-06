/**
 * Phase 45 WS6 — light a11y helpers (focus ring, touch targets, action labels).
 */

/** Tailwind classes — use on high-stakes actions when global CSS is not enough. */
export const FARMER_FOCUS_RING =
  'focus:outline-none focus-visible:ring-2 focus-visible:ring-green-500/90 focus-visible:ring-offset-2 focus-visible:ring-offset-zinc-950'

/** ~44px min touch target on mobile (sit-in Dismiss/Confirm). */
export const FARMER_TOUCH_TARGET =
  'min-h-[44px] min-w-[44px] sm:min-h-0 sm:min-w-0 inline-flex items-center justify-center'

/**
 * @param {'confirm'|'dismiss'|'refine'} action
 * @param {string} [summary]
 */
export function guardianProposalAriaLabel(action, summary = '') {
  const short = String(summary || 'proposed action').replace(/\s+/g, ' ').trim().slice(0, 100)
  if (action === 'confirm') {
    return `Confirm proposed action: ${short}`
  }
  if (action === 'dismiss') {
    return `Dismiss suggestion without changing farm data: ${short}`
  }
  if (action === 'refine') {
    return `Refine proposed action before confirming: ${short}`
  }
  return short
}

/**
 * @param {string} [zoneName]
 * @param {string} [programName]
 */
export function runFeedNowAriaLabel(zoneName = '', programName = '') {
  const zone = zoneName || 'this zone'
  const prog = programName ? ` (${programName})` : ''
  return `Run feeding program now for ${zone}${prog}`
}

/**
 * @param {string} [actuatorName]
 * @param {number} [seconds]
 */
export function runPulseAriaLabel(actuatorName = 'actuator', seconds = 0) {
  const sec = Number(seconds) > 0 ? ` for ${seconds} seconds` : ''
  return `Run timed pulse on ${actuatorName || 'actuator'}${sec}`
}
