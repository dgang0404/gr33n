/**
 * Phase 174 — Today visual hierarchy closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')
const uiSrc = join(process.cwd(), 'src')

describe('Phase 174 WS1 — Today naming', () => {
  it('TopBar labels / as Today', () => {
    const topBar = readFileSync(join(uiSrc, 'components/TopBar.vue'), 'utf8')
    expect(topBar).toContain("'/'")
    expect(topBar).toMatch(/'\/':\s*'Today'/)
  })

  it('guardianRouteRef and starters use Today for dashboard surface', () => {
    const routeRef = readFileSync(join(uiSrc, 'lib/guardianRouteRef.js'), 'utf8')
    expect(routeRef).toMatch(/'\/':\s*'Today'/)

    const starters = readFileSync(join(uiSrc, 'lib/guardianStarters.js'), 'utf8')
    expect(starters).toContain("const pageName = onChat ? 'Farm Guardian chat' : 'Today'")
    expect(starters).toContain("operationsRouteRef('/', 'Today', 'dashboard_ops')")
  })
})

describe('Phase 174 WS2 — FarmTodayHeader', () => {
  it('Dashboard imports header and drops duplicate attention row', () => {
    const dash = readFileSync(join(uiSrc, 'views/Dashboard.vue'), 'utf8')
    expect(dash).toContain('FarmTodayHeader')
    expect(dash).toContain('filter-attention')
    expect(dash).not.toContain('dashboard-attention-row')
    expect(dash).toContain('document.title')
  })

  it('ships farmTodayHeader rollup lib', () => {
    const lib = readFileSync(join(uiSrc, 'lib/farmTodayHeader.js'), 'utf8')
    expect(lib).toContain('buildFarmTodayRollup')
    expect(lib).toContain('todayTimeGreeting')
  })
})

describe('Phase 174 WS3 — section rhythm', () => {
  it('FarmCanvas has taller min-height and richer background', () => {
    const canvas = readFileSync(join(uiSrc, 'components/FarmCanvas.vue'), 'utf8')
    expect(canvas).toContain('min-height: 420px')
    expect(canvas).toContain('min-height: 480px')
    expect(canvas).toContain('opacity-45')
  })

  it('FarmZoneStack uses Your farm heading', () => {
    const stack = readFileSync(join(uiSrc, 'components/FarmZoneStack.vue'), 'utf8')
    expect(stack).toContain('Your farm')
    expect(stack).not.toContain('Your zones')
  })
})

describe('Phase 174 WS4 — tile polish', () => {
  it('FarmCanvasZoneTile adds hover glow and empty-zone styling', () => {
    const tile = readFileSync(join(uiSrc, 'components/FarmCanvasZoneTile.vue'), 'utf8')
    expect(tile).toContain('hover:shadow-xl')
    expect(tile).toContain('Ready to plant')
    expect(tile).toContain(':title="plantsLine"')
    expect(tile).toContain('shadow-amber-900')
  })
})

describe('Phase 174 WS5 — docs', () => {
  it('documents phase 174 in current-state', () => {
    const state = readFileSync(join(repoRoot, 'docs/current-state.md'), 'utf8')
    expect(state).toContain('174')
  })
})
