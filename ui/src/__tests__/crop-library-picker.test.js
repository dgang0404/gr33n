/**
 * Phase 82 WS4f — crop library picker unit tests.
 */
import { describe, it, expect } from 'vitest'
import {
  filterPickerGroups,
  findPickerItemByProfileId,
  itemMatchesQuery,
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
  it('filters by alias and partial name', () => {
    const groups = filterPickerGroups(samplePicker, 'toma')
    expect(groups).toHaveLength(1)
    expect(groups[0].items).toHaveLength(1)
    expect(groups[0].items[0].crop_key).toBe('tomato')
  })

  it('finds item by profile id', () => {
    const item = findPickerItemByProfileId(samplePicker, 1)
    expect(item?.crop_key).toBe('tomato')
  })

  it('matches search terms on item', () => {
    expect(itemMatchesQuery(samplePicker.groups[0].items[0], 'toma')).toBe(true)
    expect(itemMatchesQuery(samplePicker.groups[0].items[1], 'tomato')).toBe(false)
  })
})
