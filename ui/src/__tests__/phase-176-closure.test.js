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

describe('Phase 176 WS5 — docs', () => {
  it('documents phase 176 in current-state', () => {
    const state = readFileSync(join(repoRoot, 'docs/current-state.md'), 'utf8')
    expect(state).toContain('176')
  })
})
