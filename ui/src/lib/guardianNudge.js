/**
 * Phase 61 — proactive Guardian nudge helpers.
 */

/** Build chat payload when the operator taps Review on a nudge strip. */
export function buildNudgeReviewPayload(nudge) {
  if (!nudge?.category) return null
  const plain = String(nudge.message || '').replace(/\s*—\s*tap to review$/i, '').trim()
  const message =
    plain.length > 0
      ? `Please help me with this Guardian nudge: ${plain}`
      : 'Please help me review this Guardian nudge.'
  return {
    message,
    contextRef: {
      type: 'route',
      path: nudge.action_route || '/',
      nudge_category: nudge.category,
      nudge_id: nudge.nudge_id || '',
    },
  }
}
