import { describe, it, expect } from 'vitest'
import router from '../router/index.js'
import { buildLegacyRedirectRoutes, redirectSunsetWorkspace } from '../lib/workspaces.js'

describe('Phase 68 / 78 WS4 — workspace legacy redirects', () => {
  const legacyPaths = buildLegacyRedirectRoutes().map((r) => r.path)

  it('registers a redirect for every absorbed legacy list route', () => {
    for (const path of legacyPaths) {
      const resolved = router.resolve(path)
      expect(resolved.matched.length).toBeGreaterThan(0)
      expect(resolved.name).not.toBe('login')
    }
  })

  it('redirect config sends /feeding with zone_id to zone water tab', () => {
    const entry = buildLegacyRedirectRoutes().find((r) => r.path === '/feeding')
    expect(entry).toBeTruthy()
    const result = entry.redirect({ path: '/feeding', query: { zone_id: '2' }, hash: '', fullPath: '/feeding?zone_id=2', matched: [], meta: {}, name: undefined, params: {}, redirectedFrom: undefined })
    expect(result.path).toBe('/zones/2')
    expect(result.query.tab).toBe('water')
    expect(result.query.zone_id).toBeUndefined()
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

  it('sunset redirects /feed-water and /hardware away from retired workspace routes', () => {
    expect(redirectSunsetWorkspace({ path: '/feed-water', query: { zone_id: '3' } }).path).toBe('/zones/3')
    expect(redirectSunsetWorkspace({ path: '/hardware', query: {} }).path).toBe('/zones')
    expect(router.resolve('/hardware').matched.length).toBeGreaterThan(0)
    expect(router.resolve('/feed-water').matched.length).toBeGreaterThan(0)
  })

  it('registers active workspace routes by name', () => {
    expect(router.resolve('/money').name).toBe('money')
    expect(router.resolve('/zones').name).toBe('zones')
  })
})
