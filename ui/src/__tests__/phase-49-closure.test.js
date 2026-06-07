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
  const groups = buildNavGroups('/farms/1/crop-cycles/compare')

  it('disambiguates feeding nav labels', () => {
    const grow = groups.find((g) => g.label === 'Grow')
    const ops = groups.find((g) => g.label === 'Operations')
    const advanced = groups.find((g) => g.label === 'Advanced')

    expect(grow.items.find((i) => i.to === '/feeding')?.label).toBe('Feed & water')
    expect(ops.items.find((i) => i.to === '/operations/feeding')?.label).toBe('Feeding admin')
    expect(advanced.items.find((i) => i.to === '/fertigation')?.label).toBe('Fertigation')
  })

  it('SideNav implements related-route hover affordance', () => {
    const sideNav = readFileSync(join(process.cwd(), 'src/components/SideNav.vue'), 'utf8')
    expect(sideNav).toContain('navRelations.js')
    expect(sideNav).toContain('nav-related')
    expect(sideNav).toContain('prefers-reduced-motion')
    expect(sideNav).toContain('hoveredRoute')
  })

  it('navRelations map exists', () => {
    const rel = readFileSync(join(process.cwd(), 'src/lib/navRelations.js'), 'utf8')
    expect(rel).toContain("'/zones'")
    expect(rel).toContain("relatedTo")
  })

  it('operator-tour documents Fertigation in Advanced nav', () => {
    const tour = readFileSync(join(repoDocs, 'operator-tour.md'), 'utf8')
    expect(tour).toContain('Fertigation')
    expect(tour).toContain('Feeding admin')
  })

  it('OC-49 marked completed in operational closure plan', () => {
    const closure = readFileSync(
      join(repoDocs, 'plans/phase_35_37_operational_closure.plan.md'),
      'utf8',
    )
    expect(closure).toContain('oc-49-closure')
    expect(closure).toMatch(/oc-49-closure[\s\S]*status: completed/)
    expect(closure).toContain('## Phase 49 — Sidebar nav polish')
  })

  it('phase 49 plan marks all workstreams completed', () => {
    const plan = readFileSync(
      join(repoDocs, 'plans/phase_49_sidebar_nav_polish.plan.md'),
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
