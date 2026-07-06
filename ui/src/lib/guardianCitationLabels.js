/**
 * Phase 133 — citation source_type display labels and chip colors.
 */

const SOURCE_LABELS = {
  field_guide: 'Field guide',
  platform_doc: 'Platform doc',
  symptom_guide: 'Symptom guide',
  task: 'Farm note',
  crop_cycle: 'Farm note',
  schedule: 'Farm note',
  automation_rule: 'Farm note',
  alert_notification: 'Farm note',
}

const CURATED_TYPES = new Set(['field_guide', 'platform_doc', 'symptom_guide'])

export function citationSourceLabel(sourceType) {
  const t = String(sourceType || '').trim()
  if (!t) return 'Source'
  if (SOURCE_LABELS[t]) return SOURCE_LABELS[t]
  if (CURATED_TYPES.has(t)) return t.replace(/_/g, ' ')
  return 'Farm note'
}

export function citationChipClass(sourceType) {
  const t = String(sourceType || '').trim()
  if (t === 'field_guide') return 'bg-green-950/60 border-green-800/70 text-green-200'
  if (t === 'platform_doc') return 'bg-blue-950/50 border-blue-800/70 text-blue-200'
  if (t === 'symptom_guide') return 'bg-emerald-950/50 border-emerald-800/70 text-emerald-200'
  return 'bg-zinc-900 border-zinc-700 text-zinc-300'
}

export function trimWarningMessage(trimSummary) {
  if (!trimSummary || typeof trimSummary !== 'object') return ''
  const parts = []
  if (trimSummary.history_turns) parts.push(`history ${trimSummary.history_turns}`)
  if (trimSummary.rag_top_k) parts.push(`RAG ${trimSummary.rag_top_k}`)
  if (trimSummary.snapshot_reduced) parts.push('snapshot trimmed')
  if (!parts.length) return ''
  const window = trimSummary.effective_context_window
  const windowBit = window ? ` (${window} token budget)` : ''
  return `Long chat — ${parts.join(', ')} for CPU model${windowBit}. Start a new chat for best results.`
}
