import { describe, it, expect } from 'vitest'
import { buildSymptomGrowStarters, buildZoneGrowStripStarters } from '../lib/guardianStarters.js'

describe('Phase 106 — symptom Guardian starters', () => {
  it('buildSymptomGrowStarters includes lookup_crop_symptoms', () => {
    const starters = buildSymptomGrowStarters({
      zone: { id: 3, name: 'Flower Room' },
      activeCycle: { id: 9, current_stage: 'early_flower' },
      cropKey: 'tomato',
      cropDisplayName: 'Tomato',
    })
    expect(starters.length).toBeGreaterThanOrEqual(2)
    expect(starters[0].id).toBe('whats-wrong')
    expect(starters[0].message).toContain('lookup_crop_symptoms')
    expect(starters[1].message).toContain('Yellow leaves')
  })

  it('buildZoneGrowStripStarters prepends symptom chips when crop_key present', () => {
    const starters = buildZoneGrowStripStarters({
      zone: { id: 1, name: 'Veg' },
      activeCycle: { id: 2, current_stage: 'early_veg', crop_key: 'tomato', catalog_display_name: 'Tomato' },
      farmId: 1,
    })
    expect(starters.some((s) => s.id === 'whats-wrong')).toBe(true)
    expect(starters.some((s) => s.id === 'yellow-leaves')).toBe(true)
  })
})
