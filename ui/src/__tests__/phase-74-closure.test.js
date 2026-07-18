/**
 * Phase 74 WS6 / OC-74 — zone ops inbox closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { buildNavGroups, collectSidebarRoutes } from '../lib/navGroups.js'
import { buildLegacyRedirectRoutes, buildZoneOpsRedirectRoutes, WORKSPACES } from '../lib/workspaces.js'

const uiSrc = join(process.cwd(), 'src')
const repoDocs = join(process.cwd(), '..', 'docs')

describe('Phase 74 WS6 / OC-74 — zone ops inbox closure', () => {
  const groups = buildNavGroups()
  const routes = collectSidebarRoutes(groups)

  it('sidebar omits Tasks, Alerts, and Plants', () => {
    expect(routes).not.toContain('/tasks')
    expect(routes).not.toContain('/alerts')
    expect(routes).not.toContain('/plants')
    const today = groups.find((g) => g.label === 'Today')
    expect(today.items).toHaveLength(1)
    expect(today.items[0].to).toBe('/')
  })

  it('zones workspace includes Plants tab and absorbs /plants', () => {
    expect(WORKSPACES.zones.tabs.some((t) => t.id === 'plants')).toBe(true)
    expect(WORKSPACES.zones.absorbs['/plants']).toEqual({ tab: 'plants' })
  })

  it('/tasks redirects into zone Ops or Today', () => {
    const opsRoutes = buildZoneOpsRedirectRoutes()
    const tasks = opsRoutes.find((r) => r.path === '/tasks')
    expect(tasks?.redirect({ query: { zone_id: '3' } }).path).toBe('/zones/3')
    expect(tasks?.redirect({ query: { zone_id: '3' } }).query).toMatchObject({ tab: 'ops', ops: 'tasks' })
    expect(tasks?.redirect({ query: {} }).path).toBe('/')
    expect(opsRoutes.some((r) => r.path === '/alerts')).toBe(false)
  })

  it('/plants redirects to zones plants tab', () => {
    const entry = buildLegacyRedirectRoutes().find((r) => r.path === '/plants')
    expect(entry).toBeTruthy()
    const result = entry.redirect({ path: '/plants', query: {}, hash: '', fullPath: '/plants', matched: [], meta: {}, params: {} })
    expect(result.path).toBe('/zones')
    expect(result.query.tab).toBe('plants')
  })

  it('zone detail ships Ops and Plants tabs', () => {
    const zoneDetail = readFileSync(join(uiSrc, 'views/ZoneDetail.vue'), 'utf8')
    expect(zoneDetail).toContain('ZoneOpsSection')
    expect(zoneDetail).toContain('ZonePlantsSection')
    expect(zoneDetail).toMatch(/id: 'ops'/)
    expect(zoneDetail).toMatch(/id: 'plants'/)
  })

  it('operator-tour documents zone ops inbox (Phase 74)', () => {
    const tour = readFileSync(join(repoDocs, 'operator-tour.md'), 'utf8')
    expect(tour).toMatch(/7g\. Zone ops inbox/i)
  })
})
