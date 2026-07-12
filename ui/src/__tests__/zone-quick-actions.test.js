import { describe, it, expect } from 'vitest'
import {
  resolveWaterNowAction,
  sortZonesForStack,
  zoneStackSortKey,
  lightActuatorsForZone,
  zoneTasksForSheet,
  zoneAlertsForSheet,
} from '../lib/zoneQuickActions.js'
import { computeZoneVisualStatus } from '../lib/farmVisualStatus.js'
import { buildZoneQuickStarters } from '../lib/guardianStarters.js'

describe('Phase 167 — zone quick actions lib', () => {
  const herbZone = { id: 5, name: 'Herb & Greens Room', zone_type: 'indoor' }
  const flowerZone = { id: 2, name: 'Flower Room', zone_type: 'indoor' }

  it('resolves water now via active program', () => {
    const action = resolveWaterNowAction({
      zone: herbZone,
      programs: [{
        id: 12,
        name: 'Herb Room Gravity Drip',
        target_zone_id: 5,
        is_active: true,
        run_duration_seconds: 180,
      }],
      actuators: [{ id: 7, zone_id: 5, actuator_type: 'drip', name: 'Herb Room Gravity Drip Valve' }],
    })
    expect(action.mode).toBe('program')
    expect(action.confirm).toContain('Herb Room Gravity Drip')
  })

  it('falls back to pulse when no program but water actuator exists', () => {
    const action = resolveWaterNowAction({
      zone: { id: 9, name: 'Bench' },
      programs: [],
      actuators: [{ id: 3, zone_id: 9, actuator_type: 'pump', name: 'Bench pump' }],
    })
    expect(action.mode).toBe('pulse')
  })

  it('suggests setup when no program or actuator', () => {
    const action = resolveWaterNowAction({
      zone: { id: 4, name: 'Empty' },
      programs: [],
      actuators: [],
    })
    expect(action.mode).toBe('setup')
    expect(action.link.path).toContain('/zones/4')
  })

  it('orders stack with alert zones first', () => {
    const zones = [
      { id: 1, name: 'Healthy' },
      { id: 2, name: 'Flower Room' },
      { id: 3, name: 'Empty' },
    ]
    const statusFor = (z) => {
      if (z.id === 2) return { health: 'warn', attention: [{ kind: 'alert' }], plants: { state: 'growing' } }
      if (z.id === 3) return { health: 'unconfigured', plants: { state: 'empty' } }
      return { health: 'ok', plants: { state: 'growing' } }
    }
    const sorted = sortZonesForStack(zones, statusFor)
    expect(sorted[0].name).toBe('Flower Room')
    expect(sorted[sorted.length - 1].name).toBe('Empty')
    expect(zoneStackSortKey({ health: 'alert' })).toBeLessThan(zoneStackSortKey({ health: 'ok' }))
  })

  it('finds light actuators for zone', () => {
    const lights = lightActuatorsForZone([
      { id: 1, zone_id: 2, actuator_type: 'light', name: 'Flower LED' },
      { id: 2, zone_id: 3, actuator_type: 'pump', name: 'Pump' },
    ], 2)
    expect(lights).toHaveLength(1)
  })

  it('builds zone Guardian starters from tile status', () => {
    const status = computeZoneVisualStatus({
      zone: flowerZone,
      sensors: [{ id: 10, zone_id: 2, sensor_type: 'humidity', alert_threshold_high: 65 }],
      readings: { 10: { value_raw: 72.4, is_valid: true, reading_time: new Date().toISOString() } },
      alerts: [{ is_read: false, is_acknowledged: false, severity: 'warning', title: 'Humidity high' }],
      tasks: [],
      schedules: [],
      programs: [],
      actuators: [],
      cropCycles: [],
      fertigationEvents: [],
    })
    const starters = buildZoneQuickStarters({ zone: flowerZone, status, farmId: 1 })
    expect(starters.length).toBeGreaterThan(0)
    expect(starters[0].message).toMatch(/Flower Room/)
  })
})
