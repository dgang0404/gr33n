/**
 * Phase 53 WS3 — money closure (Vitest bundle guard).
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'
import {
  isAutologgedTransaction,
  autologPlainLabel,
  autologContextLink,
  buildAutologMoneyRows,
  buildManualMoneyRows,
  activeCyclesForZone,
} from '../lib/moneyHub.js'
import { computeFarmMorningSnapshot } from '../lib/farmGrowSummary.js'
import { relatedTo } from '../lib/navRelations.js'

const uiSrc = join(process.cwd(), 'src')

describe('Phase 53 WS3 — money closure', () => {
  it('detects autolog transactions and builds plain labels', () => {
    const tx = {
      related_table_name: 'mixing_events',
      related_record_id: 9,
      amount: 12,
      transaction_date: '2026-06-01',
      category: 'fertilizers_soil_amendments',
    }
    expect(isAutologgedTransaction(tx)).toBe(true)
    expect(autologPlainLabel(tx)).toBe('From a nutrient mix')
    expect(autologContextLink(tx)?.path).toBe('/operations/feeding')
  })

  it('splits autolog vs manual recent rows', () => {
    const txs = [
      {
        id: 1,
        transaction_date: '2026-06-02',
        amount: 5,
        category: 'miscellaneous',
        related_table_name: 'mixing_events',
        related_record_id: 1,
      },
      {
        id: 2,
        transaction_date: '2026-06-03',
        amount: 20,
        category: 'miscellaneous',
        description: 'OHN restock',
      },
    ]
    expect(buildAutologMoneyRows(txs)).toHaveLength(1)
    expect(buildManualMoneyRows(txs)).toHaveLength(1)
    expect(buildManualMoneyRows(txs)[0].label).toBe('OHN restock')
  })

  it('lists active cycles per zone for receipt tagging', () => {
    const rows = activeCyclesForZone([
      { id: 10, zone_id: 3, is_active: true, strain_or_variety: 'Blue Dream' },
      { id: 11, zone_id: 4, is_active: true },
    ], 3)
    expect(rows).toHaveLength(1)
    expect(rows[0].id).toBe(10)
  })

  it('morning snapshot includes month spend chip', () => {
    const { chips } = computeFarmMorningSnapshot({ monthExpenses: 142.5 })
    expect(chips.some((c) => c.id === 'month-spend' && c.to?.path === '/money')).toBe(true)
  })

  it('MoneyHub and zone cost peek UI exist', () => {
    const money = readFileSync(join(uiSrc, 'views/MoneyHub.vue'), 'utf8')
    expect(money).toContain('data-test="money-tag-zone"')
    expect(money).toContain('data-test="money-tag-cycle"')
    expect(money).toContain('data-test="money-autolog-section"')
    expect(money).toContain('data-test="money-energy-nudge"')
    expect(money).toContain('crop_cycle_id')
    const peek = readFileSync(join(uiSrc, 'components/ZoneGrowCostPeek.vue'), 'utf8')
    expect(peek).toContain('data-test="zone-grow-cost-peek"')
    expect(peek).toContain('loadCropCycleCostSummary')
    expect(existsSync(join(uiSrc, '__tests__/phase-53-ws3-money.test.js'))).toBe(true)
  })

  it('navRelations links money hub to supplies and costs', () => {
    expect(relatedTo('/operations/money')).toContain('/money')
    expect(relatedTo('/operations/money')).toContain('/zones')
  })
})
