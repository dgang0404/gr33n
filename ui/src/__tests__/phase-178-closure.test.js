/**
 * Phase 178 — online weather forecast closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')
const repoDocs = join(repoRoot, 'docs')

describe('Phase 178 — online weather forecast closure', () => {
  it('plan documents Tier 3 scope', () => {
    const plan = readFileSync(join(repoDocs, 'plans/archive/phase_178_online_weather_forecast.plan.md'), 'utf8')
    expect(plan).toContain('WEATHER_PROVIDER')
    expect(plan).toContain('online_forecast')
  })

  it('backend wires provider, routes, and Open-Meteo client', () => {
    expect(existsSync(join(repoRoot, 'internal/weather/openmeteo.go'))).toBe(true)
    expect(existsSync(join(repoRoot, 'internal/weather/config.go'))).toBe(true)
    const routes = readFileSync(join(repoRoot, 'cmd/api/routes.go'), 'utf8')
    expect(routes).toContain('weather_forecast_available')
    expect(routes).toContain('PATCH /farms/{id}/weather/settings')
  })

  it('Today and Settings surface forecast status', () => {
    const strip = readFileSync(join(process.cwd(), 'src/components/FarmSiteStrip.vue'), 'utf8')
    const settings = readFileSync(join(process.cwd(), 'src/views/Settings.vue'), 'utf8')
    expect(strip).toContain('farm-site-forecast-status')
    expect(settings).toContain('settings-weather-forecast-toggle')
    expect(settings).toContain('WEATHER_PROVIDER=openmeteo')
  })

  it('farm context patches weather settings', () => {
    const farmContext = readFileSync(join(process.cwd(), 'src/stores/farmContext.js'), 'utf8')
    expect(farmContext).toContain('patchWeatherSettings')
  })

  it('migration adds api_openmeteo enum', () => {
    const mig = readFileSync(join(repoRoot, 'db/migrations/20260712_phase178_weather_openmeteo.sql'), 'utf8')
    expect(mig).toContain('api_openmeteo')
  })
})
