import { describe, it, expect } from 'vitest'
import {
  forecastStatusLabel,
  forecastStatusTone,
  formatForecastCurrent,
  farmForecastOptedIn,
} from '../lib/siteWeatherForecast.js'

describe('siteWeatherForecast', () => {
  it('maps status to badge labels', () => {
    expect(forecastStatusLabel('connected')).toBe('● Forecast live')
    expect(forecastStatusLabel('cached_stale')).toBe('● Forecast cached (offline)')
    expect(forecastStatusLabel('disabled')).toBe('● Forecast off')
  })

  it('maps status to tone classes', () => {
    expect(forecastStatusTone('connected')).toContain('gr33n')
    expect(forecastStatusTone('offline')).toContain('amber')
  })

  it('formats current conditions', () => {
    const line = formatForecastCurrent({
      current: { temperature_celsius: 24.2, cloud_cover_percent: 35 },
    })
    expect(line).toBe('24°C · 35% clouds')
  })

  it('reads farm forecast opt-in from meta_data', () => {
    expect(farmForecastOptedIn({ meta_data: { weather_forecast_enabled: true } })).toBe(true)
    expect(farmForecastOptedIn({ meta_data: {} })).toBe(false)
  })
})
