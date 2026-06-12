/**
 * Phase 100 — crop library picker load with IndexedDB offline cache.
 */
import { buildPickerFallbackFromProfiles } from './cropLibraryPicker.js'
import {
  cacheCropPicker,
  getCachedCropPicker,
  isNetworkError,
  isStaleCatalogVersion,
} from './catalogCache.js'

/**
 * @param {number} farmId
 * @param {{ get: (url: string) => Promise<{ data: object }>, loadProfiles: () => Promise<object[]> }} deps
 */
export async function loadCropLibraryPickerWithCache(farmId, deps) {
  try {
    const r = await deps.get(`/farms/${farmId}/crop-library/picker`)
    const data = r.data
    await cacheCropPicker(farmId, data)
    return data
  } catch (e) {
    const status = e.response?.status
    if (status === 404) {
      const profiles = await deps.loadProfiles()
      const fallback = buildPickerFallbackFromProfiles(profiles)
      if (fallback.counts.with_targets > 0) {
        return { ...fallback, _degraded: true, _degradedReason: 'picker_404' }
      }
      throw e
    }
    if (isNetworkError(e)) {
      const cached = await getCachedCropPicker(farmId)
      if (cached?.groups?.length) {
        const meta = cached._cacheMeta || {}
        const catalogVersion = meta.catalog_version ?? cached.version
        const { _cacheMeta, ...picker } = cached
        return {
          ...picker,
          _offline: true,
          _offlineFetchedAt: meta.fetched_at,
          _stale: isStaleCatalogVersion(catalogVersion),
          _cachedCatalogVersion: catalogVersion,
        }
      }
    }
    throw e
  }
}
