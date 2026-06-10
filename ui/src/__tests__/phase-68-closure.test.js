/**
 * Phase 68 WS6 / OC-68 — workspace shell closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { buildNavGroups, collectSidebarRoutes } from '../lib/navGroups.js'

const uiSrc = join(process.cwd(), 'src')
const repoDocs = join(process.cwd(), '..', 'docs')

describe('Phase 68 WS6 / OC-68 — workspace shell closure', () => {
  const groups = buildNavGroups()
  const routes = collectSidebarRoutes(groups)

  it('sidebar uses four compact groups with workspace entries', () => {
    expect(groups.map((g) => g.label)).toEqual(['Today', 'Grow & operate', 'More'])
    const grow = groups.find((g) => g.label === 'Grow & operate')
    expect(grow.items.some((i) => i.to === '/zones')).toBe(true)
    expect(grow.items.some((i) => i.to === '/feed-water' && i.label === 'Feed & water')).toBe(true)
    expect(grow.items.some((i) => i.to === '/comfort-targets' && i.label === 'Comfort & automation')).toBe(true)
    expect(grow.items.some((i) => i.to === '/hardware')).toBe(true)
    expect(grow.items.some((i) => i.to === '/money')).toBe(true)
    expect(routes).not.toContain('/feeding')
    expect(routes).not.toContain('/fertigation')
    expect(routes).not.toContain('/operations/supplies')
  })

  it('WorkspaceShell supports deep-linkable tabs and jump rail', () => {
    const shell = readFileSync(join(uiSrc, 'components/WorkspaceShell.vue'), 'utf8')
    expect(shell).toContain('data-test="workspace-shell"')
    expect(shell).toContain('route.query.tab')
    expect(shell).toContain('v-nav-hint')
    expect(shell).toContain('prefers-reduced-motion')
    expect(shell).toContain('Jump to')
  })

  it('workspace views wrap existing pages', () => {
    for (const file of [
      'views/workspaces/ZonesWorkspace.vue',
      'views/workspaces/FeedWaterWorkspace.vue',
      'views/workspaces/MoneyWorkspace.vue',
      'views/workspaces/HardwareWorkspace.vue',
      'views/workspaces/ComfortWorkspace.vue',
      'views/workspaces/HelpWorkspace.vue',
    ]) {
      const src = readFileSync(join(uiSrc, file), 'utf8')
      expect(src).toContain('WorkspaceShell')
    }
  })

  it('nav-hint maps legacy paths to workspace sidebar routes', () => {
    const navHint = readFileSync(join(uiSrc, 'directives/navHint.js'), 'utf8')
    expect(navHint).toContain('canonicalSidebarPath')
  })

  it('operator-tour documents workspaces', () => {
    const tour = readFileSync(join(repoDocs, 'operator-tour.md'), 'utf8')
    expect(tour).toMatch(/workspace/i)
  })
})
