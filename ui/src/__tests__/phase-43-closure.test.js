/**
 * Phase 43 WS7 / OC-43 — operations hub docs closure (Vitest bundle guard).
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'
import { buildNavGroups } from '../lib/navGroups.js'
import { buildSuppliesHubStarters, buildMoneyHubStarters } from '../lib/guardianStarters.js'
import { listLowStockBatches, buildSupplyRows } from '../lib/suppliesHub.js'
import { buildProgramAdminCards } from '../lib/feedingAdminHub.js'
import { computeMonthSummary } from '../lib/moneyHub.js'

const uiSrc = join(process.cwd(), 'src')
const repoDocs = join(process.cwd(), '..', 'docs')

describe('Phase 43 WS7 / OC-43 — operations hub closure', () => {
  it('documents operator-tour §7 and architecture §7.0i as shipped', () => {
    const tour = readFileSync(join(repoDocs, 'operator-tour.md'), 'utf8')
    const arch = readFileSync(join(repoDocs, 'farm-guardian-architecture.md'), 'utf8')
    expect(tour).toContain('## 7. Supplies, feeding & money (Phase 43)')
    expect(tour).toContain('**Shipped.**')
    expect(tour).toContain('/operations/supplies')
    expect(arch).toContain('### 7.0i Operations hub — supplies, feeding, money (Phase 43)')
    expect(arch).not.toContain('7.0i Operations hub — supplies, feeding, money (Phase 43 — planned)')
    expect(tour).toContain('summarize_farm_low_stock')
  })

  it('WS8 Guardian starters exist for operations hubs', () => {
    const low = buildSuppliesHubStarters({
      lowStockRows: [{ inputName: 'OHN' }],
      zones: [{ id: 1, name: 'Veg' }],
    })
    expect(low.some((s) => s.id === 'whats-running-low')).toBe(true)
    expect(buildMoneyHubStarters().length).toBeGreaterThanOrEqual(1)
  })

  it('Grow & operate nav exposes workspace hubs (Phase 68)', () => {
    const grow = buildNavGroups().find((g) => g.label === 'Grow & operate')
    expect(grow).toBeTruthy()
    const paths = grow.items.map((i) => i.to)
    expect(paths).toContain('/zones')
    expect(paths).toContain('/money')
    expect(paths).not.toContain('/feed-water')
    expect(paths).not.toContain('/hardware')
  })

  it('supplies hub helpers surface low-stock rows', () => {
    const rows = listLowStockBatches([
      { id: 1, input_definition_id: 10, current_quantity_remaining: 1, low_stock_threshold: 5 },
    ], [{ id: 10, name: 'OHN' }])
    expect(rows).toHaveLength(1)
    expect(rows[0].inputName).toBe('OHN')
    const supply = buildSupplyRows([
      { id: 1, input_definition_id: 10, current_quantity_remaining: 1, low_stock_threshold: 5, status: 'ready_for_use' },
    ], [{ id: 10, name: 'OHN' }])
    expect(supply[0].lowStock).toBe(true)
  })

  it('feeding admin builds program cards with zone names', () => {
    const cards = buildProgramAdminCards(
      [{ id: 1, name: 'Bloom', target_zone_id: 3, is_active: true, irrigation_only: true }],
      [{ id: 3, name: 'Flower' }],
      [],
    )
    expect(cards[0].zoneName).toBe('Flower')
    expect(cards[0].irrigationOnly).toBe(true)
  })

  it('money hub computes month summary', () => {
    const ref = new Date('2026-06-15')
    const summary = computeMonthSummary([
      { transaction_date: '2026-06-02', amount: 40, is_income: false },
    ], ref)
    expect(summary.expenses).toBe(40)
    expect(summary.monthLabel).toContain('June')
  })

  it('closure Vitest files exist', () => {
    for (const f of [
      '__tests__/supplies-hub.test.js',
      '__tests__/feeding-admin-hub.test.js',
      '__tests__/money-hub.test.js',
      '__tests__/nav-groups.test.js',
      '__tests__/farm-grow-summary.test.js',
      '__tests__/zone-feeding-water.test.js',
      'views/SuppliesHub.vue',
      'views/FeedingAdminHub.vue',
      'views/MoneyHub.vue',
    ]) {
      expect(existsSync(join(uiSrc, f))).toBe(true)
    }
  })
})
