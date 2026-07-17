/**
 * Phase 196 — compact revision timeline from session user turns.
 */

const SNIPPET_MAX = 96

export function snippetUserMessage(text, max = SNIPPET_MAX) {
  const s = String(text || '').replace(/\s+/g, ' ').trim()
  if (!s) return ''
  if (s.length <= max) return s
  return `${s.slice(0, max - 1)}…`
}

/** Heuristic cue for a revise turn (mirrors common proposals_revise patterns). */
export function inferRevisionCue(userMessage, tool) {
  const q = String(userMessage || '').trim()
  if (!q) return null

  if (
    /\bzone(?:\s+id)?\s*#?\d+\b/i.test(q)
    || /\bput it in\b/i.test(q)
    || /\bzone for this task\b/i.test(q)
    || /\bassign(?:ment)?\s+(?:to|in)\s+/i.test(q)
  ) {
    return 'zone assigned'
  }

  if (/\bdue(?:\s+date)?\b|\bdue tomorrow\b|\bdue in \d+/i.test(q)) {
    return 'due_date set'
  }

  if (
    (tool === 'create_task' || tool === 'create_task_from_alert')
    && /(?:call it|rename|title|instead of|make the title)/i.test(q)
  ) {
    return 'title updated'
  }

  if (
    /\d+(?:\.\d+)?\s*(?:l\b|liters?\b|litres?\b)/i.test(q)
    && (tool === 'patch_fertigation_program' || tool === 'apply_grow_setup_pack')
  ) {
    return 'volume updated'
  }

  if (/\b(?:revise|instead|correction|use \d)/i.test(q)) {
    return 'revised'
  }

  return null
}

/**
 * @param {Array<{ user_message?: string, userMessage?: string }>} turns
 * @param {{ tool?: string }} [opts]
 */
export function buildRevisionTimeline(turns, opts = {}) {
  const tool = opts.tool || ''
  const withUser = (turns || [])
    .map((t) => String(t.user_message || t.userMessage || '').trim())
    .filter(Boolean)

  return withUser.map((userMessage, idx) => ({
    index: idx + 1,
    userMessage: snippetUserMessage(userMessage),
    cue: idx === 0 ? null : inferRevisionCue(userMessage, tool),
  }))
}

export function revisionTimelineLabel(revision, turnCount) {
  const rev = Number(revision) || 1
  const turns = Number(turnCount) || 0
  if (turns > 0) return `Revision history (${turns} turn${turns === 1 ? '' : 's'})`
  if (rev > 1) return `Revision history (${rev} revisions)`
  return 'Revision history'
}
