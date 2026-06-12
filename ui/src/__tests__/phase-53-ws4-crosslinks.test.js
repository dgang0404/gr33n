/**
 * Phase 53 WS4 — cross-links, checklist, nav relations (Vitest bundle guard).
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'
import { computeFirstRunChecklist, isFirstRunComplete } from '../lib/firstRunChecklist.js'
import { relatedTo } from '../lib/navRelations.js'
import { isAutologgedTransaction } from '../lib/moneyHub.js'

const uiSrc = join(process.cwd(), 'src')

describe('Phase 53 WS4 — cross-links & checklist', () => {
  it('extends getting started with optional grow/stock/money rows', () => {
    const items = computeFirstRunChecklist({
      farmId: 9,
      zones: [{ id: 2, name: 'Flower' }],
      includeGrowClosure: true,
    })
    expect(items).toHaveLength(7)
    expect(items.find((i) => i.id === 'start_grow')?.optional).toBe(true)
    expect(items.find((i) => i.id === 'restock_input')?.to).toBe('/operations/supplies')
    expect(items.find((i) => i.id === 'log_receipt')?.to).toBe('/money')
    expect(items.find((i) => i.id === 'start_grow')?.to).toEqual({
      path: '/zones/2',
      query: { start_grow: '1' },
    })
  })

  it('marks grow closure done from farm data', () => {
    const items = computeFirstRunChecklist({
      farmId: 1,
      zones: [{ id: 1 }],
      cropCycles: [{ id: 5, zone_id: 1, is_active: true }],
      nfBatches: [{ current_quantity_remaining: 12, initial_quantity: 5 }],
      costTransactions: [
        { id: 1, amount: 20, transaction_date: '2026-06-01', category: 'miscellaneous', description: 'Soil' },
      ],
      includeGrowClosure: true,
    })
    expect(items.find((i) => i.id === 'start_grow')?.done).toBe(true)
    expect(items.find((i) => i.id === 'restock_input')?.done).toBe(true)
    expect(items.find((i) => i.id === 'log_receipt')?.done).toBe(true)
    expect(isFirstRunComplete(items)).toBe(false)
  })

  it('base first-run completion ignores optional rows', () => {
    const items = computeFirstRunChecklist({
      farmId: 3,
      zones: [{ id: 1 }],
      devices: [{ id: 1 }],
      setpoints: [{ min_value: 1 }],
      schedules: [{ is_active: true }],
      includeGrowClosure: true,
    })
    expect(isFirstRunComplete(items)).toBe(true)
    expect(items.filter((i) => !i.optional).every((i) => i.done)).toBe(true)
    expect(items.filter((i) => i.optional).some((i) => !i.done)).toBe(true)
  })

  it('navRelations links grow, plants, money, and supplies', () => {
    expect(relatedTo('/plants')).toContain('/zones')
    expect(relatedTo('/zones')).toContain('/comfort-targets')
    expect(relatedTo('/operations/money')).toContain('/money')
    expect(relatedTo('/fertigation')).toContain('/zones')
  })

  it('phase-53 CTAs use v-nav-hint on harvest, restock, receipt, compare', () => {
    const strip = readFileSync(join(uiSrc, 'components/ZoneCurrentGrowStrip.vue'), 'utf8')
    const money = readFileSync(join(uiSrc, 'views/MoneyHub.vue'), 'utf8')
    const supplies = readFileSync(join(uiSrc, 'views/SuppliesHub.vue'), 'utf8')
    const post = readFileSync(join(uiSrc, 'components/PostHarvestScreen.vue'), 'utf8')
    const morning = readFileSync(join(uiSrc, 'components/FarmMorningStrip.vue'), 'utf8')
    expect(strip).toContain('v-nav-hint')
    expect(strip).toContain("v-nav-hint=\"'/plants'\"")
    expect(money).toContain('data-test="money-save-receipt"')
    expect(money).toContain('v-nav-hint')
    expect(supplies).toContain('data-test="supplies-restock-btn"')
    expect(supplies).toContain('v-nav-hint')
    expect(supplies).toContain('data-test="supplies-refill-task"')
    expect(post).toContain('v-nav-hint="compareRoute"')
    expect(morning).toContain('v-nav-hint="chip.to"')
  })

  it('operator guide documents post-harvest and restock paths', () => {
    const guide = readFileSync(join(uiSrc, 'views/OperatorGuide.vue'), 'utf8')
    expect(guide).toContain('/operations/supplies')
    expect(guide).toContain('Harvest weigh-in')
    expect(guide).toContain('/money')
  })

  it('manual receipt detection excludes autolog rows', () => {
    const tx = {
      related_table_name: 'mixing_events',
      related_record_id: 1,
      amount: 5,
      transaction_date: '2026-06-01',
    }
    expect(isAutologgedTransaction(tx)).toBe(true)
    const items = computeFirstRunChecklist({
      costTransactions: [tx],
      includeGrowClosure: true,
    })
    expect(items.find((i) => i.id === 'log_receipt')?.done).toBe(false)
  })

  it('phase-53 ws4 test file exists', () => {
    expect(existsSync(join(uiSrc, '__tests__/phase-53-ws4-crosslinks.test.js'))).toBe(true)
  })
})
