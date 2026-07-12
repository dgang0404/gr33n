/**
 * Phase 175 — Today farm-first actions closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')
const uiSrc = join(process.cwd(), 'src')

describe('Phase 175 WS1 — FarmTodayActionBar', () => {
  it('ships action bar with farm workspace links', () => {
    const bar = readFileSync(join(uiSrc, 'components/FarmTodayActionBar.vue'), 'utf8')
    expect(bar).toContain('farm-today-action-bar')
    expect(bar).toContain('Feed &amp; water')
    expect(bar).toContain('What runs when')
    expect(bar).toContain('My zones')
  })

  it('Dashboard wires action bar below canvas', () => {
    const dash = readFileSync(join(uiSrc, 'views/Dashboard.vue'), 'utf8')
    expect(dash).toContain('FarmTodayActionBar')
    expect(dash).toContain('comfortSchedulesLink')
  })
})

describe('Phase 175 WS2 — Guardian demotion', () => {
  it('uses single FarmTodayAskGr33n instead of four chip rows', () => {
    const dash = readFileSync(join(uiSrc, 'views/Dashboard.vue'), 'utf8')
    expect(dash).toContain('FarmTodayAskGr33n')
    expect(dash).toContain('curatedTodayAskStarters')
    expect(dash).not.toContain('dashboard-attention-starters')
    expect(dash).not.toContain('dashboard-morning-check-starters')
    expect(dash).not.toContain('dashboard-weather-starters')
    const chipRows = (dash.match(/<GuardianStarterChips/g) || []).length
    expect(chipRows).toBeLessThanOrEqual(3)
  })

  it('moves full starter set into details subsection', () => {
    const dash = readFileSync(join(uiSrc, 'views/Dashboard.vue'), 'utf8')
    expect(dash).toContain('dashboard-details-guardian')
    expect(dash).toContain('detailsGuardianStarters')
    expect(dash).toContain('mergeTodayDetailsGuardianStarters')
  })

  it('ships curated ask lib', () => {
    const lib = readFileSync(join(uiSrc, 'lib/farmTodayAskGr33n.js'), 'utf8')
    expect(lib).toContain('buildCuratedTodayAskStarters')
    expect(lib).toContain('shouldOfferMorningCheckOnToday')
  })
})

describe('Phase 175 WS3 — zone quick actions own zone Guardian', () => {
  it('ZoneQuickActions still exposes Guardian starters', () => {
    const sheet = readFileSync(join(uiSrc, 'components/ZoneQuickActions.vue'), 'utf8')
    expect(sheet).toContain('zone-quick-guardian')
    expect(sheet).toContain('buildZoneQuickStarters')
  })
})

describe('Phase 175 WS4 — empty farm contract', () => {
  it('keeps setup Guardian chips for empty farm', () => {
    const dash = readFileSync(join(uiSrc, 'views/Dashboard.vue'), 'utf8')
    expect(dash).toContain('dashboard-empty-farm-starters')
    expect(dash).toContain('dashboard-setup-starters')
  })
})

describe('Phase 175 WS5 — docs', () => {
  it('documents phase 175 in current-state', () => {
    const state = readFileSync(join(repoRoot, 'docs/current-state.md'), 'utf8')
    expect(state).toContain('175')
  })
})
