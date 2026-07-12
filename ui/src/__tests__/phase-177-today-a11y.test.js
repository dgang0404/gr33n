/**
 * Phase 177 — Today tab order smoke (DOM sequence).
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const uiSrc = join(process.cwd(), 'src')

describe('Phase 177 WS3 — Today focus order smoke', () => {
  it('Dashboard hero sections appear before action bar and details', () => {
    const dash = readFileSync(join(uiSrc, 'views/Dashboard.vue'), 'utf8')
    const markers = [
      'FarmTodayHeader',
      'FarmSiteStrip',
      'FarmTodayAttentionStrip',
      'FarmTodayZoneFilterBar',
      'today-farm-hero',
      'FarmTodayActionBar',
      'FarmTodayAskGr33n',
      'dashboard-details',
    ]
    let last = -1
    for (const marker of markers) {
      const idx = dash.indexOf(marker)
      expect(idx).toBeGreaterThan(-1)
      expect(idx).toBeGreaterThan(last)
      last = idx
    }
  })

  it('attention strip chips are keyboard-focusable buttons', () => {
    const strip = readFileSync(join(uiSrc, 'components/FarmTodayAttentionStrip.vue'), 'utf8')
    expect(strip).toContain('<button')
    expect(strip).toContain('aria-label')
    expect(strip).toContain('min-h-[44px]')
  })

  it('coach mark controls meet touch target size', () => {
    const coach = readFileSync(join(uiSrc, 'components/TodayCoachMarks.vue'), 'utf8')
    expect(coach).toContain('min-h-[44px]')
    expect(coach).toContain('aria-label="Dismiss tips"')
  })
})
