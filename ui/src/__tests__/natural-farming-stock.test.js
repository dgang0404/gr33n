/**
 * Phase 209 WS5 — on-hand stock helpers.
 */
import { describe, it, expect } from 'vitest'
import {
  filterReadyBatches,
  lowStockFromReady,
  stockRows,
} from '../lib/naturalFarmingStock.js'

describe('naturalFarmingStock', () => {
  const inputs = [{ id: 1, name: 'JMS', unit_cost: 2.5, unit_cost_currency: 'USD' }]
  const batches = [
    { id: 10, input_definition_id: 1, status: 'ready_for_use', batch_identifier: 'A', current_quantity_remaining: 2, low_stock_threshold: 5 },
    { id: 11, input_definition_id: 1, status: 'fermenting_brewing', batch_identifier: 'B', current_quantity_remaining: 20, low_stock_threshold: 5 },
    { id: 12, input_definition_id: 1, status: 'partially_used', batch_identifier: 'C', current_quantity_remaining: 8, low_stock_threshold: 3 },
  ]

  it('keeps only ready and partially used batches', () => {
    expect(filterReadyBatches(batches).map((b) => b.id)).toEqual([10, 12])
  })

  it('flags low stock on ready batches only', () => {
    const low = lowStockFromReady(batches, inputs)
    expect(low).toHaveLength(1)
    expect(low[0].batch.id).toBe(10)
  })

  it('builds supply rows with unit cost label', () => {
    const rows = stockRows(batches, inputs)
    expect(rows).toHaveLength(2)
    expect(rows[0].lowStock).toBe(true)
    expect(rows[0].unitCostLabel).toBe('$2.50')
  })
})
