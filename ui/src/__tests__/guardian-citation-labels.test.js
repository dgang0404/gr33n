import { describe, it, expect } from 'vitest'
import { citationSourceLabel, citationChipClass, trimWarningMessage, accuracyNoteMessage } from '../lib/guardianCitationLabels.js'

describe('Phase 133 — citation labels', () => {
  it('labels field_guide and platform_doc', () => {
    expect(citationSourceLabel('field_guide')).toBe('Field guide')
    expect(citationSourceLabel('platform_doc')).toBe('Platform doc')
    expect(citationSourceLabel('task')).toBe('Farm note')
  })

  it('chip classes differ by type', () => {
    expect(citationChipClass('field_guide')).toMatch(/green/)
    expect(citationChipClass('platform_doc')).toMatch(/blue/)
    expect(citationChipClass('task')).toMatch(/zinc/)
  })

  it('trim warning summarizes reductions', () => {
    const msg = trimWarningMessage({
      history_turns: '20→8',
      snapshot_reduced: true,
      effective_context_window: 4096,
    })
    expect(msg).toContain('history 20→8')
    expect(msg).toContain('snapshot trimmed')
    expect(msg).toContain('4096')
  })
})

describe('Phase 152 — live accuracy note messages', () => {
  it('returns empty string when no note is present', () => {
    expect(accuracyNoteMessage('')).toBe('')
    expect(accuracyNoteMessage(null)).toBe('')
    expect(accuracyNoteMessage(undefined)).toBe('')
  })

  it('maps known detector codes to a farmer-facing caveat', () => {
    expect(accuracyNoteMessage('citation_number_mismatch: claim near [3] matches [5] instead')).toMatch(
      /wrong source number/,
    )
    expect(accuracyNoteMessage('truncated_answer_tail: ade0:')).toMatch(/cut off/)
    expect(accuracyNoteMessage('uncited_timeline_claim: Week 9')).toMatch(/week\/day count/)
    expect(accuracyNoteMessage('invented_assumption_math: if we assume')).toMatch(/assumption/)
  })

  it('falls back to a generic caveat for unmapped codes', () => {
    expect(accuracyNoteMessage('some_future_detector: detail')).toMatch(/flagged part of this answer/)
  })
})
