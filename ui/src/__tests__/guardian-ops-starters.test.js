import { describe, it, expect } from 'vitest'
import {
  buildSuppliesHubStarters,
  buildFeedingAdminStarters,
  buildMoneyHubStarters,
  buildDashboardOpsStarters,
  lowStockInputFromAlert,
} from '../lib/guardianStarters.js'

describe('Phase 43 WS8 — operations Guardian starters', () => {
  const zones = [{ id: 3, name: 'Flower Room' }]

  it('supplies hub shows low-stock chip when batches are below threshold', () => {
    const starters = buildSuppliesHubStarters({
      lowStockRows: [{ inputName: 'OHN', remaining: 1, threshold: 3 }],
      recipes: [{ id: 1, name: 'Bloom mix' }],
      zones,
    })
    const low = starters.find((s) => s.id === 'whats-running-low')
    expect(low).toBeTruthy()
    expect(low.message).toContain('low-stock threshold')
    expect(low.message).not.toMatch(/inventory/i)
  })

  it('supplies hub offers refill task from low-stock alert', () => {
    const starters = buildSuppliesHubStarters({
      lowStockAlerts: [{
        id: 12,
        subject_rendered: 'Inventory low: OHN at 1.00 (threshold 3.00)',
        triggering_event_source_type: 'inventory_low_stock',
      }],
      zones,
    })
    const refill = starters.find((s) => s.id === 'refill-from-alert')
    expect(refill?.message).toContain('alert #12')
    expect(refill?.message).toContain('OHN')
    expect(refill?.contextRef?.alert_id).toBe(12)
  })

  it('feeding admin offers schedule chip when program is active', () => {
    const starters = buildFeedingAdminStarters({
      zones,
      zoneContextId: 3,
      programs: [{ id: 1, name: 'Bloom', target_zone_id: 3, is_active: true }],
    })
    expect(starters.some((s) => s.id === 'next-feed-schedule')).toBe(true)
    expect(starters[0].message).toContain('plain language')
  })

  it('money hub offers month spend summary chips', () => {
    const starters = buildMoneyHubStarters()
    expect(starters.length).toBeGreaterThanOrEqual(2)
    expect(starters.some((s) => s.message.includes('by category'))).toBe(true)
    expect(starters.some((s) => s.message.includes('no accounting jargon'))).toBe(true)
    expect(starters[0].contextRef.path).toBe('/money')
  })

  it('supplies hub offers restock-first when low stock', () => {
    const starters = buildSuppliesHubStarters({
      lowStockRows: [{ inputName: 'JMS', remaining: 0.5, threshold: 2 }],
      zones,
    })
    expect(starters.some((s) => s.id === 'restock-first')).toBe(true)
  })

  it('dashboard ops starters appear only when low stock exists', () => {
    expect(buildDashboardOpsStarters({ lowStockCount: 0 })).toEqual([])
    const starters = buildDashboardOpsStarters({
      lowStockCount: 2,
      lowStockAlerts: [{ id: 9, subject_rendered: 'Inventory low: JMS at 0.5 (threshold 2)' }],
    })
    expect(starters.length).toBeLessThanOrEqual(2)
    expect(starters[0].id).toBe('whats-running-low')
  })

  it('parses input name from low-stock alert subject', () => {
    expect(lowStockInputFromAlert({ subject_rendered: 'Inventory low: OHN at 1.00 (threshold 3.00)' })).toBe('OHN')
  })
})
