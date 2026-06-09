/**
 * Phase 69 WS6 / OC-69 — zone workspace hub closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { buildLegacyRedirectRoutes } from '../lib/workspaces.js'

const uiSrc = join(process.cwd(), 'src')
const repoDocs = join(process.cwd(), '..', 'docs')

describe('Phase 69 WS6 / OC-69 — zone workspace hub closure', () => {
  it('legacy sensor/control/lighting routes redirect to zones fleet', () => {
    for (const path of ['/sensors', '/actuators', '/lighting']) {
      const entry = buildLegacyRedirectRoutes().find((r) => r.path === path)
      expect(entry).toBeTruthy()
      const result = entry.redirect({ path, query: {}, hash: '', fullPath: path, matched: [], meta: {}, params: {} })
      expect(result.path).toBe('/zones')
      expect(result.query.tab).toBe('fleet')
    }
  })

  it('zone inline components and fleet grouping ship', () => {
    expect(readFileSync(join(uiSrc, 'components/ZoneLightingEditor.vue'), 'utf8')).toContain('zone-lighting-editor')
    expect(readFileSync(join(uiSrc, 'lib/fleetGrouping.js'), 'utf8')).toContain('groupEntitiesByZone')
  })

  it('operator-tour documents zone inline edit and fleet', () => {
    const tour = readFileSync(join(repoDocs, 'operator-tour.md'), 'utf8')
    expect(tour).toMatch(/Phase 69|zone.*inline|Fleet/i)
  })
})
