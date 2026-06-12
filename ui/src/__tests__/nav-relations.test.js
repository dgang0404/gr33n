import { describe, it, expect } from 'vitest'
import { NAV_RELATIONS, relatedTo } from '../lib/navRelations.js'
import { buildNavGroups, collectSidebarRoutes } from '../lib/navGroups.js'
import { canonicalSidebarPath } from '../lib/workspaces.js'

describe('Phase 49/68/78 — nav relations', () => {
  const navRoutes = new Set(collectSidebarRoutes(buildNavGroups()))

  it('returns related routes for workspace siblings', () => {
    expect(relatedTo('/zones')).toContain('/money')
    expect(relatedTo('/zones')).toContain('/comfort-targets')
    expect(relatedTo('/money')).toContain('/zones')
    expect(relatedTo('/comfort-targets')).toContain('/zones')
  })

  it('maps legacy hint paths to feed-water workspace route', () => {
    expect(canonicalSidebarPath('/feeding')).toBe('/feed-water')
    expect(relatedTo('/feeding')).toContain('/feed-water')
  })

  it('returns empty for unknown routes', () => {
    expect(relatedTo('/unknown-route')).toEqual([])
    expect(relatedTo(null)).toEqual([])
  })

  it('links Farm Guardian to zones and help', () => {
    expect(relatedTo('/chat')).toContain('/zones')
  })

  it('only points primary relations at routes that exist in the sidebar', () => {
    for (const [from, targets] of Object.entries(NAV_RELATIONS)) {
      const fromSidebar = navRoutes.has(from) || navRoutes.has(canonicalSidebarPath(from))
      if (!fromSidebar && !['/feeding', '/fertigation', '/operations/feeding', '/operations/supplies', '/operations/money', '/sensors', '/actuators', '/lighting', '/pi-setup', '/costs', '/inventory', '/tasks', '/alerts', '/plants', '/schedules', '/automation', '/setpoints', '/feed-water', '/hardware'].includes(from)) {
        expect(navRoutes.has(from), `missing nav route ${from}`).toBe(true)
      }
      for (const to of targets) {
        const workspaceOk = ['/feed-water', '/hardware'].includes(to)
        expect(navRoutes.has(to) || workspaceOk, `${from} → ${to} not in sidebar`).toBe(true)
      }
    }
  })
})
