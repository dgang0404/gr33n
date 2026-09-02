/**
 * Phase 217 WS4/WS5/WS6 — qualitative sensor reading scale.
 *
 * Replaces the old 3-tier ok/warn/danger classification with a 5-tier
 * critical/bad/mid/good/awesome scale that doesn't require memorizing raw
 * ranges. Pure functions only — SensorTile.vue supplies whatever
 * low/high/target/sensitivity it resolved from (per-sensor override, a
 * crop-library-derived stage row, or hardcoded defaults).
 */

export const QUALITY_TIERS = ['critical', 'bad', 'mid', 'good', 'awesome']

export const QUALITY_LABEL = {
  critical: 'Critical',
  bad: 'Bad',
  mid: 'Okay',
  good: 'Good',
  awesome: 'Awesome',
}

/** Sensitivity multiplies how wide the good/awesome bands are around the target. */
const SENSITIVITY_FACTOR = { narrow: 0.7, normal: 1, wide: 1.3 }

/**
 * @param {number|null|undefined} value
 * @param {{ low: number, high: number, target?: number }} range
 * @param {'narrow'|'normal'|'wide'} [sensitivity]
 * @returns {'critical'|'bad'|'mid'|'good'|'awesome'|'unknown'}
 */
export function classifySensorQuality(value, range, sensitivity = 'normal') {
  if (value == null || Number.isNaN(Number(value))) return 'unknown'
  const { low, high } = range || {}
  if (low == null || high == null || high <= low) return 'unknown'
  const target = range.target != null ? range.target : (low + high) / 2
  const factor = SENSITIVITY_FACTOR[sensitivity] ?? 1
  const halfRange = (high - low) / 2

  // Bands widths shrink (narrow/stricter) or grow (wide/lenient) with sensitivity.
  const awesomeBand = halfRange * 0.15 * factor
  const goodBand = halfRange * 0.5 * factor
  const badMargin = (high - low) * 0.25 * factor

  const v = Number(value)
  if (v < low - badMargin || v > high + badMargin) return 'critical'
  if (v < low || v > high) return 'bad'

  const distFromTarget = Math.abs(v - target)
  if (distFromTarget <= awesomeBand) return 'awesome'
  if (distFromTarget <= goodBand) return 'good'
  return 'mid'
}

/**
 * Sensitivity precedence: per-sensor override > farm-wide default > 'normal'.
 * @param {{ quality_sensitivity?: string|null }} sensor
 * @param {{ meta_data?: { default_sensor_sensitivity?: string } }} [farm]
 */
export function resolveSensitivity(sensor, farm) {
  if (sensor?.quality_sensitivity) return sensor.quality_sensitivity
  const farmDefault = farm?.meta_data?.default_sensor_sensitivity
  if (farmDefault) return farmDefault
  return 'normal'
}

/** Maps a sensor_type to the crop_library-shaped stage row fields that describe its ideal range. */
const SENSOR_TYPE_TO_STAGE_FIELDS = {
  ec: ['ec_min', 'ec_target', 'ec_max'],
  ph: ['ph_min', null, 'ph_max'],
  temperature: ['temp_min_c', null, 'temp_max_c'],
  temp: ['temp_min_c', null, 'temp_max_c'],
  air_temp: ['temp_min_c', null, 'temp_max_c'],
  humidity: ['rh_min_pct', null, 'rh_max_pct'],
  rh: ['rh_min_pct', null, 'rh_max_pct'],
}

/**
 * Derives { low, high, target } from a crop_library/effective-profile stage
 * row (e.g. `{ ec_min, ec_target, ec_max, ph_min, ph_max, temp_min_c, ... }`)
 * for the given sensor_type. Returns null when there's no active crop cycle
 * stage row, or the sensor_type has no field mapping (e.g. PAR, CO2, DO) —
 * callers fall back to hardcoded defaults in that case.
 *
 * @param {string} sensorType
 * @param {object|null|undefined} stageRow
 */
export function idealRangeFromCropStage(sensorType, stageRow) {
  if (!stageRow) return null
  const fields = SENSOR_TYPE_TO_STAGE_FIELDS[sensorType]
  if (!fields) return null
  const [lowKey, targetKey, highKey] = fields
  const low = stageRow[lowKey]
  const high = stageRow[highKey]
  if (low == null || high == null) return null
  const target = targetKey ? stageRow[targetKey] : null
  return { low: Number(low), high: Number(high), target: target != null ? Number(target) : undefined }
}
