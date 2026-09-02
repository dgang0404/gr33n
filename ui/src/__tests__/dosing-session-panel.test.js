import { describe, it, expect } from 'vitest'
import { dosingStatusLabel, dosingElapsed, isDosingSessionTerminal } from '../lib/dosingSession.js'

describe('dosingSession lib (Phase 216 WS6/WS8)', () => {
  describe('dosingStatusLabel', () => {
    it('maps known statuses to operator-facing labels', () => {
      expect(dosingStatusLabel({ status: 'idle' })).toBe('Waiting to dose')
      expect(dosingStatusLabel({ status: 'circulating' })).toBe('Circulating…')
      expect(dosingStatusLabel({ status: 'done' })).toBe('Done')
      expect(dosingStatusLabel({ status: 'aborted' })).toBe('Aborted')
    })

    it('falls back to the raw status for unknown values', () => {
      expect(dosingStatusLabel({ status: 'some_future_state' })).toBe('some_future_state')
    })

    it('handles missing session gracefully', () => {
      expect(dosingStatusLabel(null)).toBe('')
      expect(dosingStatusLabel({})).toBe('')
    })
  })

  describe('dosingElapsed', () => {
    it('shows <1 min for a session started seconds ago', () => {
      const now = Date.now()
      const session = { started_at: new Date(now - 5000).toISOString() }
      expect(dosingElapsed(session, now)).toBe('<1 min')
    })

    it('shows whole minutes under an hour', () => {
      const now = Date.now()
      const session = { started_at: new Date(now - 25 * 60000).toISOString() }
      expect(dosingElapsed(session, now)).toBe('25 min')
    })

    it('shows hours + minutes past an hour', () => {
      const now = Date.now()
      const session = { started_at: new Date(now - 90 * 60000).toISOString() }
      expect(dosingElapsed(session, now)).toBe('1h 30m')
    })

    it('returns empty string for a missing/invalid started_at', () => {
      expect(dosingElapsed({})).toBe('')
      expect(dosingElapsed({ started_at: 'not-a-date' })).toBe('')
    })
  })

  describe('isDosingSessionTerminal', () => {
    it('is true for done and aborted', () => {
      expect(isDosingSessionTerminal({ status: 'done' })).toBe(true)
      expect(isDosingSessionTerminal({ status: 'aborted' })).toBe(true)
    })

    it('is false for in-progress statuses', () => {
      for (const status of ['idle', 'dosing', 'circulating', 'settling', 'sampling', 'evaluating']) {
        expect(isDosingSessionTerminal({ status })).toBe(false)
      }
    })
  })
})
