/**
 * Phase 169 — Today attention cockpit closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')
const uiSrc = join(process.cwd(), 'src')

describe('Phase 169 WS1–WS3 — attention cockpit', () => {
  it('ships FarmTodayAttentionStrip', () => {
    const strip = readFileSync(join(uiSrc, 'components/FarmTodayAttentionStrip.vue'), 'utf8')
    expect(strip).toContain('farm-today-attention')
    expect(strip).toContain('listAttentionZones')
  })

  it('sorts FarmCanvas zones attention-first', () => {
    const canvas = readFileSync(join(uiSrc, 'components/FarmCanvas.vue'), 'utf8')
    expect(canvas).toContain('sortZonesForStack')
    expect(canvas).toContain('zoneHasTasksDueToday')
  })

  it('exports attention helpers and starters', () => {
    const lib = readFileSync(join(uiSrc, 'lib/zoneQuickActions.js'), 'utf8')
    expect(lib).toContain('zoneNeedsAttention')
    expect(lib).toContain('listAttentionZones')

    const starters = readFileSync(join(uiSrc, 'lib/guardianStarters.js'), 'utf8')
    expect(starters).toContain('buildTodayAttentionStarters')
    expect(starters).toContain('dashboard_attention')
  })
})

describe('Phase 169 WS4 — docs', () => {
  it('documents attention cockpit in current-state', () => {
    const state = readFileSync(join(repoRoot, 'docs/current-state.md'), 'utf8')
    expect(state).toContain('Phase 169')
  })
})
