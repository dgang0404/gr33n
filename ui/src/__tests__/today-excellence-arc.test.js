/**
 * Today excellence arc (166–177) — Dashboard wiring + component chain.
 * Canonical home for Dashboard.vue source assertions (Phase 202).
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')
const uiSrc = join(process.cwd(), 'src')

const ARC_COMPONENTS = [
  'FarmTodayHeader.vue',
  'FarmSiteStrip.vue',
  'FarmTodayAttentionStrip.vue',
  'FarmTodayZoneFilterBar.vue',
  'FarmCanvas.vue',
  'FarmZoneStack.vue',
  'FarmTodayActionBar.vue',
  'FarmTodayAskGr33n.vue',
  'TodayCoachMarks.vue',
]

const ARC_LIBS = [
  'farmTodayZoneFilter.js',
  'farmTodayHeader.js',
  'farmTodayAskGr33n.js',
  'farmTodayPulse.js',
  'farmTodayCoachMarks.js',
]

describe('Today excellence arc — component chain', () => {
  it('ships all arc surface components', () => {
    for (const name of ARC_COMPONENTS) {
      const src = readFileSync(join(uiSrc, 'components', name), 'utf8')
      expect(src.length).toBeGreaterThan(20)
    }
  })

  it('ships all arc libs', () => {
    for (const name of ARC_LIBS) {
      const src = readFileSync(join(uiSrc, 'lib', name), 'utf8')
      expect(src.length).toBeGreaterThan(20)
    }
  })

  it('Dashboard wires canvas hero, site strip, and detail collapse (166–167)', () => {
    const dash = readFileSync(join(uiSrc, 'views/Dashboard.vue'), 'utf8')
    expect(dash).toContain('FarmCanvas')
    expect(dash).toContain('FarmSiteStrip')
    expect(dash).toContain('FarmZoneStack')
    expect(dash).toContain('hidden md:block')
    expect(dash).toContain('All the details')
    expect(dash).toContain('loadLayoutBackground')
    expect(dash).toContain('refreshReadings')
    expect(dash).not.toContain('FarmConfigCard')
    expect(dash).not.toContain('GettingStartedChecklist')
    expect(dash).not.toContain('showFirstRunChecklist')
    expect(dash).not.toContain('firstRunDismissed')
    expect(existsSync(join(uiSrc, 'components/GettingStartedChecklist.vue'))).toBe(false)
  })

  it('Dashboard wires attention, filter bar, and large-farm navigation (169, 173)', () => {
    const dash = readFileSync(join(uiSrc, 'views/Dashboard.vue'), 'utf8')
    expect(dash).toContain('FarmTodayAttentionStrip')
    expect(dash).toContain('buildTodayAttentionStarters')
    expect(dash).toContain('FarmTodayZoneFilterBar')
    expect(dash).toContain('filterZonesForToday')
    expect(dash).toContain('filteredZones')
    expect(dash).toContain('todayZoneFilter')
  })

  it('Dashboard wires Today header rhythm and document title (174)', () => {
    const dash = readFileSync(join(uiSrc, 'views/Dashboard.vue'), 'utf8')
    expect(dash).toContain('FarmTodayHeader')
    expect(dash).toContain('filter-attention')
    expect(dash).not.toContain('dashboard-attention-row')
    expect(dash).toContain('document.title')
  })

  it('Dashboard demotes Guardian to single ask row + details subsection (175)', () => {
    const dash = readFileSync(join(uiSrc, 'views/Dashboard.vue'), 'utf8')
    expect(dash).toContain('FarmTodayActionBar')
    expect(dash).toContain('comfortSchedulesLink')
    expect(dash).toContain('FarmTodayAskGr33n')
    expect(dash).toContain('curatedTodayAskStarters')
    expect(dash).toContain('buildMorningWalkthroughStarters')
    expect(dash).not.toContain('dashboard-attention-starters')
    expect(dash).not.toContain('dashboard-morning-check-starters')
    expect(dash).not.toContain('dashboard-weather-starters')
    expect(dash).toContain('dashboard-details-guardian')
    expect(dash).toContain('detailsGuardianStarters')
    expect(dash).toContain('mergeTodayDetailsGuardianStarters')
    expect(dash).toContain('showEmptyFarmStarters')
    expect(dash).toContain('emptyFarmStarters')
    expect(dash).toContain('buildSetupStarters')
    expect(dash).toContain('dashboard-empty-farm-starters')
    expect(dash).toContain('dashboard-setup-starters')
    const chipRows = (dash.match(/<GuardianStarterChips/g) || []).length
    expect(chipRows).toBeLessThanOrEqual(3)
  })

  it('Dashboard passes pulse data into site strip without extra row (176)', () => {
    const dash = readFileSync(join(uiSrc, 'views/Dashboard.vue'), 'utf8')
    expect(dash).toContain(':crop-cycles="cropCycles"')
    expect(dash).toContain(':devices="store.devices"')
    expect(dash).toContain(':queue-depth="queueDepth"')
    expect(dash).not.toContain('FarmTodayPulse')
  })

  it('Dashboard wires coach marks and non-blocking refresh (177)', () => {
    const dash = readFileSync(join(uiSrc, 'views/Dashboard.vue'), 'utf8')
    expect(dash).toContain('TodayCoachMarks')
    expect(dash).toContain('today-farm-hero')
    expect(dash).toContain(':has-attention="hasAttentionZones"')
    expect(dash).toContain('zonesCached')
    expect(dash).toContain('void store.loadLayoutBackground(fid)')
    expect(dash).toContain('void fetchSiteWeather(fid)')
  })
})

describe('Today excellence arc — closure tests', () => {
  it('links phase 173 through 177 closure bundles', () => {
    for (const n of [173, 174, 175, 176, 177]) {
      const test = readFileSync(join(uiSrc, '__tests__', `phase-${n}-closure.test.js`), 'utf8')
      expect(test).toContain(`Phase ${n}`)
    }
    const roadmap = readFileSync(join(repoRoot, 'docs/plans/phase_173_177_today_excellence_roadmap.plan.md'), 'utf8')
    expect(roadmap).toContain('173')
    expect(roadmap).toContain('177')
  })
})
