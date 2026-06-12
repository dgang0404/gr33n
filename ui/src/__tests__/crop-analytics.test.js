import { describe, it, expect } from 'vitest'
import { groupCyclesByCropKey, cyclePickerLabel } from '../lib/cropAnalytics.js'
import { lastHarvestedCycleWithCropKey, buildPostHarvestCompareRoute } from '../lib/growHub.js'

describe('cropAnalytics (Phase 104)', () => {
  it('groupCyclesByCropKey buckets by catalog crop_key', () => {
    const cycles = [
      { id: 1, crop_key: 'cannabis', catalog_display_name: 'Cannabis', name: 'Run A' },
      { id: 2, crop_key: 'cannabis', name: 'Run B' },
      { id: 3, crop_key: 'tomato', catalog_display_name: 'Tomato', name: 'Tomato 1' },
      { id: 4, name: 'Legacy' },
    ]
    const groups = groupCyclesByCropKey(cycles)
    expect(groups).toHaveLength(3)
    expect(groups[0].key).toBe('cannabis')
    expect(groups[0].cycles).toHaveLength(2)
    expect(groups.find((g) => g.key === 'tomato')?.cycles).toHaveLength(1)
    expect(groups.find((g) => g.key === '')?.cycles).toHaveLength(1)
  })

  it('cyclePickerLabel includes crop and batch', () => {
    const label = cyclePickerLabel({
      id: 5,
      name: 'Flower run',
      crop_key: 'cannabis',
      catalog_display_name: 'Cannabis',
      batch_label: 'OG',
    })
    expect(label).toContain('Flower run')
    expect(label).toContain('Cannabis')
    expect(label).toContain('OG')
  })

  it('buildPostHarvestCompareRoute prefers same crop_key over zone', () => {
    const cycles = [
      { id: 10, crop_key: 'cannabis', is_active: false, zone_id: 1 },
      { id: 9, crop_key: 'cannabis', is_active: false, zone_id: 2 },
      { id: 8, crop_key: 'tomato', is_active: false, zone_id: 1 },
    ]
    const route = buildPostHarvestCompareRoute(1, cycles, 10, 1)
    expect(route.query.ids).toBe('10,9')
  })

  it('lastHarvestedCycleWithCropKey ignores other crops', () => {
    const cycles = [
      { id: 3, crop_key: 'cannabis', is_active: false },
      { id: 2, crop_key: 'tomato', is_active: false },
    ]
    const prior = lastHarvestedCycleWithCropKey(cycles, 'cannabis', 3)
    expect(prior).toBeNull()
  })
})
