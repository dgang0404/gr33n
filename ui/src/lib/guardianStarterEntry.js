/**
 * Phase 170 — when a Guardian starter should open Farm counsel vs Quick chat,
 * and when the drawer should auto-send (Today one-tap path).
 */

/**
 * @param {object|null|undefined} starter
 */
export function starterPrefersFarmCounsel(starter) {
  if (!starter) return false
  if (starter.setupMode) return true
  const mode = starter.contextRef?.guardian_mode
  if (mode === 'farm_counsel' || mode === 'morning_walkthrough') return true
  if (starter.contextRef?.type === 'zone') return true
  return false
}

/**
 * @param {object|null|undefined} starter
 * @param {{ inline?: boolean }} [opts]
 */
export function starterShouldAutoSend(starter, { inline = false } = {}) {
  if (inline || !starter) return false
  if (starter.autoSend === true) return true
  if (starter.autoSend === false) return false
  if (starter.setupMode) return false
  const mode = starter.contextRef?.guardian_mode
  if (mode === 'morning_walkthrough' || mode === 'farm_counsel') return true
  if (starter.contextRef?.type === 'zone') return true
  return false
}
