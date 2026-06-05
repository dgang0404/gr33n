/**
 * Phase 42 — comfort band status and validation (shared hub + zone editor).
 */

import { PLANT_NEEDS, sensorPlantNeed } from './plantNeeds.js'
import { sensorTypeLabel } from './sensorTypeLabel.js'

/** @typedef {'missing'|'ok'|'out_of_range'} ComfortBandStatus */

/**
 * @param {object|null|undefined} setpoint
 * @param {object|null|undefined} reading
 * @returns {ComfortBandStatus}
 */
export function comfortBandStatus(setpoint, reading) {
  if (!setpointHasValues(setpoint)) return 'missing'
  const val = readingValue(reading)
  if (val == null) return 'ok'
  if (setpoint.min_value != null && val < Number(setpoint.min_value)) return 'out_of_range'
  if (setpoint.max_value != null && val > Number(setpoint.max_value)) return 'out_of_range'
  return 'ok'
}

/**
 * @param {object|null|undefined} setpoint
 */
export function setpointHasValues(setpoint) {
  if (!setpoint) return false
  return setpoint.min_value != null || setpoint.ideal_value != null || setpoint.max_value != null
}

/**
 * @param {object|null|undefined} reading
 */
export function readingValue(reading) {
  if (!reading) return null
  const raw = reading.value_raw ?? reading.value_normalized
  if (raw == null || raw === '') return null
  const n = Number(raw)
  return Number.isFinite(n) ? n : null
}

/**
 * Zone-level setpoint row (no crop_cycle_id; prefer stage-less band).
 * @param {object[]} setpoints
 * @param {number} zoneId
 * @param {string} sensorType
 */
export function zoneSetpointForType(setpoints, zoneId, sensorType) {
  const matches = (setpoints || []).filter(
    (sp) => Number(sp.zone_id) === Number(zoneId)
      && !sp.crop_cycle_id
      && sp.sensor_type === sensorType,
  )
  return matches.find((sp) => !sp.stage) || matches[0] || null
}

/**
 * Climate sensor types present in a zone.
 * @param {object[]} sensors
 * @param {number} zoneId
 */
export function climateSensorTypesForZone(sensors, zoneId) {
  const types = new Set(
    (sensors || [])
      .filter((s) => Number(s.zone_id) === Number(zoneId) && sensorPlantNeed(s.sensor_type) === PLANT_NEEDS.air)
      .map((s) => s.sensor_type)
      .filter(Boolean),
  )
  return [...types]
}

/**
 * @param {object} params
 * @returns {Array<{ sensorType: string, label: string, status: ComfortBandStatus, setpoint: object|null, reading: object|null, liveValue: number|null }>}
 */
export function buildZoneComfortBands({
  zoneId,
  sensors = [],
  setpoints = [],
  readings = {},
}) {
  const types = climateSensorTypesForZone(sensors, zoneId)
  const zoneSensors = (sensors || []).filter((s) => Number(s.zone_id) === Number(zoneId))

  return types.map((sensorType) => {
    const setpoint = zoneSetpointForType(setpoints, zoneId, sensorType)
    const sensor = zoneSensors.find((s) => s.sensor_type === sensorType)
    const reading = sensor?.id != null ? readings[sensor.id] : null
    const liveValue = readingValue(reading)
    return {
      sensorType,
      label: sensorTypeLabel(sensorType),
      status: comfortBandStatus(setpoint, reading),
      setpoint,
      reading,
      liveValue,
    }
  })
}

/**
 * Roll up band rows into a single zone status (worst wins).
 * @param {Array<{ status: ComfortBandStatus }>} bands
 * @returns {ComfortBandStatus|'no_sensors'}
 */
export function summarizeZoneComfortStatus(bands) {
  if (!bands?.length) return 'no_sensors'
  if (bands.some((b) => b.status === 'out_of_range')) return 'out_of_range'
  if (bands.some((b) => b.status === 'missing')) return 'missing'
  return 'ok'
}

export const COMFORT_STATUS_META = {
  ok: { label: 'In range', tone: 'ok' },
  missing: { label: 'Missing band', tone: 'warn' },
  out_of_range: { label: 'Out of range', tone: 'danger' },
  no_sensors: { label: 'No climate sensors', tone: 'muted' },
}

/**
 * @param {number|string|null|undefined} v
 */
export function parseComfortNumber(v) {
  if (v === '' || v == null || Number.isNaN(Number(v))) return null
  return Number(v)
}

/**
 * @param {object} payload
 * @returns {string} empty when valid
 */
export function validateComfortBandPayload(payload) {
  if (!payload.sensor_type) return 'Sensor type is required'
  if (payload.min_value != null && payload.max_value != null && payload.min_value > payload.max_value) {
    return 'Too low must be ≤ too high'
  }
  if (payload.ideal_value != null) {
    if (payload.min_value != null && payload.ideal_value < payload.min_value) {
      return 'Just right must be ≥ too low'
    }
    if (payload.max_value != null && payload.ideal_value > payload.max_value) {
      return 'Just right must be ≤ too high'
    }
  }
  return ''
}
