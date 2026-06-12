import { describe, it, expect, beforeEach } from 'vitest'
import { loadBootstrapCatalog, resetBootstrapCatalogCache, getBootstrapCatalog } from '../lib/bootstrapCatalog.js'
import { BOOTSTRAP_TEMPLATE_KEYS } from '../lib/bootstrapCatalog.fallback.js'

describe('Phase 91 — bootstrap catalog loader', () => {
  beforeEach(() => {
    resetBootstrapCatalogCache()
  })

  it('loads templates from API', async () => {
    const api = {
      get: async () => ({
        data: {
          templates: [
            {
              template_key: 'jadam_indoor_photoperiod_v1',
              label: 'Indoor photoperiod starter (v1)',
              short_label: 'Indoor photoperiod v1',
              summary_title: 'Included',
              summary_bullets: ['Four zones'],
              wizard_primary: true,
              recommended: true,
              sort_order: 10,
            },
          ],
        },
      }),
    }
    await loadBootstrapCatalog(api)
    const cat = getBootstrapCatalog()
    expect(cat.templates).toHaveLength(1)
    expect(cat.summariesByKey.jadam_indoor_photoperiod_v1.bullets).toContain('Four zones')
  })

  it('falls back when API fails', async () => {
    const api = { get: async () => { throw new Error('offline') } }
    await loadBootstrapCatalog(api)
    expect(getBootstrapCatalog().byKey[BOOTSTRAP_TEMPLATE_KEYS.JADAM_INDOOR_PHOTOPERIOD_V1]).toBeTruthy()
  })
})
