/**
 * Phase 71 — Feed & Water unification closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { buildLegacyRedirectRoutes } from '../lib/workspaces.js'
import router from '../router/index.js'

describe('Phase 71 — feed-water closure', () => {
  it('legacy feeding routes redirect into feed-water workspace tabs', () => {
    const paths = {
      '/feeding': 'daily',
      '/operations/feeding': 'programs',
      '/fertigation': 'advanced',
    }
    for (const [legacy, tab] of Object.entries(paths)) {
      const entry = buildLegacyRedirectRoutes().find((r) => r.path === legacy)
      expect(entry, legacy).toBeTruthy()
      const result = entry.redirect({
        path: legacy,
        query: {},
        hash: '',
        fullPath: legacy,
        matched: [],
        meta: {},
        name: undefined,
        params: {},
        redirectedFrom: undefined,
      })
      expect(result.path).toBe('/feed-water')
      expect(result.query.tab).toBe(tab)
    }
  })

  it('registers /feed-water and /hardware workspace routes', () => {
    expect(router.resolve('/feed-water').name).toBe('feed-water')
    expect(router.resolve('/hardware').name).toBe('hardware')
  })

  it('hardware workspace restored for GPIO board (Phase 70)', () => {
    const hw = readFileSync(join(process.cwd(), 'src/views/workspaces/HardwareWorkspace.vue'), 'utf8')
    expect(hw).toContain('GpioBoard')
    expect(readFileSync(join(process.cwd(), 'src/lib/workspaces.js'), 'utf8')).toContain("route: '/hardware'")
  })
})
