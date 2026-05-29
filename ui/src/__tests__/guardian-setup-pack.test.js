import { describe, it, expect } from 'vitest'
import { formatSetupPackBundle } from '../lib/guardianSetupPack.js'

describe('formatSetupPackBundle', () => {
  it('formats house plant bundle steps', () => {
    const { profile, steps } = formatSetupPackBundle({
      profile: 'house_plant',
      zone_id: 12,
      zone_name: 'Living Room',
      plant: {
        display_name: 'Philodendron',
        variety_or_cultivar: 'heartleaf',
        notes: 'RO water only',
      },
      cycle: {
        name: 'Philodendron — Living Room',
        current_stage: 'vegetative',
        started_at: '2026-05-27',
      },
      program: {
        name: 'Philodendron light feed',
        total_volume_liters: 0.5,
        ec_trigger_low: 0.8,
        ph_trigger_low: 5.8,
        ph_trigger_high: 6.5,
      },
      optional_task: { title: 'Monitor new Philodendron — first two weeks' },
    })

    expect(profile).toBe('house_plant')
    expect(steps).toHaveLength(5)
    expect(steps[0]).toContain('Philodendron')
    expect(steps[0]).toContain('heartleaf')
    expect(steps[1]).toContain('Living Room')
    expect(steps[2]).toContain('vegetative')
    expect(steps[3]).toContain('0.5L')
    expect(steps[3]).toContain('EC low 0.8')
    expect(steps[3]).toContain('pH 5.8–6.5')
    expect(steps[4]).toContain('Monitor new Philodendron')
  })
})
