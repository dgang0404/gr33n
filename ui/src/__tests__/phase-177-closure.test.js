/**
 * Phase 177 — Today first impression closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')
const uiSrc = join(process.cwd(), 'src')

describe('Phase 177 WS1 — demo showcase seed', () => {
  it('adds propagation light for demo tile story', () => {
    const seed = readFileSync(join(repoRoot, 'db/seeds/master_seed.sql'), 'utf8')
    expect(seed).toContain('seed_phase177')
    expect(seed).toContain('Propagation T5 Rack')
    expect(seed).toContain('Propagation 24h Photoperiod')
  })
})

describe('Phase 177 WS2 — TodayCoachMarks', () => {
  it('ships coach marks lib and component', () => {
    const lib = readFileSync(join(uiSrc, 'lib/farmTodayCoachMarks.js'), 'utf8')
    expect(lib).toContain('gr33n_today_coach_done')
    expect(lib).toContain('buildTodayCoachSteps')
    const coach = readFileSync(join(uiSrc, 'components/TodayCoachMarks.vue'), 'utf8')
    expect(coach).toContain('today-coach-marks')
    expect(coach).toContain('prefers-reduced-motion')
    expect(coach).not.toContain('Guardian')
  })

  it('Dashboard wires coach marks for populated farms', () => {
    const dash = readFileSync(join(uiSrc, 'views/Dashboard.vue'), 'utf8')
    expect(dash).toContain('TodayCoachMarks')
    expect(dash).toContain('today-farm-hero')
    expect(dash).toContain(':has-attention="hasAttentionZones"')
  })
})

describe('Phase 177 WS3 — perf and a11y', () => {
  it('refreshAll does not block on layout background or weather', () => {
    const dash = readFileSync(join(uiSrc, 'views/Dashboard.vue'), 'utf8')
    expect(dash).toContain('zonesCached')
    expect(dash).toContain('void store.loadLayoutBackground(fid)')
    expect(dash).toContain('void fetchSiteWeather(fid)')
  })

  it('attention strip announces count changes', () => {
    const strip = readFileSync(join(uiSrc, 'components/FarmTodayAttentionStrip.vue'), 'utf8')
    expect(strip).toContain('aria-live="polite"')
    expect(strip).toContain('sr-only')
  })
})

describe('Phase 177 WS4 — docs', () => {
  it('documents phase 177 and arc closure', () => {
    const state = readFileSync(join(repoRoot, 'docs/current-state.md'), 'utf8')
    expect(state).toContain('177')
    const tour = readFileSync(join(repoRoot, 'docs/operator-tour.md'), 'utf8')
    expect(tour).toContain('7l')
    expect(tour).toContain('173')
  })
})
