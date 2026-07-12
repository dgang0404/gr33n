/**
 * Phase 178 — online forecast status labels for Today + Settings.
 */

/** @typedef {'disabled'|'no_coords'|'connected'|'cached'|'cached_stale'|'offline'|'misconfigured'} ForecastStatus */

/**
 * @param {ForecastStatus|undefined|null} status
 */
export function forecastStatusLabel(status) {
  switch (status) {
    case 'connected':
      return '● Forecast live'
    case 'cached':
      return '● Forecast cached'
    case 'cached_stale':
      return '● Forecast cached (offline)'
    case 'offline':
      return '● Forecast offline'
    case 'no_coords':
      return '● Set location for forecast'
    case 'misconfigured':
      return '● Forecast misconfigured'
    case 'disabled':
    default:
      return '● Forecast off'
  }
}

/**
 * @param {ForecastStatus|undefined|null} status
 */
export function forecastStatusTone(status) {
  switch (status) {
    case 'connected':
    case 'cached':
      return 'text-gr33n-400'
    case 'cached_stale':
    case 'offline':
    case 'misconfigured':
      return 'text-amber-400'
    case 'no_coords':
      return 'text-amber-300/90'
    default:
      return 'text-zinc-500'
  }
}

/**
 * @param {object|null|undefined} onlineForecast
 */
export function formatForecastCurrent(onlineForecast) {
  const cur = onlineForecast?.current
  if (!cur) return null
  const parts = []
  if (cur.temperature_celsius != null) {
    parts.push(`${Math.round(cur.temperature_celsius)}°C`)
  }
  if (cur.cloud_cover_percent != null) {
    parts.push(`${Math.round(cur.cloud_cover_percent)}% clouds`)
  }
  return parts.length ? parts.join(' · ') : null
}

/**
 * @param {object|null|undefined} farm
 */
export function farmForecastOptedIn(farm) {
  const meta = farm?.meta_data
  if (!meta || typeof meta !== 'object') return false
  return meta.weather_forecast_enabled === true
}
