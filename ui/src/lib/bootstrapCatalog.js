/**
 * Phase 91 — bootstrap template catalog from GET /platform/bootstrap-templates.
 */
import {
  FALLBACK_BOOTSTRAP_CATALOG,
  indexBootstrapCatalog,
} from './bootstrapCatalog.fallback.js'

/**
 * @typedef {object} BootstrapTemplate
 * @property {string} template_key
 * @property {string} label
 * @property {string} [short_label]
 * @property {string} [tagline]
 * @property {string} summary_title
 * @property {string[]} summary_bullets
 * @property {string[]} [module_hints]
 * @property {string} [icon]
 * @property {boolean} [recommended]
 * @property {boolean} [wizard_primary]
 * @property {string} [related_commons_slug]
 * @property {number} [sort_order]
 */

/** @typedef {ReturnType<typeof indexBootstrapCatalog>} BootstrapCatalog */

let cached = /** @type {BootstrapCatalog|null} */ (null)
/** @type {Promise<BootstrapCatalog>|null} */
let loadPromise = null

/**
 * @param {object|null|undefined} payload
 * @returns {BootstrapCatalog}
 */
export function normalizeBootstrapCatalog(payload) {
  if (!payload?.templates?.length) {
    return FALLBACK_BOOTSTRAP_CATALOG
  }
  return indexBootstrapCatalog(payload)
}

/**
 * @param {{ get: (url: string) => Promise<{ data: object }> }} api
 */
export async function loadBootstrapCatalog(api) {
  if (cached) return cached
  if (loadPromise) return loadPromise
  loadPromise = api
    .get('/platform/bootstrap-templates')
    .then(({ data }) => {
      cached = normalizeBootstrapCatalog(data)
      return cached
    })
    .catch(() => {
      cached = FALLBACK_BOOTSTRAP_CATALOG
      return cached
    })
    .finally(() => {
      loadPromise = null
    })
  return loadPromise
}

/** @returns {BootstrapCatalog} */
export function getBootstrapCatalog() {
  return cached || FALLBACK_BOOTSTRAP_CATALOG
}

export function resetBootstrapCatalogCache() {
  cached = null
  loadPromise = null
}
