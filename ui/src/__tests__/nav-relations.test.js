import { describe, it, expect } from 'vitest'
import { NAV_RELATIONS, relatedTo } from '../lib/navRelations.js'
import { buildNavGroups, collectSidebarRoutes } from '../lib/navGroups.js'
import { canonicalSidebarPath } from '../lib/workspaces.js'

describe('Phase 49/68 — nav relations', () => {
  const navRoutes = new Set(collectSidebarRoutes(buildNavGroups('/farms/1/crop-cycles/compare')))

  it('returns related routes for workspace siblings', () => {
    expect(relatedTo('/zones')).toContain('/feed-water')
    expect(relatedTo('/zones')).toContain('/hardware')
    expect(relatedTo('/feed-water')).toContain('/zones')
    expect(relatedTo('/money')).toContain('/feed-water')
  })

  it('maps legacy hint paths to sidebar workspace routes', () => {
    expect(canonicalSidebarPath('/feeding')).toBe('/feed-water')
    expect(relatedTo('/feeding')).toContain('/zones')
  })

  it('returns empty for unknown routes', () => {
    expect(relatedTo('/chat')).toEqual([])
    expect(relatedTo(null)).toEqual([])
  })

  it('only points primary relations at routes that exist in the sidebar', () => {
    for (const [from, targets] of Object.entries(NAV_RELATIONS)) {
      const fromSidebar = navRoutes.has(from) || navRoutes.has(canonicalSidebarPath(from))
      if (!fromSidebar && !['/feeding', '/fertigation', '/operations/feeding', '/operations/supplies', '/operations/money', '/sensors', '/actuators', '/lighting', '/pi-setup', '/costs', '/inventory', '/tasks', '/alerts', '/plants'].includes(from)) {
        expect(navRoutes.has(from), `missing nav route ${from}`).toBe(true)
      }
      for (const to of targets) {
        expect(navRoutes.has(to), `${from} → ${to} not in sidebar`).toBe(true)
      }
    }
  })
})
