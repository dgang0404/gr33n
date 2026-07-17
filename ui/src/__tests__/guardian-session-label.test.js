/**
 * Phase 197 — session sidebar labels for pending threads.
 */
import { describe, it, expect } from 'vitest'
import {
  baseSessionLabel,
  pendingProposalsBySessionId,
  sessionDisplayLabel,
} from '../lib/guardianSessionLabel.js'

describe('guardianSessionLabel (Phase 197)', () => {
  const session = {
    session_id: 'sess-task',
    title: '',
    first_user_message: 'Create a task to refill calcium nitrate when stock is low.',
    turn_count: 4,
  }

  it('prefers pending proposal summary over first user message', () => {
    const label = sessionDisplayLabel(session, {
      status: 'pending',
      summary: 'Create task: Refill calcium nitrate',
    })
    expect(label).toBe('Pending: Create task: Refill calcium nitrate')
  })

  it('falls back to title then first user message', () => {
    expect(baseSessionLabel({ title: 'My rename' })).toBe('My rename')
    expect(baseSessionLabel(session)).toContain('refill calcium nitrate')
    expect(sessionDisplayLabel(session, null)).toContain('refill calcium nitrate')
  })

  it('indexes pending proposals by session_id', () => {
    const map = pendingProposalsBySessionId([
      { session_id: 'a', status: 'pending', summary: 'First' },
      { session_id: 'a', status: 'pending', summary: 'Second' },
      { session_id: 'b', status: 'dismissed', summary: 'Gone' },
    ])
    expect(map.a.summary).toBe('First')
    expect(map.b).toBeUndefined()
  })
})
