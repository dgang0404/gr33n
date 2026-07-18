/**
 * Phase 49 WS4 / OC-49 — sidebar nav polish closure (Vitest bundle guard).
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { buildNavGroups } from '../lib/navGroups.js'

const repoRoot = join(process.cwd(), '..')
const repoDocs = join(repoRoot, 'docs')

describe('Phase 49 WS4 / OC-49 — sidebar nav closure', () => {
  const groups = buildNavGroups()

  it('disambiguates feeding via zone tabs (Phase 78)', () => {
    const grow = groups.find((g) => g.label === 'Grow & operate')
    expect(grow.items.find((i) => i.to === '/zones')?.label).toBeTruthy()
    expect(grow.items.some((i) => i.to === '/feed-water')).toBe(false)
    expect(groups.find((g) => g.label === 'Advanced')).toBeUndefined()
    expect(grow.items.some((i) => i.to === '/comfort-targets')).toBe(true)
  })

  it('SideNav wiggles only the hinted sidebar tab (no related-route fan-out)', () => {
    const sideNav = readFileSync(join(process.cwd(), 'src/components/SideNav.vue'), 'utf8')
    expect(sideNav).not.toContain('navRelations.js')
    expect(sideNav).toContain('nav-related')
    expect(sideNav).toContain('isHighlightedNav')
    expect(sideNav).toContain('navHighlight')
    expect(sideNav).toContain('useNavHighlightStore')
    expect(sideNav).toContain('prefers-reduced-motion')
    expect(sideNav).not.toContain('hoveredRoute')
    expect(sideNav).not.toContain('relatedTo')
  })

  it('navRelations map exists', () => {
    const rel = readFileSync(join(process.cwd(), 'src/lib/navRelations.js'), 'utf8')
    expect(rel).toContain("'/zones'")
    expect(rel).toContain("relatedTo")
  })

  it('SideNav wiggles the sidebar destination of hovered in-page links', () => {
    const sideNav = readFileSync(join(process.cwd(), 'src/components/SideNav.vue'), 'utf8')
    expect(sideNav).toContain('navHighlight')
    expect(sideNav).toContain('useNavHighlightStore')
  })

  it('nav-hint directive is registered globally', () => {
    const main = readFileSync(join(process.cwd(), 'src/main.js'), 'utf8')
    expect(main).toContain("app.directive('nav-hint', navHint)")
  })

  it('zone water in-page links carry v-nav-hint to their sidebar destination', () => {
    const needSection = readFileSync(join(process.cwd(), 'src/components/ZoneNeedSection.vue'), 'utf8')
    expect(needSection).toContain('v-nav-hint="link.to"')
    const growStory = readFileSync(join(process.cwd(), 'src/components/ZoneWaterGrowStory.vue'), 'utf8')
    expect(growStory).toContain('v-nav-hint="advancedFeedingLink"')
  })

  it('operator-tour documents workspaces and feeding tabs', () => {
    const tour = readFileSync(join(repoDocs, 'operator-tour.md'), 'utf8')
    expect(tour).toMatch(/workspace/i)
    expect(tour).toContain('Feed & Water')
  })

  it('OC-49 marked completed in operational closure plan', () => {
    const closure = readFileSync(
      join(repoDocs, 'plans/archive/phase_35_37_operational_closure.plan.md'),
      'utf8',
    )
    expect(closure).toContain('oc-49-closure')
    expect(closure).toMatch(/oc-49-closure[\s\S]*status: completed/)
    expect(closure).toContain('## Phase 49 — Sidebar nav polish')
  })

  it('phase 49 plan marks all workstreams completed', () => {
    const plan = readFileSync(
      join(repoDocs, 'plans/archive/phase_49_sidebar_nav_polish.plan.md'),
      'utf8',
    )
    for (const id of [
      'ws1-fertigation-rename',
      'ws2-route-relationship-map',
      'ws3-hover-wiggle',
      'ws4-docs-tests',
    ]) {
      expect(plan).toContain(`id: ${id}`)
      expect(plan).toMatch(new RegExp(`${id}[\\s\\S]*status: completed`))
    }
    expect(plan).toContain('**Shipped.**')
  })
})
