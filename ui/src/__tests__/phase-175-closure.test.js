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
})

describe('Phase 175 WS2 — Guardian demotion', () => {
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
  it('setup starters lib covers empty farm path', () => {
    const starters = readFileSync(join(uiSrc, 'lib/guardianStarters.js'), 'utf8')
    expect(starters).toContain('buildSetupStarters')
  })
})

describe('Phase 175 WS5 — docs', () => {
  it('documents phase 175 in current-state', () => {
    const state = readFileSync(join(repoRoot, 'docs/current-state.md'), 'utf8')
    expect(state).toContain('175')
  })
})
