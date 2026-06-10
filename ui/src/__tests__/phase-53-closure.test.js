/**
 * Phase 53 WS6 / OC-53 — grow + stock + money closure (Vitest bundle guard).
 * Individual workstreams: phase-53-ws1 … ws5 test files.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'
import {
  buildStartGrowPayload,
  buildPostHarvestCompareRoute,
  activeCycleForZone,
} from '../lib/growHub.js'
import { nextQuantityAfterRestock } from '../lib/suppliesHub.js'
import { activeCyclesForZone, buildAutologMoneyRows } from '../lib/moneyHub.js'
import { computeFirstRunChecklist } from '../lib/firstRunChecklist.js'
import {
  buildZoneGrowStripStarters,
  buildSuppliesHubStarters,
  buildMoneyHubStarters,
} from '../lib/guardianStarters.js'
import { relatedTo } from '../lib/navRelations.js'
import { computeFarmMorningSnapshot } from '../lib/farmGrowSummary.js'

const uiSrc = join(process.cwd(), 'src')
const repoDocs = join(process.cwd(), '..', 'docs')

describe('Phase 53 WS6 / OC-53 — grow + stock + money closure', () => {
  it('documents operator-tour §7c, §6i, and architecture §7.0q as shipped', () => {
    const tour = readFileSync(join(repoDocs, 'operator-tour.md'), 'utf8')
    const arch = readFileSync(join(repoDocs, 'farm-guardian-architecture.md'), 'utf8')
    const plan = readFileSync(
      join(repoDocs, 'plans/phase_53_grow_stock_money_closure.plan.md'),
      'utf8',
    )
    expect(tour).toContain('### 7c. Grow + stock + money closure (Phase 53)')
    expect(tour).toContain('**Shipped.**')
    expect(tour).toContain('Harvest weigh-in')
    expect(tour).toContain('phase-53-closure.test.js')
    expect(tour).toContain('### 6i. Guardian on grow closure (Phase 53 — shipped)')
    expect(arch).toContain('### 7.0q Grow + stock + money closure (Phase 53 — shipped)')
    expect(arch).not.toContain('7.0q Grow + stock + money closure (Phase 53 — planned)')
    expect(plan).toMatch(/ws6-docs-tests[\s\S]*status: completed/)
    expect(plan).toContain('**Shipped (WS1–WS6).**')
  })

  it('OC-53 row is closed in operational closure doc', () => {
    const oc = readFileSync(
      join(repoDocs, 'plans/phase_35_37_operational_closure.plan.md'),
      'utf8',
    )
    expect(oc).toContain('## Phase 53 — Grow + stock + money closure')
    expect(oc).toMatch(/oc-53-closure[\s\S]*status: completed/)
    expect(oc).toContain('phase-53-closure.test.js')
  })

  it('roadmap marks Phase 53 shipped', () => {
    const hub = readFileSync(join(repoDocs, 'plans/phase_53_59_roadmap.plan.md'), 'utf8')
    const farmer = readFileSync(join(repoDocs, 'plans/farmer_ux_roadmap_40_plus.plan.md'), 'utf8')
    expect(hub).toContain('**53** ✅')
    expect(hub).toMatch(/OC-53[\s\S]*✅/)
    expect(farmer).toContain('**53** ✅')
  })

  it('grow helpers support start, harvest compare, and active cycle lookup', () => {
    const cycles = [
      { id: 1, zone_id: 2, is_active: false },
      { id: 2, zone_id: 2, is_active: true, name: 'Spring' },
    ]
    expect(activeCycleForZone(cycles, 2)?.id).toBe(2)
    const start = buildStartGrowPayload({ zoneId: 2, strain: 'Basil', name: 'Basil run' })
    expect(start.is_active).toBe(true)
    expect(buildPostHarvestCompareRoute(5, cycles, 2, 2).query.ids).toBe('2,1')
  })

  it('supplies restock and money tagging helpers work end-to-end', () => {
    expect(nextQuantityAfterRestock(10, 5)).toBe(15)
    const cycles = [{ id: 9, zone_id: 3, is_active: true, strain_or_variety: 'OG' }]
    expect(activeCyclesForZone(cycles, 3)).toHaveLength(1)
    const autolog = buildAutologMoneyRows([
      {
        id: 1,
        transaction_date: '2026-06-01',
        amount: 3,
        category: 'miscellaneous',
        related_table_name: 'mixing_events',
        related_record_id: 2,
      },
    ])
    expect(autolog).toHaveLength(1)
    expect(autolog[0].isAutolog).toBe(true)
  })

  it('first-run checklist and nav relations include grow closure rows', () => {
    const items = computeFirstRunChecklist({
      farmId: 1,
      zones: [{ id: 2 }],
      includeGrowClosure: true,
    })
    expect(items.some((i) => i.id === 'start_grow')).toBe(true)
    expect(items.some((i) => i.id === 'restock_input')).toBe(true)
    expect(items.some((i) => i.id === 'log_receipt')).toBe(true)
    expect(relatedTo('/plants')).toContain('/zones')
    expect(relatedTo('/zones')).toContain('/hardware')
  })

  it('Guardian starters cover grow strip, supplies, and money', () => {
    const grow = buildZoneGrowStripStarters({
      zone: { id: 1, name: 'Veg' },
      activeCycle: { id: 5, name: 'Run A' },
      farmId: 3,
    })
    expect(grow.some((s) => s.id === 'grow-room-cost')).toBe(true)
    const supplies = buildSuppliesHubStarters({
      lowStockRows: [{ inputName: 'OHN' }],
      zones: [{ id: 1, name: 'Veg' }],
    })
    expect(supplies.some((s) => s.id === 'restock-first')).toBe(true)
    expect(buildMoneyHubStarters().some((s) => s.id === 'month-spend-by-category')).toBe(true)
  })

  it('dashboard morning strip exposes month spend chip', () => {
    const snap = computeFarmMorningSnapshot({ monthExpenses: 42.5 })
    const chip = snap.chips.find((c) => c.id === 'month-spend')
    expect(chip?.to).toEqual({ path: '/money', query: { tab: 'summary' } })
    expect(chip?.value).toContain('42.50')
  })

  it('closure Vitest files and hub components exist', () => {
    for (const f of [
      '__tests__/phase-53-ws1-grow.test.js',
      '__tests__/phase-53-ws2-supplies.test.js',
      '__tests__/phase-53-ws3-money.test.js',
      '__tests__/phase-53-ws4-crosslinks.test.js',
      '__tests__/phase-53-ws5-guardian.test.js',
      '__tests__/phase-53-closure.test.js',
      'lib/growHub.js',
      'components/ZoneCurrentGrowStrip.vue',
      'components/StartGrowWizard.vue',
      'components/HarvestWeighIn.vue',
      'components/PostHarvestScreen.vue',
      'components/ZoneGrowConnectionLine.vue',
      'components/ZoneGrowCostPeek.vue',
      'views/SuppliesHub.vue',
      'views/MoneyHub.vue',
    ]) {
      expect(existsSync(join(uiSrc, f)), `missing ${f}`).toBe(true)
    }
  })
})
