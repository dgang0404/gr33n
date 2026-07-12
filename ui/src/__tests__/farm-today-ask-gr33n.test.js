import { describe, it, expect, beforeEach, afterEach } from 'vitest'
import {
  buildCuratedTodayAskStarters,
  mergeTodayDetailsGuardianStarters,
  shouldOfferMorningCheckOnToday,
} from '../lib/farmTodayAskGr33n.js'

const morning = [{
  id: 'morning-check',
  label: 'Morning check',
  message: 'Walk the farm',
  contextRef: { type: 'route', path: '/', name: 'Today' },
}]

describe('Phase 175 — farmTodayAskGr33n', () => {
  describe('shouldOfferMorningCheckOnToday', () => {
    beforeEach(() => sessionStorage.clear())
    afterEach(() => sessionStorage.clear())

    it('offers morning check before noon', () => {
      expect(shouldOfferMorningCheckOnToday(new Date('2026-07-12T09:00:00'))).toBe(true)
    })

    it('offers morning check once per day after noon on first visit', () => {
      const afternoon = new Date('2026-07-12T14:00:00')
      expect(shouldOfferMorningCheckOnToday(afternoon)).toBe(true)
      expect(shouldOfferMorningCheckOnToday(afternoon)).toBe(false)
    })
  })

  it('buildCuratedTodayAskStarters caps at two chips', () => {
    const starters = buildCuratedTodayAskStarters({
      morningStarters: morning,
      showMorningCheck: true,
      farmName: 'Demo Farm',
    })
    expect(starters).toHaveLength(2)
    expect(starters[0].id).toBe('morning-check')
    expect(starters[1].id).toBe('ask-about-farm')
    expect(starters[1].message).toContain('Demo Farm')
  })

  it('buildCuratedTodayAskStarters omits morning when not offered', () => {
    const starters = buildCuratedTodayAskStarters({ showMorningCheck: false })
    expect(starters).toHaveLength(1)
    expect(starters[0].label).toBe('Ask about your farm')
  })

  it('mergeTodayDetailsGuardianStarters dedupes by id', () => {
    const merged = mergeTodayDetailsGuardianStarters(
      [{ id: 'a', label: 'A' }],
      [{ id: 'a', label: 'A dup' }, { id: 'b', label: 'B' }],
    )
    expect(merged).toHaveLength(2)
    expect(merged.map((s) => s.id)).toEqual(['a', 'b'])
  })
})
