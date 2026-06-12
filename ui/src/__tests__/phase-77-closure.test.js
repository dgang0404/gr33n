/**
 * Phase 77 WS6 / OC-77 — post-arc UI polish closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'
import { buildNavGroups, collectSidebarRoutes } from '../lib/navGroups.js'
import { WORKSPACES } from '../lib/workspaces.js'
import { buildCompareRoute } from '../lib/growHub.js'

const repoDocs = join(process.cwd(), '..', 'docs')
const uiSrc = join(process.cwd(), 'src')

describe('Phase 77 WS6 / OC-77 — post-arc polish closure', () => {
  const groups = buildNavGroups()
  const routes = collectSidebarRoutes(groups)

  it('sidebar stays compact without orphan More entries', () => {
    expect(groups.map((g) => g.label)).toEqual(['Today', 'Grow & operate', 'More'])
    expect(routes.length).toBeLessThanOrEqual(11)
    expect(routes).toContain('/chat')
    expect(routes).not.toContain('/farm-knowledge')
    expect(routes).not.toContain('/catalog')
    expect(routes.filter((r) => r.includes('crop-cycles/compare'))).toHaveLength(0)

    const more = groups.find((g) => g.label === 'More')
    expect(more.items.map((i) => i.to)).toEqual([
      '/chat',
      '/animals',
      '/aquaponics',
      '/operator-guide',
      '/settings',
    ])
    expect(more.items.find((i) => i.label === 'Help')?.to).toBe('/operator-guide')
    expect(more.items.find((i) => i.label === 'Farm Guardian')?.to).toBe('/chat')
  })

  it('help and money grows workspaces ship', () => {
    expect(WORKSPACES.help.tabs.map((t) => t.id)).toEqual(['guide', 'pi-setup', 'knowledge', 'catalog'])
    expect(WORKSPACES.help.absorbs['/farm-knowledge']).toEqual({ tab: 'knowledge' })
    expect(WORKSPACES.money.tabs.some((t) => t.id === 'grows')).toBe(true)
    expect(existsSync(join(uiSrc, 'views/workspaces/HelpWorkspace.vue'))).toBe(true)
    expect(existsSync(join(uiSrc, 'components/MoneyGrowsSection.vue'))).toBe(true)
    expect(existsSync(join(uiSrc, 'components/FarmConfigCard.vue'))).toBe(true)
  })

  it('compare analytics home is zones plants tab, not sidebar Analytics', () => {
    const plants = readFileSync(join(uiSrc, 'views/Plants.vue'), 'utf8')
    expect(plants).toContain('strains-compare-banner')
    expect(plants).toContain('crop-cycles/compare')
    expect(buildCompareRoute(null, [1])).toEqual({ path: '/zones', query: { tab: 'plants' } })

    const compare = readFileSync(join(uiSrc, 'views/CropCycleCompare.vue'), 'utf8')
    expect(compare).toContain("tab: 'plants'")
  })

  it('Guardian drawer and full page both reachable', () => {
    const drawer = readFileSync(join(uiSrc, 'components/GuardianDrawer.vue'), 'utf8')
    expect(drawer).toContain('Open full chat')
    const nav = buildNavGroups()
    expect(nav.find((g) => g.label === 'More')?.items.some((i) => i.to === '/chat')).toBe(true)
    const settings = readFileSync(join(uiSrc, 'views/Settings.vue'), 'utf8')
    expect(settings).toContain("tab: 'fleet', fleet: 'sensors'")
  })

  it('Today dashboard surfaces farm config card', () => {
    const dash = readFileSync(join(uiSrc, 'views/Dashboard.vue'), 'utf8')
    expect(dash).toContain('FarmConfigCard')
  })

  it('plan and operator-tour document Phase 77 shipped', () => {
    const plan = readFileSync(join(repoDocs, 'plans/phase_77_post_arc_ui_polish.plan.md'), 'utf8')
    const tour = readFileSync(join(repoDocs, 'operator-tour.md'), 'utf8')
    const roadmap = readFileSync(join(repoDocs, 'plans/phase_68_73_spa_workspace_roadmap.plan.md'), 'utf8')
    expect(plan).toContain('**Shipped.**')
    expect(tour).toMatch(/7j\. Post-arc polish \(Phase 77/i)
    expect(tour).toMatch(/Shipped/)
    expect(roadmap).toMatch(/OC-77.*Shipped/)
  })
})
