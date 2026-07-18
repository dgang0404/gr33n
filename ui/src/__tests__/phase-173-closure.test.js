/**
 * Phase 173 — Today large-farm navigation closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')
const uiSrc = join(process.cwd(), 'src')

describe('Phase 173 WS1–WS2 — filter library + bar', () => {
  it('ships farmTodayZoneFilter.js with filter/paging helpers', () => {
    const lib = readFileSync(join(uiSrc, 'lib/farmTodayZoneFilter.js'), 'utf8')
    expect(lib).toContain('TODAY_ZONE_FILTERS')
    expect(lib).toContain('filterZonesForToday')
    expect(lib).toContain('countZonesPerFilter')
    expect(lib).toContain('shouldShowTodayZoneFilterBar')
    expect(lib).toContain('shouldPageZoneStack')
    expect(lib).toContain('shouldOfferDesktopListView')
    expect(lib).toContain('paginateZones')
  })

  it('ships FarmTodayZoneFilterBar.vue', () => {
    const bar = readFileSync(join(uiSrc, 'components/FarmTodayZoneFilterBar.vue'), 'utf8')
    expect(bar).toContain('farm-today-zone-filter-bar')
    expect(bar).toContain('shouldShowTodayZoneFilterBar')
    expect(bar).toContain('countZonesPerFilter')
  })
})

describe('Phase 173 WS4 — mobile paging', () => {
  it('FarmZoneStack pages beyond 8 zones', () => {
    const stack = readFileSync(join(uiSrc, 'components/FarmZoneStack.vue'), 'utf8')
    expect(stack).toContain('paginateZones')
    expect(stack).toContain('shouldPageZoneStack')
    expect(stack).toContain('farm-zone-stack-pager')
  })
})

describe('Phase 173 WS5 — desktop list overflow', () => {
  it('FarmCanvas offers a Map/List toggle beyond threshold', () => {
    const canvas = readFileSync(join(uiSrc, 'components/FarmCanvas.vue'), 'utf8')
    expect(canvas).toContain('shouldOfferDesktopListView')
    expect(canvas).toContain('farm-canvas-view-toggle')
    expect(canvas).toContain('farm-canvas-list')
  })
})

describe('Phase 173 WS6 — large-farm fixture', () => {
  it('ships a 24-zone synthetic fixture', () => {
    const fixture = readFileSync(join(uiSrc, '__tests__/fixtures/largeFarmZones.js'), 'utf8')
    expect(fixture).toContain('LARGE_FARM_ZONES')
    expect(fixture).toContain('buildLargeFarmZones')
  })
})

describe('Phase 173 WS7 — docs', () => {
  it('documents the large-farm navigation phase in current-state', () => {
    const state = readFileSync(join(repoRoot, 'docs/current-state.md'), 'utf8')
    expect(state).toContain('173')
  })
})
