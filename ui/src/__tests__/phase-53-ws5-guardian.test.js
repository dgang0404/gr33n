/**
 * Phase 53 WS5 — grow/stock/money Guardian starters (Vitest bundle guard).
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'
import {
  buildZoneGrowStripStarters,
  buildHarvestFlowStarters,
  buildSuppliesHubStarters,
  buildMoneyHubStarters,
} from '../lib/guardianStarters.js'

const uiSrc = join(process.cwd(), 'src')

describe('Phase 53 WS5 — Guardian starters', () => {
  const zone = { id: 4, name: 'Flower Room' }
  const activeCycle = { id: 22, zone_id: 4, name: 'OG Spring', strain_or_variety: 'OG Kush', is_active: true }
  const priorCycle = { id: 18, zone_id: 4, name: 'OG Winter', is_active: false }

  it('zone grow strip offers cost and compare starters', () => {
    const starters = buildZoneGrowStripStarters({
      zone,
      activeCycle,
      farmId: 7,
      priorHarvestedCycle: priorCycle,
    })
    expect(starters.length).toBeGreaterThanOrEqual(3)
    const vpd = starters.find((s) => s.id === 'vpd-on-target')
    expect(vpd?.label).toBe('Is my VPD on target?')
    expect(vpd?.message).toContain('grow_advisor')
    const cost = starters.find((s) => s.id === 'grow-room-cost')
    expect(cost?.message).toContain('Flower Room')
    expect(cost?.message).toContain('OG Spring')
    expect(cost?.message).toContain('summarize_cycle_cost')
    expect(cost?.contextRef.crop_cycle_id).toBe(22)
    const compare = starters.find((s) => s.id === 'compare-last-cycle')
    expect(compare?.message).toContain('OG Winter')
    expect(compare?.contextRef.compare_path).toBe('/farms/7/crop-cycles/compare')
  })

  it('harvest flow offers prior yield starter', () => {
    const starters = buildHarvestFlowStarters({
      zone,
      activeCycle,
      priorHarvestedCycle: priorCycle,
    })
    expect(starters).toHaveLength(2)
    const priorYield = starters.find((s) => s.id === 'prior-yield')
    expect(priorYield?.label).toBe('Last run yield')
    expect(priorYield?.message).toContain('OG Winter')
    expect(priorYield?.contextRef.prior_crop_cycle_id).toBe(18)
  })

  it('supplies hub prioritizes restock-first chip when low stock', () => {
    const starters = buildSuppliesHubStarters({
      lowStockRows: [{ inputName: 'OHN', remaining: 1, threshold: 3 }],
      zones: [zone],
    })
    expect(starters[0].id).toBe('restock-first')
    expect(starters[0].message).toContain('restock first')
  })

  it('money hub offers category spend summary chip', () => {
    const starters = buildMoneyHubStarters()
    expect(starters).toHaveLength(3)
    expect(starters[0].id).toBe('month-spend-by-category')
    expect(starters[0].message).toContain('summarize_farm_spending')
  })

  it('components wire grow and harvest starters', () => {
    const strip = readFileSync(join(uiSrc, 'components/ZoneCurrentGrowStrip.vue'), 'utf8')
    const harvest = readFileSync(join(uiSrc, 'components/HarvestWeighIn.vue'), 'utf8')
    expect(strip).toContain('buildZoneGrowStripStarters')
    expect(strip).toContain('data-test="zone-grow-strip-starters"')
    expect(harvest).toContain('buildHarvestFlowStarters')
    expect(harvest).toContain('data-test="harvest-flow-starters"')
  })

  it('phase-53 ws5 test file exists', () => {
    expect(existsSync(join(uiSrc, '__tests__/phase-53-ws5-guardian.test.js'))).toBe(true)
  })
})
