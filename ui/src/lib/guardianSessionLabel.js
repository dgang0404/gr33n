/**
 * Phase 197 — session sidebar labels when a pending proposal is linked.
 */

const LABEL_MAX = 80

export function baseSessionLabel(session) {
  if (!session) return 'Untitled'
  if (session.title && String(session.title).trim()) return String(session.title).trim()
  if (session.first_user_message && String(session.first_user_message).trim()) {
    return String(session.first_user_message).trim()
  }
  return 'Untitled'
}

export function trimSessionLabel(text, max = LABEL_MAX) {
  const s = String(text || '').replace(/\s+/g, ' ').trim()
  if (!s) return ''
  if (s.length <= max) return s
  return `${s.slice(0, max - 1)}…`
}

/**
 * @param {object} session
 * @param {{ summary?: string, status?: string } | null | undefined} pendingProposal
 */
export function sessionDisplayLabel(session, pendingProposal) {
  if (
    pendingProposal
    && pendingProposal.status === 'pending'
    && pendingProposal.summary
  ) {
    return trimSessionLabel(`Pending: ${String(pendingProposal.summary).trim()}`)
  }
  return trimSessionLabel(baseSessionLabel(session))
}

/**
 * @param {Array<{ session_id?: string, status?: string, summary?: string }>} proposals
 */
export function pendingProposalsBySessionId(proposals) {
  const out = {}
  for (const p of proposals || []) {
    if (p.status !== 'pending' || !p.session_id) continue
    if (!out[p.session_id]) out[p.session_id] = p
  }
  return out
}
