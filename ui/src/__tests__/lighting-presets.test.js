import { describe, it, expect, beforeEach } from 'vitest'
import {
  mapApiPreset,
  loadLightingPresets,
  findLightingPreset,
  presetDisplayLabel,
  resetLightingPresetsCache,
} from '../lib/lightingPresets.js'

describe('Phase 89 — lighting presets loader', () => {
  beforeEach(() => {
    resetLightingPresetsCache()
  })

  it('maps API rows to chip shape', () => {
    expect(mapApiPreset({ key: 'veg_18_6', name: 'Veg 18/6', on_hours: 18, off_hours: 6 })).toEqual({
      key: 'veg_18_6',
      label: 'Veg 18/6',
      onHours: 18,
      offHours: 6,
    })
  })

  it('loads and caches presets from API', async () => {
    let calls = 0
    const api = {
      get: async () => {
        calls += 1
        return {
          data: [
            { key: 'peas_22_2', name: 'Peas 22/2', on_hours: 22, off_hours: 2 },
            { key: 'veg_18_6', name: 'Veg 18/6', on_hours: 18, off_hours: 6 },
          ],
        }
      },
    }
    const first = await loadLightingPresets(api)
    const second = await loadLightingPresets(api)
    expect(first).toHaveLength(2)
    expect(second).toBe(first)
    expect(calls).toBe(1)
    expect(findLightingPreset(first, 'peas_22_2')?.onHours).toBe(22)
    expect(presetDisplayLabel(first, 'veg_18_6')).toBe('Veg 18/6')
    expect(presetDisplayLabel(first, 'missing')).toBe('missing')
  })
})
