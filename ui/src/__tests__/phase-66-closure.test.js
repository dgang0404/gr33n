/**
 * Phase 66 WS6 / OC-66 — weather & site context closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'
import { buildWeatherStarters } from '../lib/guardianStarters.js'
import { daylightChipFromSiteWeather } from '../lib/siteWeather.js'

const repoDocs = join(process.cwd(), '..', 'docs')
const repoRoot = join(process.cwd(), '..')

describe('Phase 66 WS6 / OC-66 — weather & site closure', () => {
  it('documents offline solar and plan shipped', () => {
    const arch = readFileSync(join(repoDocs, 'farm-guardian-architecture.md'), 'utf8')
    const plan = readFileSync(join(repoDocs, 'plans/phase_66_weather_site_context.plan.md'), 'utf8')
    const tour = readFileSync(join(repoDocs, 'operator-tour.md'), 'utf8')
    expect(existsSync(join(repoRoot, 'internal/solar/solar.go'))).toBe(true)
    expect(arch).toContain('site_weather')
    expect(plan).toContain('**Shipped.**')
    expect(tour).toContain('Farm site')
  })

  it('supplemental-light starter uses site_weather', () => {
    const starters = buildWeatherStarters({ surface: 'dashboard', farmName: 'Demo' })
    expect(starters[0].label).toMatch(/supplemental light/i)
    expect(starters[0].message).toContain('site_weather')
  })

  it('daylight chip from site weather response', () => {
    const chip = daylightChipFromSiteWeather({ solar: { daylength_hours: 14.2 } })
    expect(chip?.value).toBe('14.2 h')
  })

  it('Settings and API wire site coords and weather routes', () => {
    const settings = readFileSync(join(process.cwd(), 'src/views/Settings.vue'), 'utf8')
    const routes = readFileSync(join(repoRoot, 'cmd/api/routes.go'), 'utf8')
    expect(settings).toContain('settings-farm-site')
    const farmContext = readFileSync(join(process.cwd(), 'src/stores/farmContext.js'), 'utf8')
    expect(farmContext).toContain('patchSite')
    expect(routes).toContain('GET /farms/{id}/site-weather')
    expect(routes).toContain('PATCH /farms/{id}/site')
  })
})
