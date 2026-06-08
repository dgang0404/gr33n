import { describe, it, expect } from 'vitest'
import {
  isBatchLowStock,
  listLowStockBatches,
  buildSupplyRows,
  filterLowStockAlerts,
  nextQuantityAfterRestock,
  formatInputUnitCost,
  buildRefillTaskPayload,
} from '../lib/suppliesHub.js'

describe('Phase 43 WS2 — supplies hub helpers', () => {
  const inputs = [{ id: 1, name: 'OHN' }, { id: 2, name: 'JMS' }]

  it('detects batches below low-stock threshold', () => {
    expect(isBatchLowStock({
      current_quantity_remaining: 2,
      low_stock_threshold: 5,
    })).toBe(true)
    expect(isBatchLowStock({
      current_quantity_remaining: 5,
      low_stock_threshold: 5,
    })).toBe(false)
    expect(isBatchLowStock({ current_quantity_remaining: 1 })).toBe(false)
  })

  it('lists low-stock batches with input names', () => {
    const rows = listLowStockBatches([
      { id: 10, input_definition_id: 1, current_quantity_remaining: 1, low_stock_threshold: 3, batch_identifier: 'A' },
      { id: 11, input_definition_id: 2, current_quantity_remaining: 10, low_stock_threshold: 2, batch_identifier: 'B' },
    ], inputs)
    expect(rows).toHaveLength(1)
    expect(rows[0].inputName).toBe('OHN')
    expect(rows[0].remaining).toBe(1)
  })

  it('builds supply rows with low-stock sorted first', () => {
    const rows = buildSupplyRows([
      { id: 1, input_definition_id: 2, current_quantity_remaining: 10, status: 'ready_for_use' },
      { id: 2, input_definition_id: 1, current_quantity_remaining: 1, low_stock_threshold: 5, status: 'partially_used' },
    ], inputs)
    expect(rows[0].lowStock).toBe(true)
    expect(rows[0].inputName).toBe('OHN')
    expect(rows[1].lowStock).toBe(false)
    expect(rows[0].scope).toBe('farm')
  })

  it('adds restock quantity to current on hand', () => {
    expect(nextQuantityAfterRestock(2, 3)).toBe(5)
    expect(nextQuantityAfterRestock(null, 1)).toBe(1)
    expect(nextQuantityAfterRestock(5, 0)).toBeNull()
  })

  it('formats input unit cost for display', () => {
    expect(formatInputUnitCost({ unit_cost: 12.5, unit_cost_currency: 'USD' })).toBe('$12.50')
    expect(formatInputUnitCost({})).toBeNull()
  })

  it('builds refill task payload in plain language', () => {
    const p = buildRefillTaskPayload({
      inputName: 'OHN',
      remaining: 1,
      threshold: 5,
    })
    expect(p.title).toBe('Refill OHN')
    expect(p.description).toContain('1')
    expect(p.priority).toBe(2)
  })

  it('includes unit cost on supply rows', () => {
    const rows = buildSupplyRows([
      { id: 1, input_definition_id: 1, current_quantity_remaining: 10, status: 'ready_for_use' },
    ], [{ id: 1, name: 'OHN', unit_cost: 4.25, unit_cost_currency: 'USD' }])
    expect(rows[0].unitCostLabel).toBe('$4.25')
    expect(rows[0].inputDefinitionId).toBe(1)
  })

  it('filters unread inventory_low_stock alerts', () => {
    const alerts = filterLowStockAlerts([
      { id: 1, is_read: false, is_acknowledged: false, triggering_event_source_type: 'inventory_low_stock' },
      { id: 2, is_read: false, is_acknowledged: false, triggering_event_source_type: 'sensor' },
      { id: 3, is_read: true, is_acknowledged: false, triggering_event_source_type: 'inventory_low_stock' },
    ])
    expect(alerts.map((a) => a.id)).toEqual([1])
  })
})
