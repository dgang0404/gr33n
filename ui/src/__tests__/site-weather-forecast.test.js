import { describe, it, expect } from 'vitest'
import {
  forecastStatusLabel,
  forecastStatusTone,
  formatForecastCurrent,
  formatTemperature,
  farmForecastOptedIn,
  farmTemperatureUnit,
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
    const wx = { current: { temperature_celsius: 24.2, cloud_cover_percent: 35 } }
    expect(formatForecastCurrent(wx)).toBe('24°C · 35% clouds')
    expect(formatForecastCurrent(wx, 'fahrenheit')).toBe('76°F · 35% clouds')
  })

  it('formats standalone temperature', () => {
    expect(formatTemperature(0, 'celsius')).toBe('0°C')
    expect(formatTemperature(0, 'fahrenheit')).toBe('32°F')
    expect(formatTemperature(22.9, 'fahrenheit')).toBe('73°F')
  })

  it('reads farm temperature unit from meta_data', () => {
    expect(farmTemperatureUnit({ meta_data: { temperature_unit: 'fahrenheit' } })).toBe('fahrenheit')
    expect(farmTemperatureUnit({ meta_data: {} })).toBe('celsius')
  })

  it('reads farm forecast opt-in from meta_data', () => {
    expect(farmForecastOptedIn({ meta_data: { weather_forecast_enabled: true } })).toBe(true)
    expect(farmForecastOptedIn({ meta_data: {} })).toBe(false)
  })
})
