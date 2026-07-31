/**
 * Phase 133 — citation source_type display labels and chip colors.
 */

const SOURCE_LABELS = {
  field_guide: 'Field guide',
  platform_doc: 'Platform doc',
  symptom_guide: 'Symptom guide',
  task: 'Task',
  crop_cycle: 'Crop cycle',
  schedule: 'Schedule',
  automation_rule: 'Automation',
  executable_action: 'Schedule action',
  alert_notification: 'Farm note',
  input_batch: 'Batch',
  input_definition: 'Input',
  cost_transaction: 'Cost',
  fertigation_program: 'Feeding program',
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

/**
 * Phase 152 WS1 — farmer-facing text for a live AnswerAccuracyNote flag.
 * The backend note is a terse `code: detail` string (e.g.
 * "citation_number_mismatch: ..."); map the code prefix to a short, honest
 * caveat so operators know to double-check before acting, without exposing
 * the internal detector name.
 */
const ACCURACY_NOTE_MESSAGES = [
  ['citation_number_mismatch', 'Guardian may have attached the wrong source number to a claim.'],
  ['truncated_answer_tail', "Guardian's answer looks cut off mid-sentence."],
  ['uncited_timeline_claim', 'Guardian mentioned a week/day count without citing where it came from.'],
  ['invented_assumption_math', 'Guardian estimated a number using an assumption, not a farm record.'],
  ['duplicate_list_item', 'Guardian may have listed the same item twice under different numbers.'],
  ['garbled_token', "Guardian's answer contains a garbled word — generation may have glitched."],
  ['missing_numbered_citations', "Guardian listed items without citing its sources."],
  ['multiple_citations_per_list_item', 'Guardian attached more than one source number to a single item.'],
  ['ph_ec_unit_confusion', 'Guardian may have mislabeled a pH value with EC units (or vice versa).'],
]

export function accuracyNoteMessage(accuracyNote) {
  const note = String(accuracyNote || '').trim()
  if (!note) return ''
  const code = note.split(':')[0].trim()
  const match = ACCURACY_NOTE_MESSAGES.find(([prefix]) => prefix === code)
  const detail = match ? match[1] : 'Guardian flagged part of this answer for review.'
  return `${detail} Double-check citations before acting.`
}

/** Phase 158 — screen-reader label for citation deep links. */
export function citationLinkAriaLabel(citation) {
  const source = citationSourceLabel(citation?.source_type)
  const id = citation?.source_id ?? ''
  const excerpt = String(citation?.excerpt || '').replace(/\s+/g, ' ').trim().slice(0, 80)
  const excerptBit = excerpt ? ` — ${excerpt}` : ''
  return `Open ${source} #${id}${excerptBit}`
}

/**
 * Prefer server-resolved route; hub fallbacks for citations finalized before
 * a source type had a resolver (or when zone hops fail).
 * @param {{ route?: string, source_type?: string, source_id?: number|string } | null | undefined} citation
 * @returns {string | null}
 */
export function citationDisplayRoute(citation) {
  const route = String(citation?.route || '').trim()
  if (route) return route
  const st = String(citation?.source_type || '').trim()
  const id = citation?.source_id
  switch (st) {
    case 'alert_notification':
      return '/alerts'
    case 'cost_transaction':
      return '/money?tab=ledger'
    case 'input_batch':
      return id != null && id !== '' ? `/money?tab=supplies&batch_id=${id}` : '/money?tab=supplies'
    case 'input_definition':
      return '/natural-farming?tab=batch'
    case 'automation_rule':
      return '/comfort-targets?tab=automations'
    case 'executable_action':
    case 'schedule':
      return '/comfort-targets?tab=schedules'
    case 'symptom_guide':
      return '/operator-guide?tab=symptoms'
    case 'field_guide':
      return '/operator-guide?tab=knowledge'
    case 'platform_doc':
      return '/operator-guide?tab=library&section=guide'
    case 'fertigation_program':
      return '/feed-water?tab=programs'
    case 'task':
      return '/zones'
    case 'crop_cycle':
      return id != null && id !== '' ? `/crop-cycles/${id}/summary` : null
    default:
      return null
  }
}
