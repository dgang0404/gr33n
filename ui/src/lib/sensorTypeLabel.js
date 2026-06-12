/**
 * Farmer-facing sensor type labels (Phase 40 / Phase 90 registry).
 */
import { getDeviceTaxonomy } from './deviceTaxonomy.js'

/**
 * @param {string|undefined|null} sensorType
 * @returns {string}
 */
export function sensorTypeLabel(sensorType) {
  const t = String(sensorType || '').trim().toLowerCase()
  if (!t) return 'Sensor'
  const row = getDeviceTaxonomy().sensorsByKey[t]
  if (row?.display_label) return row.display_label
  return t.replace(/_/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase())
}
