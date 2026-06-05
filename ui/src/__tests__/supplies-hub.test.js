import { describe, it, expect } from 'vitest'
import {
  isBatchLowStock,
  listLowStockBatches,
  buildSupplyRows,
  filterLowStockAlerts,
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

  it('filters unread inventory_low_stock alerts', () => {
    const alerts = filterLowStockAlerts([
      { id: 1, is_read: false, is_acknowledged: false, triggering_event_source_type: 'inventory_low_stock' },
      { id: 2, is_read: false, is_acknowledged: false, triggering_event_source_type: 'sensor' },
      { id: 3, is_read: true, is_acknowledged: false, triggering_event_source_type: 'inventory_low_stock' },
    ])
    expect(alerts.map((a) => a.id)).toEqual([1])
  })
})
