/**
 * Phase 82 WS4f — crop library picker unit tests.
 */
import { describe, it, expect } from 'vitest'
import {
  filterPickerGroups,
  findPickerItemByProfileId,
  itemMatchesQuery,
  buildPickerFallbackFromProfiles,
  formatStageTargetLine,
} from '../lib/cropLibraryPicker.js'

const samplePicker = {
  groups: [
    {
      key: 'fruiting',
      label: 'Fruiting',
      items: [
        {
          crop_key: 'tomato',
          display_name: 'Tomato',
          crop_profile_id: 1,
          has_targets: true,
          search_terms: ['tomato', 'toma'],
        },
        {
          crop_key: 'zucchini',
          display_name: 'Zucchini',
          has_targets: false,
          cousin_label: 'Cucumber',
          search_terms: ['zucchini'],
        },
      ],
    },
  ],
}

describe('cropLibraryPicker', () => {
  it('filters by prefix on name and aliases', () => {
    const groups = filterPickerGroups(samplePicker, 'toma')
    expect(groups).toHaveLength(1)
    expect(groups[0].items).toHaveLength(1)
    expect(groups[0].items[0].crop_key).toBe('tomato')
    expect(filterPickerGroups(samplePicker, 'ato')).toHaveLength(0)
  })

  it('finds item by profile id', () => {
    const item = findPickerItemByProfileId(samplePicker, 1)
    expect(item?.crop_key).toBe('tomato')
  })

  it('builds fallback picker from builtin profiles', () => {
    const p = buildPickerFallbackFromProfiles([
      { id: 2, crop_key: 'tomato', display_name: 'Tomato', is_builtin: true },
    ])
    expect(p.counts.with_targets).toBe(1)
    expect(p.groups[0].items[0].crop_profile_id).toBe(2)
  })

  it('formats stage target line', () => {
    const line = formatStageTargetLine({
      stage: 'early_flower',
      ec_min: 1.2,
      ec_max: 1.8,
      ec_target: 1.5,
      dli_target: 35,
      photoperiod_hrs: 12,
    })
    expect(line).toContain('early flower')
    expect(line).toContain('mS/cm')
    expect(line).toContain('DLI')
  })

  it('matches search terms on item by prefix only', () => {
    expect(itemMatchesQuery(samplePicker.groups[0].items[0], 'toma')).toBe(true)
    expect(itemMatchesQuery(samplePicker.groups[0].items[0], 'ato')).toBe(false)
    expect(itemMatchesQuery(samplePicker.groups[0].items[1], 'tomato')).toBe(false)
  })
})
