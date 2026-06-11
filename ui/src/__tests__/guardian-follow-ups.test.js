import { describe, it, expect } from 'vitest'
import { deriveFollowUps } from '../lib/guardianFollowUps.js'

describe('deriveFollowUps', () => {
  it('returns an array of max 3 chips', () => {
    const result = deriveFollowUps('tell me about cannabis', 'Cannabis prefers 18/6 during veg and 12/12 for flowering.')
    expect(result).toHaveLength(3)
    result.forEach((chip) => {
      expect(chip).toHaveProperty('id')
      expect(chip).toHaveProperty('label')
      expect(chip).toHaveProperty('message')
    })
  })

  it('deduplicates chip ids', () => {
    const result = deriveFollowUps('cannabis flowering EC pH', 'cannabis EC pH nutrients photoperiod lighting DLI VPD')
    const ids = result.map((c) => c.id)
    expect(new Set(ids).size).toBe(ids.length)
  })

  it('detects cannabis topics and returns flip or harvest chip', () => {
    const result = deriveFollowUps(
      'how do i water cannabis vs eggplant',
      'Cannabis requires 18/6 lighting during veg and 12/12 for flowering.',
    )
    const ids = result.map((c) => c.id)
    expect(ids.some((id) => ['flip-schedule', 'harvest-window', 'cannabis-vpd'].includes(id))).toBe(true)
  })

  it('skips flip-schedule chip when user already asked about flipping', () => {
    const result = deriveFollowUps(
      'when should i flip to 12/12?',
      'You should flip once the plant has reached 50% of its final desired height.',
    )
    const ids = result.map((c) => c.id)
    expect(ids).not.toContain('flip-schedule')
  })

  it('detects EC/nutrient topics', () => {
    const result = deriveFollowUps(
      'what EC should i use for tomatoes?',
      'Tomatoes do well with an EC of 2.0-3.5 mS/cm during fruiting.',
    )
    const ids = result.map((c) => c.id)
    expect(ids.some((id) => ['ec-measure', 'ec-runoff'].includes(id))).toBe(true)
  })

  it('detects pH topics', () => {
    const result = deriveFollowUps(
      'what pH do orchids prefer?',
      'Orchids prefer a slightly acidic pH between 5.5 and 6.5.',
    )
    const ids = result.map((c) => c.id)
    expect(ids.some((id) => ['ph-adjust', 'ph-symptoms'].includes(id))).toBe(true)
  })

  it('detects orchid topics', () => {
    const result = deriveFollowUps(
      'how do i care for my orchid?',
      'Phalaenopsis orchids prefer bright indirect light and weekly watering.',
    )
    const ids = result.map((c) => c.id)
    expect(ids.some((id) => ['orchid-repot', 'orchid-rebloom'].includes(id))).toBe(true)
  })

  it('detects ramps topics', () => {
    const result = deriveFollowUps(
      'tell me about ramps',
      'Ramps (Allium tricoccum) are spring ephemerals that prefer moist woodland soil.',
    )
    const ids = result.map((c) => c.id)
    expect(ids.some((id) => ['ramps-dormancy', 'ramps-harvest'].includes(id))).toBe(true)
  })

  it('detects eggplant topics', () => {
    const result = deriveFollowUps(
      'how do i grow eggplant indoors?',
      'Eggplant needs full sun equivalent — 14-16 hours of light per day indoors.',
    )
    const ids = result.map((c) => c.id)
    expect(ids.some((id) => ['pollination', 'fruiting-nutrients'].includes(id))).toBe(true)
  })

  it('returns generic fallback chips when no topics are detected', () => {
    const result = deriveFollowUps('hello', 'Hello! How can I help today?')
    expect(result).toHaveLength(3)
    const ids = result.map((c) => c.id)
    expect(ids.some((id) => ['grow-status', 'next-action', 'optimize-yield'].includes(id))).toBe(true)
  })

  it('handles empty strings without throwing', () => {
    expect(() => deriveFollowUps('', '')).not.toThrow()
    const result = deriveFollowUps('', '')
    expect(Array.isArray(result)).toBe(true)
  })

  it('detects lighting topics', () => {
    const result = deriveFollowUps(
      'what lighting do i need for indoor growing?',
      'LED fixtures at 200-400 µmol/m²/s PPFD are ideal. Aim for a DLI of 20-40 mol/m²/day.',
    )
    const ids = result.map((c) => c.id)
    expect(ids.some((id) => ['dli-target', 'light-intensity'].includes(id))).toBe(true)
  })

  it('detects alert topics', () => {
    const result = deriveFollowUps(
      'what does this alert mean?',
      'The humidity alert indicates your RH has exceeded the comfort band threshold.',
    )
    const ids = result.map((c) => c.id)
    expect(ids).toContain('alert-next-step')
  })
})
