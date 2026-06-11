/**
 * Phase 76 WS6 / OC-76 — Today dashboard & mobile nav alignment closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'
import { mobileBottomNav } from '../lib/navGroups.js'
import { computeFarmMorningSnapshot } from '../lib/farmGrowSummary.js'

const repoDocs = join(process.cwd(), '..', 'docs')
const uiSrc = join(process.cwd(), 'src')

describe('Phase 76 WS6 / OC-76 — Today dashboard alignment', () => {
  it('mobile bottom nav uses Targets instead of Alerts', () => {
    expect(mobileBottomNav.map((i) => i.to)).toEqual([
      '/',
      '/zones',
      '/comfort-targets',
      '/money',
      '/settings',
    ])
    expect(mobileBottomNav.find((i) => i.to === '/alerts')).toBeUndefined()
    expect(mobileBottomNav.find((i) => i.to === '/comfort-targets')?.label).toBe('Targets')
  })

  it('morning chips link to workspaces not legacy routes', () => {
    const today = new Date().toISOString().slice(0, 10)
    const snap = computeFarmMorningSnapshot({
      tasks: [{ status: 'todo', due_date: today, zone_id: 3 }],
      alerts: [{ is_read: false }],
      schedules: [],
      devices: [],
      zones: [{ id: 3, name: 'Veg' }],
      programs: [],
      queueDepth: 1,
      lowStockCount: 1,
      monthExpenses: 42.5,
    })
    expect(snap.chips.find((c) => c.id === 'tasks-due').to).toEqual({
      path: '/zones/3',
      query: { tab: 'ops', ops: 'tasks' },
    })
    expect(snap.chips.find((c) => c.id === 'feeding')?.to).toEqual({
      path: '/zones/3',
      query: { tab: 'water' },
    })
    expect(snap.chips.find((c) => c.id === 'low-stock').to).toEqual({
      path: '/money',
      query: { tab: 'supplies' },
    })
    expect(snap.chips.find((c) => c.id === 'next-schedule').to).toEqual({
      path: '/comfort-targets',
      query: { tab: 'schedules' },
    })
  })

  it('dashboard ops starters use money supplies workspace', () => {
    const starters = readFileSync(join(uiSrc, 'lib/guardianStarters.js'), 'utf8')
    expect(starters).toContain("operationsRouteRef('/money?tab=supplies'")
  })

  it('plan and operator-tour document Phase 76 shipped', () => {
    const plan = readFileSync(join(repoDocs, 'plans/phase_76_today_dashboard_nav_alignment.plan.md'), 'utf8')
    const tour = readFileSync(join(repoDocs, 'operator-tour.md'), 'utf8')
    const roadmap = readFileSync(join(repoDocs, 'plans/phase_68_73_spa_workspace_roadmap.plan.md'), 'utf8')
    expect(plan).toContain('**Shipped.**')
    expect(tour).toMatch(/7i\. Today dashboard alignment \(Phase 76/i)
    expect(tour).toMatch(/Shipped/)
    expect(roadmap).toMatch(/OC-76.*Shipped/)
    expect(existsSync(join(uiSrc, 'lib/dashboardWorkspaceLinks.js'))).toBe(true)
  })
})
