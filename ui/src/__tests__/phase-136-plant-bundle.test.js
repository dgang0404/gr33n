import { describe, it, expect } from 'vitest'
import { buildZoneGrowStripStarters } from '../lib/guardianStarters.js'

describe('Phase 136 — plant context bundle starter', () => {
  it('includes How is this grow doing chip with crop_cycle_id', () => {
    const starters = buildZoneGrowStripStarters({
      zone: { id: 3, name: 'Veg Room' },
      activeCycle: { id: 9, name: 'Veg canopy (18/6)', current_stage: 'late_veg', crop_key: 'chrysanthemum' },
      farmId: 1,
    })
    const chip = starters.find((s) => s.id === 'how-is-grow')
    expect(chip).toBeTruthy()
    expect(chip.label).toMatch(/how is this grow/i)
    expect(chip.message).toContain('plant_context_bundle')
    expect(chip.contextRef.crop_cycle_id).toBe(9)
    expect(chip.contextRef.id).toBe(3)
  })
})
