/**
 * Phase 89 — lighting preset chips from GET /lighting-programs/presets (Guardian PresetList parity).
 */

/** UI-only first option for zone wizard — not an API preset. */
export const LIGHTING_PRESET_SKIP = { key: '', label: 'Skip for now', onHours: null }

/**
 * @param {{ key?: string, name?: string, on_hours?: number, off_hours?: number }} row
 */
export function mapApiPreset(row) {
  return {
    key: row.key,
    label: row.name || row.key,
    onHours: row.on_hours ?? null,
    offHours: row.off_hours ?? null,
  }
}

let cachedPresets = null
/** @type {Promise<object[]>|null} */
let loadPromise = null

/**
 * Fetch and cache lighting presets for the session.
 * @param {{ get: (url: string) => Promise<{ data: object[] }> }} api
 */
export async function loadLightingPresets(api) {
  if (cachedPresets) return cachedPresets
  if (loadPromise) return loadPromise
  loadPromise = api
    .get('/lighting-programs/presets')
    .then(({ data }) => {
      cachedPresets = (data ?? [])
        .map(mapApiPreset)
        .sort((a, b) => String(a.key).localeCompare(String(b.key)))
      return cachedPresets
    })
    .finally(() => {
      loadPromise = null
    })
  return loadPromise
}

/** @param {object[]} presets @param {string} key */
export function findLightingPreset(presets, key) {
  if (!key) return null
  return (presets || []).find((p) => p.key === key) ?? null
}

/** @param {object[]} presets @param {string} key */
export function presetDisplayLabel(presets, key) {
  const p = findLightingPreset(presets, key)
  return p?.label || key || ''
}

/** Test helper — reset in-memory cache between Vitest cases. */
export function resetLightingPresetsCache() {
  cachedPresets = null
  loadPromise = null
}
