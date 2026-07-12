/**
 * Phase 176 — Today farm pulse closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')
const uiSrc = join(process.cwd(), 'src')

describe('Phase 176 WS1 — farmTodayPulse lib', () => {
  it('ships pulse builders', () => {
    const lib = readFileSync(join(uiSrc, 'lib/farmTodayPulse.js'), 'utf8')
    expect(lib).toContain('buildFarmTodayPulse')
    expect(lib).toContain('resolveNextWaterCell')
    expect(lib).toContain('resolveGrowingCell')
    expect(lib).toContain('resolveDevicesCell')
  })
})

describe('Phase 176 WS2 — FarmSiteStrip pulse cells', () => {
  it('renders pulse cells inside site strip', () => {
    const strip = readFileSync(join(uiSrc, 'components/FarmSiteStrip.vue'), 'utf8')
    expect(strip).toContain('buildFarmTodayPulse')
    expect(strip).toContain('farm-site-pulse-')
    expect(strip).toContain('pulseCells')
  })
})

describe('Phase 176 WS3 — Dashboard wiring', () => {
  it('passes pulse data into FarmSiteStrip without a new row component', () => {
    const dash = readFileSync(join(uiSrc, 'views/Dashboard.vue'), 'utf8')
    expect(dash).toContain('FarmSiteStrip')
    expect(dash).toContain(':crop-cycles="cropCycles"')
    expect(dash).toContain(':devices="store.devices"')
    expect(dash).toContain(':queue-depth="queueDepth"')
    expect(dash).not.toContain('FarmTodayPulse')
  })
})

describe('Phase 176 WS5 — docs', () => {
  it('documents phase 176 in current-state', () => {
    const state = readFileSync(join(repoRoot, 'docs/current-state.md'), 'utf8')
    expect(state).toContain('176')
  })
})
