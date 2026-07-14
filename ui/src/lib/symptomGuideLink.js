/**
 * Phase 183 WS1 — contextual links into the symptom guide with crop pre-selected.
 */

/** @param {string | null | undefined} cropKey */
export function symptomGuideRoute(cropKey) {
  const key = String(cropKey || '').trim()
  if (!key) return null
  return { path: '/symptom-guide', query: { crop_key: key } }
}

/** @param {string | null | undefined} cropKey */
export function symptomGuideLabel(cropKey) {
  const key = String(cropKey || '').trim()
  if (!key) return ''
  const pretty = key.replace(/[-_]/g, ' ')
  return `Symptoms for ${pretty}`
}
