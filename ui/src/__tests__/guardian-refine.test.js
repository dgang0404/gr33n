import { describe, it, expect } from 'vitest'
import { refinePrefillForProposal } from '../lib/guardianRefine.js'

describe('guardianRefine', () => {
  it('prefills a correction prompt from proposal summary', () => {
    const text = refinePrefillForProposal({ summary: 'Set feeding plan volume to 0.3L' })
    expect(text).toContain('Set feeding plan volume to 0.3L')
    expect(text).toContain('Correction:')
  })
})
