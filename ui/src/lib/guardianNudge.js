/**
 * Phase 61 — proactive Guardian nudge helpers.
 */

/** Parse alert id from nudge_id values like `alert-42`. */
export function parseAlertIdFromNudgeId(nudgeId) {
  const m = String(nudgeId || '').match(/^alert-(\d+)$/i)
  return m ? Number(m[1]) : null
}

/** Build chat payload when the operator taps Review on a nudge strip. */
export function buildNudgeReviewPayload(nudge) {
  if (!nudge?.category) return null
  const plain = String(nudge.message || '').replace(/\s*—\s*tap to review$/i, '').trim()
  const message =
    plain.length > 0
      ? `Please help me with this Guardian nudge: ${plain}`
      : 'Please help me review this Guardian nudge.'
  const base = {
    nudge_category: nudge.category,
    nudge_id: nudge.nudge_id || '',
  }
  const alertId = parseAlertIdFromNudgeId(nudge.nudge_id)
  const contextRef = alertId
    ? { type: 'alert', id: alertId, ...base }
    : {
        type: 'route',
        path: nudge.action_route || '/',
        ...base,
      }
  return { message, contextRef }
}
