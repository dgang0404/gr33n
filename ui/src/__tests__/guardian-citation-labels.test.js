import { describe, it, expect } from 'vitest'
import { citationSourceLabel, citationChipClass, trimWarningMessage } from '../lib/guardianCitationLabels.js'

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
