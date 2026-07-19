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
 * @param {object|null|undefined} farm
 * @returns {'celsius'|'fahrenheit'}
 */
export function farmTemperatureUnit(farm) {
  const unit = farm?.meta_data?.temperature_unit
  return unit === 'fahrenheit' ? 'fahrenheit' : 'celsius'
}

/**
 * @param {number|null|undefined} celsius
 * @param {'celsius'|'fahrenheit'} [unit]
 */
export function formatTemperature(celsius, unit = 'celsius') {
  if (celsius == null || !Number.isFinite(Number(celsius))) return null
  const c = Number(celsius)
  if (unit === 'fahrenheit') {
    return `${Math.round(c * 9 / 5 + 32)}°F`
  }
  return `${Math.round(c)}°C`
}

/**
 * @param {object|null|undefined} onlineForecast
 * @param {'celsius'|'fahrenheit'} [unit]
 */
export function formatForecastCurrent(onlineForecast, unit = 'celsius') {
  const cur = onlineForecast?.current
  if (!cur) return null
  const parts = []
  const temp = formatTemperature(cur.temperature_celsius, unit)
  if (temp) parts.push(temp)
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
