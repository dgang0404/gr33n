import { describe, it, expect, beforeEach } from 'vitest'
import {
  loadDomainEnums,
  growthStageValues,
  resetDomainEnumsCache,
  wizardZoneTypes,
  adminZoneTypes,
  greenhouseCoverTypes,
} from '../lib/domainEnums.js'
import { FALLBACK_GROWTH_STAGE_VALUES } from '../lib/domainEnums.fallback.js'
import { GROWTH_STAGES } from '../lib/growHub.js'

describe('Phase 88 — domain enums loader', () => {
  beforeEach(() => {
    resetDomainEnumsCache()
  })

  it('fallback growth stages include transition and flush', () => {
    expect(FALLBACK_GROWTH_STAGE_VALUES).toHaveLength(11)
    expect(FALLBACK_GROWTH_STAGE_VALUES).toContain('transition')
    expect(FALLBACK_GROWTH_STAGE_VALUES).toContain('flush')
    expect(GROWTH_STAGES).toEqual(FALLBACK_GROWTH_STAGE_VALUES)
  })

  it('loads and caches domain enums from API', async () => {
    let calls = 0
    const api = {
      get: async () => {
        calls += 1
        return {
          data: {
            growth_stages: [{ value: 'transition', label: 'transition' }],
            cost_categories: [{ value: 'miscellaneous', label: 'miscellaneous' }],
          },
        }
      },
    }
    const first = await loadDomainEnums(api)
    const second = await loadDomainEnums(api)
    expect(first).toBe(second)
    expect(calls).toBe(1)
    expect(growthStageValues(first)).toEqual(['transition'])
  })

  it('fallback zone vocabulary includes wizard subset and legacy admin types', () => {
    expect(adminZoneTypes()).toHaveLength(8)
    expect(adminZoneTypes().some((z) => z.value === 'veg' && !z.wizard_visible)).toBe(true)
    expect(wizardZoneTypes()).toHaveLength(3)
    expect(wizardZoneTypes().map((z) => z.value)).toEqual(['indoor', 'greenhouse', 'outdoor'])
    expect(greenhouseCoverTypes().map((c) => c.value)).toContain('film')
  })
})
