import { describe, it, expect } from 'vitest'
import { defaultNaturalFarmingTab, NATURAL_FARMING_DEFAULT_TAB } from '../lib/naturalFarmingStudio.js'

describe('naturalFarmingStudio', () => {
  it('defaults to Make a batch tab', () => {
    expect(NATURAL_FARMING_DEFAULT_TAB).toBe('batch')
    expect(defaultNaturalFarmingTab(14)).toBe('batch')
    expect(defaultNaturalFarmingTab(0)).toBe('batch')
  })
})
