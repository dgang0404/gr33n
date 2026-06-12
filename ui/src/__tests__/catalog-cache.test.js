import { describe, it, expect, beforeEach } from 'vitest'
import {
  cacheCropPicker,
  getCachedCropPicker,
  cacheDomainEnums,
  getCachedDomainEnums,
  clearCatalogCache,
  isNetworkError,
  isStaleCatalogVersion,
  formatCacheDate,
} from '../lib/catalogCache.js'
import { loadCropLibraryPickerWithCache } from '../lib/cropLibraryLoader.js'

describe('Phase 100 — catalog cache', () => {
  beforeEach(async () => {
    await clearCatalogCache()
  })

  it('stores and retrieves crop picker by farm id', async () => {
    const picker = {
      version: 4,
      counts: { with_targets: 2, catalog_only: 0, total: 2 },
      groups: [{ key: 'fruiting', label: 'Fruiting', items: [{ crop_key: 'tomato', display_name: 'Tomato', has_targets: true, crop_profile_id: 1 }] }],
    }
    await cacheCropPicker(1, picker)
    const hit = await getCachedCropPicker(1)
    expect(hit?.groups).toHaveLength(1)
    expect(hit?.groups[0].items[0].crop_key).toBe('tomato')
  })

  it('stores domain enums', async () => {
    await cacheDomainEnums({ growth_stages: [{ value: 'seedling', label: 'seedling' }] })
    const hit = await getCachedDomainEnums()
    expect(hit?.enums.growth_stages).toHaveLength(1)
  })

  it('detects network errors', () => {
    expect(isNetworkError({ code: 'ERR_NETWORK' })).toBe(true)
    expect(isNetworkError({ response: { status: 404 } })).toBe(false)
  })

  it('flags stale cache when last known version is newer', () => {
    localStorage.setItem('gr33n_last_catalog_version', '5')
    expect(isStaleCatalogVersion(4)).toBe(true)
    expect(isStaleCatalogVersion(5)).toBe(false)
  })

  it('formats cache timestamps', () => {
    expect(formatCacheDate('2026-06-12T12:00:00.000Z')).toMatch(/2026/)
  })
})

describe('Phase 100 — crop library loader', () => {
  beforeEach(async () => {
    await clearCatalogCache()
  })

  it('serves cached picker on network failure', async () => {
    await cacheCropPicker(1, {
      version: 4,
      counts: { with_targets: 1, catalog_only: 0, total: 1 },
      groups: [{
        key: 'fruiting',
        label: 'Fruiting',
        items: [{ crop_key: 'tomato', display_name: 'Tomato', has_targets: true, crop_profile_id: 9 }],
      }],
    })

    const result = await loadCropLibraryPickerWithCache(1, {
      get: async () => {
        throw { code: 'ERR_NETWORK', message: 'Network Error' }
      },
      loadProfiles: async () => [],
    })

    expect(result._offline).toBe(true)
    expect(result.groups[0].items[0].crop_key).toBe('tomato')
  })

  it('404 returns degraded profile fallback, not offline cache', async () => {
    await cacheCropPicker(1, {
      version: 99,
      counts: { with_targets: 50, catalog_only: 0, total: 50 },
      groups: [{ key: 'fruiting', label: 'Full catalog', items: [{ crop_key: 'cached', display_name: 'Cached', has_targets: true, crop_profile_id: 1 }] }],
    })

    const result = await loadCropLibraryPickerWithCache(1, {
      get: async () => {
        throw { response: { status: 404, data: { error: 'not found' } } }
      },
      loadProfiles: async () => [{
        id: 2,
        crop_key: 'tomato',
        display_name: 'Tomato',
        category: 'fruiting',
        is_builtin: true,
      }],
    })

    expect(result._degraded).toBe(true)
    expect(result._offline).toBeFalsy()
    expect(result.groups[0].label).toMatch(/knowledge base/i)
  })

  it('caches successful picker responses', async () => {
    await loadCropLibraryPickerWithCache(1, {
      get: async () => ({
        data: {
          version: 3,
          counts: { with_targets: 1, catalog_only: 0, total: 1 },
          groups: [{ key: 'herb', label: 'Herbs', items: [{ crop_key: 'basil', display_name: 'Basil', has_targets: true, crop_profile_id: 3 }] }],
        },
      }),
      loadProfiles: async () => [],
    })

    const cached = await getCachedCropPicker(1)
    expect(cached?.groups[0].items[0].crop_key).toBe('basil')
  })
})
