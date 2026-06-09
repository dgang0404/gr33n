import { describe, it, expect } from 'vitest'
import router from '../router/index.js'
import { buildLegacyRedirectRoutes } from '../lib/workspaces.js'

describe('Phase 68 WS4 — workspace legacy redirects', () => {
  const legacyPaths = buildLegacyRedirectRoutes().map((r) => r.path)

  it('registers a redirect for every absorbed legacy list route', () => {
    for (const path of legacyPaths) {
      const resolved = router.resolve(path)
      expect(resolved.matched.length).toBeGreaterThan(0)
      expect(resolved.name).not.toBe('login')
    }
  })

  it('redirect config sends /feeding to feed-water daily tab', () => {
    const entry = buildLegacyRedirectRoutes().find((r) => r.path === '/feeding')
    expect(entry).toBeTruthy()
    const result = entry.redirect({ path: '/feeding', query: { zone_id: '2' }, hash: '', fullPath: '/feeding?zone_id=2', matched: [], meta: {}, name: undefined, params: {}, redirectedFrom: undefined })
    expect(result.path).toBe('/feed-water')
    expect(result.query.tab).toBe('daily')
    expect(result.query.zone_id).toBe('2')
  })

  it('redirect config sends /sensors to zones fleet sensors', () => {
    const entry = buildLegacyRedirectRoutes().find((r) => r.path === '/sensors')
    const result = entry.redirect({ path: '/sensors', query: {}, hash: '', fullPath: '/sensors', matched: [], meta: {}, name: undefined, params: {}, redirectedFrom: undefined })
    expect(result.path).toBe('/zones')
    expect(result.query.tab).toBe('fleet')
    expect(result.query.fleet).toBe('sensors')
  })

  it('keeps detail routes without redirect', () => {
    expect(router.resolve('/sensors/3').name).toBe('sensor-detail')
    expect(router.resolve('/zones/2').name).toBe('zone-detail')
  })

  it('redirect config sends /schedules to comfort workspace schedules tab', () => {
    const entry = buildLegacyRedirectRoutes().find((r) => r.path === '/schedules')
    expect(entry).toBeTruthy()
    const result = entry.redirect({ path: '/schedules', query: {}, hash: '', fullPath: '/schedules', matched: [], meta: {}, name: undefined, params: {}, redirectedFrom: undefined })
    expect(result.path).toBe('/comfort-targets')
    expect(result.query.tab).toBe('schedules')
  })

  it('registers workspace routes by name', () => {
    expect(router.resolve('/hardware').name).toBe('hardware')
    expect(router.resolve('/feed-water').name).toBe('feed-water')
    expect(router.resolve('/money').name).toBe('money')
    expect(router.resolve('/zones').name).toBe('zones')
  })
})
