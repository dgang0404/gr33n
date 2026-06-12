/**
 * Phase 82 / 110 — Guardian chat honesty + picker closure tests.
 */
import { describe, it, expect } from 'vitest'
import { turnContextLabel, zeroChunkWarning } from '../lib/guardianChatHonesty.js'
import { filterPickerGroups } from '../lib/cropLibraryPicker.js'

describe('phase-82-closure', () => {
  it('labels zero-chunk turns as farm context', () => {
    expect(turnContextLabel({ grounded: true, context_count: 0 })).toBe(
      'farm context · 0 doc chunks',
    )
    expect(turnContextLabel({ grounded: true, context_count: 3 })).toBe('grounded · 3 chunks')
  })

  it('warns on suspicious zero-chunk assistant text', () => {
    expect(
      zeroChunkWarning({
        grounded: true,
        context_count: 0,
        assistant_message: 'Use [1] for EC targets.',
      }),
    ).toBe(true)
    expect(
      zeroChunkWarning({
        grounded: true,
        context_count: 2,
        assistant_message: 'Use [1] for EC targets.',
      }),
    ).toBe(false)
  })

  it('picker groups filter by crop name', () => {
    const groups = filterPickerGroups(
      {
        groups: [
          {
            key: 'leafy',
            label: 'Leafy',
            items: [{ crop_key: 'lettuce', display_name: 'Lettuce', search_terms: ['lettuce'] }],
          },
        ],
      },
      'lett',
    )
    expect(groups).toHaveLength(1)
    expect(groups[0].items).toHaveLength(1)
  })
})
