/**
 * Phase 166 — Today visual farm canvas closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')

describe('Phase 166 WS1 — farmVisualStatus', () => {
  it('ships zone status rollup lib', () => {
    const lib = readFileSync(join(repoRoot, 'ui/src/lib/farmVisualStatus.js'), 'utf8')
    expect(lib).toContain('computeZoneVisualStatus')
    expect(lib).toContain('classifySensorHardwareState')
    expect(lib).toContain('Not set up yet')
    expect(lib).toContain('gravity_drip')
  })
})

describe('Phase 166 WS2–WS4 — canvas components', () => {
  it('ships FarmCanvas, FarmCanvasZoneTile, FarmSiteStrip', () => {
    const canvas = readFileSync(join(repoRoot, 'ui/src/components/FarmCanvas.vue'), 'utf8')
    expect(canvas).toContain('saveZoneLayout')
    expect(canvas).toContain('arrangeMode')
    expect(canvas).toContain('uploadLayoutBackground')

    const tile = readFileSync(join(repoRoot, 'ui/src/components/FarmCanvasZoneTile.vue'), 'utf8')
    expect(tile).toContain('farm-tile-water')
    expect(tile).toContain('A zone is your grow area')

    const strip = readFileSync(join(repoRoot, 'ui/src/components/FarmSiteStrip.vue'), 'utf8')
    expect(strip).toContain('sunDialProgress')
    expect(strip).toContain('farm-site-water')
  })
})

describe('Phase 166 WS5 — Dashboard rewire', () => {
  it('imports canvas as hero and collapses detail sections', () => {
    const dash = readFileSync(join(repoRoot, 'ui/src/views/Dashboard.vue'), 'utf8')
    expect(dash).toContain('FarmCanvas')
    expect(dash).toContain('FarmSiteStrip')
    expect(dash).toContain('FarmTodayHeader')
    expect(dash).toContain('All the details')
    expect(dash).toContain('loadLayoutBackground')
    expect(dash).toContain('refreshReadings')
    expect(dash).not.toContain('FarmConfigCard')
  })
})

describe('Phase 166 WS6 — tests', () => {
  it('ships farm visual status and canvas tests', () => {
    expect(readFileSync(join(repoRoot, 'ui/src/__tests__/farm-visual-status.test.js'), 'utf8')).toContain('Humidity high')
    expect(readFileSync(join(repoRoot, 'ui/src/__tests__/farm-canvas.test.js'), 'utf8')).toContain('FarmCanvasZoneTile')
  })
})
