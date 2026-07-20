import { describe, it, expect } from 'vitest'
import {
  WORKSPACES,
  workspaceFor,
  tabsFor,
  buildLegacyRedirectRoutes,
  canonicalSidebarPath,
  resolveWorkspaceTab,
} from '../lib/workspaces.js'

describe('Phase 68 / 78 WS1 — workspaces model', () => {
  it('maps every absorbed legacy path to a workspace tab', () => {
    for (const ws of Object.values(WORKSPACES)) {
      for (const [legacyPath, target] of Object.entries(ws.absorbs ?? {})) {
        const hit = workspaceFor(legacyPath)
        expect(hit).toBeTruthy()
        expect(hit.route).toBe(ws.route)
        if (target.zoneTab) {
          expect(hit.zoneTab).toBe(target.zoneTab)
        } else {
          expect(tabsFor(hit.workspaceId).some((t) => t.id === hit.tab)).toBe(true)
        }
      }
    }
  })

  it('workspaceFor fleet and zone-scoped legacy paths', () => {
    expect(workspaceFor('/sensors')).toMatchObject({ route: '/zones', tab: 'fleet', fleet: 'sensors' })
    expect(workspaceFor('/actuators')).toMatchObject({ route: '/zones', tab: 'fleet', fleet: 'controls' })
    expect(workspaceFor('/feeding')).toMatchObject({ route: '/feed-water', tab: 'daily' })
    expect(workspaceFor('/fertigation')).toMatchObject({ route: '/feed-water', tab: 'advanced' })
    expect(workspaceFor('/costs')).toMatchObject({ route: '/money', tab: 'ledger' })
    expect(workspaceFor('/pi-setup')).toMatchObject({ route: '/hardware', tab: 'reference' })
  })

  it('canonicalSidebarPath maps legacy routes to workspace sidebar entries', () => {
    expect(canonicalSidebarPath('/feeding')).toBe('/feed-water')
    expect(canonicalSidebarPath('/operations/money')).toBe('/money')
    expect(canonicalSidebarPath('/sensors')).toBe('/zones')
    expect(canonicalSidebarPath('/comfort-targets')).toBe('/comfort-targets')
  })

  it('resolveWorkspaceTab maps legacy strains tab to plants', () => {
    expect(resolveWorkspaceTab('zones', 'strains')).toBe('plants')
    expect(resolveWorkspaceTab('zones', 'plants')).toBe('plants')
  })

  it('resolveWorkspaceTab falls back to first tab for unknown ids', () => {
    expect(resolveWorkspaceTab('zones', 'bogus')).toBe('rooms')
    expect(resolveWorkspaceTab('zones', undefined)).toBe('rooms')
  })

  it('buildLegacyRedirectRoutes covers all absorbs', () => {
    const redirects = buildLegacyRedirectRoutes()
    const paths = redirects.map((r) => r.path)
    expect(paths).toContain('/feeding')
    expect(paths).toContain('/fertigation')
    expect(paths).toContain('/operations/supplies')
    expect(paths.length).toBeGreaterThanOrEqual(10)
  })

  it('legacy /fertigation redirect preserves sub-tab as fert_tab', () => {
    const redirects = buildLegacyRedirectRoutes()
    const fert = redirects.find((r) => r.path === '/fertigation')
    expect(fert).toBeTruthy()
    const dest = fert.redirect({
      path: '/fertigation',
      query: { tab: 'programs', recipe: '10' },
    })
    expect(dest).toEqual({
      path: '/feed-water',
      query: { tab: 'advanced', fert_tab: 'programs', recipe: '10' },
    })
  })
})
