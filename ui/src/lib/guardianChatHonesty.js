/** Phase 82 WS2 — honest labels when RAG returned zero chunks. */

/** @param {{ grounded?: boolean, context_count?: number }} turn */
export function turnContextLabel(turn) {
  if (!turn?.grounded) return ''
  const n = turn.context_count || 0
  return n === 0 ? 'farm context · 0 doc chunks' : `grounded · ${n} chunks`
}

/** @param {{ grounded?: boolean, context_count?: number, assistant_message?: string }} turn */
export function zeroChunkWarning(turn) {
  if (!turn?.grounded || (turn.context_count || 0) > 0) return false
  const text = String(turn.assistant_message || '')
  return /\[\d+\]/.test(text) || /\d+\s*%\s*ec/i.test(text)
}
