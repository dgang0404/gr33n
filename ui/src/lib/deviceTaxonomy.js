/**
 * Phase 90 — device taxonomy from GET /platform/device-taxonomy.
 */
import { FALLBACK_TAXONOMY, indexTaxonomy } from './deviceTaxonomy.fallback.js'

let cached = null
/** @type {Promise<object>|null} */
let loadPromise = null

/**
 * @param {{ get: (url: string) => Promise<{ data: object }> }} api
 */
export async function loadDeviceTaxonomy(api) {
  if (cached) return cached
  if (loadPromise) return loadPromise
  loadPromise = api
    .get('/platform/device-taxonomy')
    .then(({ data }) => {
      cached = indexTaxonomy(data)
      return cached
    })
    .catch(() => {
      cached = FALLBACK_TAXONOMY
      return cached
    })
    .finally(() => {
      loadPromise = null
    })
  return loadPromise
}

/** @returns {object} */
export function getDeviceTaxonomy() {
  return cached || FALLBACK_TAXONOMY
}

/** @returns {{ value: string, label: string }[]} */
export function wiringSourceOptions() {
  return getDeviceTaxonomy().wiring_source_options || FALLBACK_TAXONOMY.wiring_source_options
}

export function resetDeviceTaxonomyCache() {
  cached = null
  loadPromise = null
}
