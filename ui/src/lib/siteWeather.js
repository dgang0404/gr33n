/**
 * Phase 66 — farm site weather (offline solar + optional readings).
 */
import api from '../api'

/**
 * @param {number} farmId
 * @param {string} [date] YYYY-MM-DD
 */
export async function fetchSiteWeather(farmId, date) {
  const params = date ? { date } : {}
  const r = await api.get(`/farms/${farmId}/site-weather`, { params })
  return r.data
}

/**
 * Parse lat/lng from farm.location_gis GeoJSON or null.
 * @param {object} farm
 */
export function parseFarmCoordinates(farm) {
  const gis = farm?.location_gis
  if (!gis) return { latitude: null, longitude: null }
  if (typeof gis === 'object' && gis.type === 'Point' && Array.isArray(gis.coordinates)) {
    return { longitude: gis.coordinates[0], latitude: gis.coordinates[1] }
  }
  return { latitude: null, longitude: null }
}

/**
 * @param {object} farm
 */
export function parseFarmElevationM(farm) {
  const m = farm?.meta_data
  if (!m || typeof m !== 'object') return null
  const v = m.elevation_m
  return v == null || v === '' ? null : Number(v)
}

/**
 * @param {object} siteWeatherResponse from fetchSiteWeather
 */
export function daylightChipFromSiteWeather(siteWeatherResponse) {
  const solar = siteWeatherResponse?.solar
  if (!solar?.daylength_hours) return null
  return {
    id: 'daylight-hours',
    icon: '☀️',
    label: 'Daylight today',
    value: `${solar.daylength_hours} h`,
    tone: 'neutral',
  }
}
