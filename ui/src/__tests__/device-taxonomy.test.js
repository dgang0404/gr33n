import { describe, it, expect, beforeEach } from 'vitest'
import { loadDeviceTaxonomy, resetDeviceTaxonomyCache } from '../lib/deviceTaxonomy.js'
import { sensorPlantNeed, PLANT_NEEDS } from '../lib/plantNeeds.js'
import { sensorTypeLabel } from '../lib/sensorTypeLabel.js'

describe('Phase 90 — device taxonomy loader', () => {
  beforeEach(() => {
    resetDeviceTaxonomyCache()
  })

  it('loads temp_f as climate from API', async () => {
    const api = {
      get: async () => ({
        data: {
          sensors: [{ type_key: 'temp_f', device_class: 'sensor', plant_need: 'air', display_label: 'Temperature (°F)' }],
          actuators: [],
          wiring_source_options: [],
        },
      }),
    }
    await loadDeviceTaxonomy(api)
    expect(sensorPlantNeed('temp_f')).toBe(PLANT_NEEDS.air)
    expect(sensorTypeLabel('temp_f')).toBe('Temperature (°F)')
  })
})
