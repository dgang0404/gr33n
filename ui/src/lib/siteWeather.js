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
 * Parse lat/lng from farm.location_gis GeoJSON, meta_data lat/lon, or null.
 * PostGIS often returns location_gis as EWKB/base64 — meta_data is the reliable UI source.
 * @param {object} farm
 */
export function parseFarmCoordinates(farm) {
  const gis = farm?.location_gis
  if (typeof gis === 'object' && gis?.type === 'Point' && Array.isArray(gis.coordinates)) {
    const longitude = Number(gis.coordinates[0])
    const latitude = Number(gis.coordinates[1])
    if (Number.isFinite(latitude) && Number.isFinite(longitude)) {
      return { latitude, longitude }
    }
  }
  if (typeof gis === 'string') {
    try {
      const parsed = JSON.parse(gis)
      if (parsed?.type === 'Point' && Array.isArray(parsed.coordinates)) {
        const longitude = Number(parsed.coordinates[0])
        const latitude = Number(parsed.coordinates[1])
        if (Number.isFinite(latitude) && Number.isFinite(longitude)) {
          return { latitude, longitude }
        }
      }
    } catch {
      /* EWKB hex / non-JSON — fall through to meta */
    }
  }
  const meta = farm?.meta_data && typeof farm.meta_data === 'object' ? farm.meta_data : null
  if (meta) {
    const latitude = Number(meta.latitude)
    const longitude = Number(meta.longitude)
    if (Number.isFinite(latitude) && Number.isFinite(longitude)) {
      return { latitude, longitude }
    }
  }
  return { latitude: null, longitude: null }
}

/**
 * Apply N/S/E/W hemisphere to a positive magnitude.
 * @param {number} value
 * @param {string} hemisphere
 */
function applyHemisphere(value, hemisphere) {
  const n = Math.abs(Number(value))
  if (!Number.isFinite(n)) return null
  const h = String(hemisphere || '').trim().toUpperCase()
  if (h === 'S' || h === 'W') return -n
  if (h === 'N' || h === 'E') return n
  return Number(value)
}

function validCoords(lat, lon) {
  return Number.isFinite(lat) && Number.isFinite(lon)
    && lat >= -90 && lat <= 90
    && lon >= -180 && lon <= 180
}

/**
 * Parse coordinates pasted from Google Maps or similar.
 * Applies N/S/E/W only when direction letters are present — safe worldwide.
 * @param {string} text
 * @returns {{ ok: true, latitude: number, longitude: number } | { ok: false, error: string }}
 */
export function parseMapsCoordinates(text) {
  const raw = String(text || '').trim()
  if (!raw) {
    return { ok: false, error: 'Paste coordinates from Google Maps' }
  }

  const latLabeled = raw.match(
    /(?:lat(?:itude)?\s*:?\s*)(-?\d+(?:\.\d+)?)\s*°?\s*([NnSs])\b/i,
  )
  const lonLabeled = raw.match(
    /(?:lon(?:g(?:itude)?)?\s*:?\s*)(-?\d+(?:\.\d+)?)\s*°?\s*([EeWw])\b/i,
  )
  if (latLabeled && lonLabeled) {
    const latitude = applyHemisphere(latLabeled[1], latLabeled[2])
    const longitude = applyHemisphere(lonLabeled[1], lonLabeled[2])
    if (validCoords(latitude, longitude)) {
      return { ok: true, latitude, longitude }
    }
  }

  const latLoose = [...raw.matchAll(/(-?\d+(?:\.\d+)?)\s*°?\s*([NnSs])\b/g)]
  const lonLoose = [...raw.matchAll(/(-?\d+(?:\.\d+)?)\s*°?\s*([EeWw])\b/g)]
  if (latLoose.length && lonLoose.length) {
    const latitude = applyHemisphere(latLoose[0][1], latLoose[0][2])
    const longitude = applyHemisphere(lonLoose[0][1], lonLoose[0][2])
    if (validCoords(latitude, longitude)) {
      return { ok: true, latitude, longitude }
    }
  }

  const signedPair = raw.match(/(-?\d+(?:\.\d+)?)\s*[,;\s]\s*(-?\d+(?:\.\d+)?)/)
  if (signedPair) {
    const a = Number(signedPair[1])
    const b = Number(signedPair[2])
    if (validCoords(a, b) && Math.abs(a) <= 90) {
      return { ok: true, latitude: a, longitude: b }
    }
    if (validCoords(b, a) && Math.abs(b) <= 90) {
      return { ok: true, latitude: b, longitude: a }
    }
  }

  return {
    ok: false,
    error: 'Could not read coordinates — paste with N/S and E/W (e.g. 40.89° N, 81.41° W) or a signed pair like 40.89, -81.41',
  }
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
