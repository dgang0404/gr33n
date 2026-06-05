/**
 * Farmer-facing sensor type labels (Phase 40 / farmer-vocabulary.md).
 */

const LABELS = {
  humidity: 'Humidity',
  rh: 'Humidity',
  air_temp: 'Air temperature',
  temperature: 'Temperature',
  temp: 'Temperature',
  co2: 'CO₂',
  vpd: 'VPD',
  dew_point: 'Dew point',
  ec: 'EC',
  ph: 'pH',
  soil_moisture: 'Soil moisture',
  moisture: 'Moisture',
  water_level: 'Water level',
  flow_rate: 'Flow rate',
  lux: 'Light level',
  light_level: 'Light level',
  par: 'PAR',
  par_umol: 'PAR',
  ppfd: 'PPFD',
}

/**
 * @param {string|undefined|null} sensorType
 * @returns {string}
 */
export function sensorTypeLabel(sensorType) {
  const t = String(sensorType || '').trim().toLowerCase()
  if (!t) return 'Sensor'
  if (LABELS[t]) return LABELS[t]
  return t.replace(/_/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase())
}
