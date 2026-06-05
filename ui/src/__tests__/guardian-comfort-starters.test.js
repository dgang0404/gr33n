import { describe, it, expect } from 'vitest'
import {
  buildComfortHubStarters,
  buildSchedulesFarmerStarters,
  buildRulesFarmerStarters,
} from '../lib/guardianStarters.js'

describe('Phase 42 WS8 — targets Guardian starters', () => {
  const zones = [{ id: 3, name: 'Flower Room' }]

  it('offers humidity band chip when band is missing', () => {
    const starters = buildComfortHubStarters({
      zones,
      zoneContextId: 3,
      cards: [{
        zone: zones[0],
        bands: [{ sensorType: 'humidity', status: 'missing' }],
      }],
      activeCycles: [{ zone_id: 3, current_stage: 'early_flower' }],
    })
    const humidity = starters.find((s) => s.id === 'set-humidity-band')
    expect(humidity).toBeTruthy()
    expect(humidity.message).toContain('humidity comfort band')
    expect(humidity.message).toContain('early flower')
    expect(humidity.message).not.toMatch(/current status/i)
  })

  it('offers pause shade chip when GH rule is active', () => {
    const starters = buildComfortHubStarters({
      zones,
      zoneContextId: 3,
      cards: [{ zone: zones[0], bands: [] }],
      rules: [{ id: 5, name: 'GH — deploy shade', is_active: true }],
    })
    expect(starters.some((s) => s.id === 'pause-shade-rule')).toBe(true)
  })

  it('schedules tab returns up to 3 specific starters', () => {
    const starters = buildSchedulesFarmerStarters({
      zones,
      zoneContextId: 3,
      schedules: [{ id: 1, name: 'Morning feed', is_active: true }],
    })
    expect(starters.length).toBeLessThanOrEqual(3)
    expect(starters[0].message).toContain('plain language')
  })

  it('rules tab offers pause shade when GH rule exists', () => {
    const starters = buildRulesFarmerStarters({
      zones,
      zoneContextId: 3,
      rules: [{ id: 2, name: 'GH shade hot', is_active: true }],
    })
    expect(starters.some((s) => s.id === 'pause-shade-chat')).toBe(true)
  })
})
