/**
 * Today excellence arc (173–177) — import chain + Guardian demotion contract.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
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

  it('Dashboard wires the full hero flow without Guardian chip wall', () => {
    const dash = readFileSync(join(uiSrc, 'views/Dashboard.vue'), 'utf8')
    expect(dash).toContain('FarmTodayHeader')
    expect(dash).toContain('FarmSiteStrip')
    expect(dash).toContain('FarmTodayAttentionStrip')
    expect(dash).toContain('FarmTodayZoneFilterBar')
    expect(dash).toContain('FarmTodayActionBar')
    expect(dash).toContain('FarmTodayAskGr33n')
    expect(dash).toContain('TodayCoachMarks')
    expect(dash).not.toContain('dashboard-attention-starters')
    expect(dash).not.toContain('dashboard-morning-check-starters')
    const chipRows = (dash.match(/<GuardianStarterChips/g) || []).length
    expect(chipRows).toBeLessThanOrEqual(3)
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
