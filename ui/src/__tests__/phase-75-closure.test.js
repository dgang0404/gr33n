/**
 * Phase 75 WS8 / OC-75 — comfort & automation workspace closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { buildNavGroups, collectSidebarRoutes } from '../lib/navGroups.js'
import { buildLegacyRedirectRoutes, WORKSPACES, resolveWorkspaceTab } from '../lib/workspaces.js'

const uiSrc = join(process.cwd(), 'src')
const repoDocs = join(process.cwd(), '..', 'docs')

describe('Phase 75 WS8 / OC-75 — comfort workspace closure', () => {
  const groups = buildNavGroups()
  const routes = collectSidebarRoutes(groups)

  it('sidebar has no Advanced group; comfort is one workspace entry', () => {
    expect(groups.map((g) => g.label)).toEqual(['Today', 'Grow & operate', 'More'])
    expect(routes).not.toContain('/schedules')
    expect(routes).not.toContain('/automation')
    expect(routes).not.toContain('/setpoints')
    const grow = groups.find((g) => g.label === 'Grow & operate')
    expect(grow.items.some((i) => i.to === '/comfort-targets' && i.label === 'Comfort & automation')).toBe(true)
  })

  it('comfort workspace has four tabs and absorbs legacy Advanced routes', () => {
    const ws = WORKSPACES.comfort
    expect(ws.route).toBe('/comfort-targets')
    expect(ws.tabs.map((t) => t.id)).toEqual(['comfort', 'schedules', 'automations', 'raw'])
    expect(ws.absorbs['/schedules']).toEqual({ tab: 'schedules' })
    expect(ws.absorbs['/automation']).toEqual({ tab: 'automations' })
    expect(ws.absorbs['/setpoints']).toEqual({ tab: 'raw' })
  })

  it('legacy comfort hub tab ids map to workspace tabs', () => {
    expect(resolveWorkspaceTab('comfort', 'bands')).toBe('comfort')
    expect(resolveWorkspaceTab('comfort', 'rules')).toBe('automations')
    expect(resolveWorkspaceTab('comfort', 'schedules')).toBe('schedules')
    expect(resolveWorkspaceTab('comfort', 'raw')).toBe('raw')
  })

  it('/schedules, /automation, /setpoints redirect into comfort workspace tabs', () => {
    for (const [legacy, tab] of [
      ['/schedules', 'schedules'],
      ['/automation', 'automations'],
      ['/setpoints', 'raw'],
    ]) {
      const entry = buildLegacyRedirectRoutes().find((r) => r.path === legacy)
      expect(entry).toBeTruthy()
      const result = entry.redirect({
        path: legacy,
        query: { zone_id: '2' },
        hash: '',
        fullPath: legacy,
        matched: [],
        meta: {},
        params: {},
      })
      expect(result.path).toBe('/comfort-targets')
      expect(result.query.tab).toBe(tab)
      expect(result.query.zone_id).toBe('2')
    }
  })

  it('ComfortWorkspace wraps shell and tab bodies', () => {
    const src = readFileSync(join(uiSrc, 'views/workspaces/ComfortWorkspace.vue'), 'utf8')
    expect(src).toContain('WorkspaceShell')
    expect(src).toContain('workspace-id="comfort"')
    expect(src).toContain('ComfortTargetsHub')
    expect(src).toContain('Schedules')
    expect(src).toContain('Automation')
    expect(src).toContain('Setpoints')
  })

  it('ZoneAdvancedHint links to single comfort workspace', () => {
    const hint = readFileSync(join(uiSrc, 'components/ZoneAdvancedHint.vue'), 'utf8')
    expect(hint).toContain('/comfort-targets')
    expect(hint).not.toContain('/automation')
    expect(hint).not.toContain('/schedules')
  })

  it('operator-tour documents comfort workspace (Phase 75)', () => {
    const tour = readFileSync(join(repoDocs, 'operator-tour.md'), 'utf8')
    expect(tour).toMatch(/7h\. Comfort/i)
  })
})
