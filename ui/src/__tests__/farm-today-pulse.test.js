import { describe, it, expect } from 'vitest'
import {
  buildFarmTodayPulse,
  resolveDevicesCell,
  resolveGrowingCell,
  resolveLightsCell,
  resolveNextWaterCell,
} from '../lib/farmTodayPulse.js'

const zones = [
  { id: 1, name: 'Veg Room', zone_type: 'indoor' },
  { id: 2, name: 'Flower Room', zone_type: 'indoor' },
]

describe('Phase 176 WS1 — farmTodayPulse', () => {
  it('resolveNextWaterCell picks first active program with schedule', () => {
    const cell = resolveNextWaterCell({
      zones,
      programs: [
        { id: 10, target_zone_id: 2, schedule_id: 20, is_active: true },
        { id: 11, target_zone_id: 1, schedule_id: 21, is_active: true },
      ],
      schedules: [
        { id: 20, is_active: true, cron_expression: '0 8 * * *', name: 'Flower feed' },
        { id: 21, is_active: true, cron_expression: '0 3 * * *', name: 'Veg feed' },
      ],
    })
    expect(cell?.id).toBe('next_water')
    expect(cell?.value).toContain('Veg Room')
    expect(cell?.value).toMatch(/3 AM|Every day/)
    expect(cell?.link?.path).toBe('/zones/1')
  })

  it('resolveGrowingCell summarizes active runs and bloom count', () => {
    const cell = resolveGrowingCell({
      cropCycles: [
        { id: 1, is_active: true, current_stage: 'veg' },
        { id: 2, is_active: true, current_stage: 'bloom' },
        { id: 3, is_active: true, current_stage: 'early_flower' },
        { id: 4, is_active: false, current_stage: 'harvested' },
      ],
    })
    expect(cell?.value).toBe('3 runs · 2 in bloom')
    expect(cell?.link).toEqual({ path: '/zones' })
  })

  it('resolveDevicesCell shows online count and queue depth', () => {
    const cell = resolveDevicesCell({
      devices: [
        { id: 1, status: 'online' },
        { id: 2, status: 'offline' },
      ],
      queueDepth: 2,
    })
    expect(cell?.value).toBe('1 of 2 online · queue 2')
    expect(cell?.link).toEqual({ path: '/hardware' })
  })

  it('resolveLightsCell prefers zones-on summary', () => {
    const cell = resolveLightsCell({
      zones,
      schedules: [],
      programs: [],
      actuators: [
        { id: 1, zone_id: 1, actuator_type: 'light', last_command: 'on' },
      ],
      sensors: [],
      readings: {},
      tasks: [],
      alerts: [],
      cropCycles: [],
      fertigationEvents: [],
    })
    expect(cell?.value).toBe('1 zone on')
  })

  it('buildFarmTodayPulse omits empty cells', () => {
    const pulse = buildFarmTodayPulse({ zones: [], programs: [], schedules: [], devices: [] })
    expect(pulse.cells).toEqual([])
  })

  it('buildFarmTodayPulse returns multiple cells when data exists', () => {
    const pulse = buildFarmTodayPulse({
      zones,
      programs: [{ id: 10, target_zone_id: 1, schedule_id: 21, is_active: true }],
      schedules: [{ id: 21, is_active: true, cron_expression: '0 3 * * *' }],
      cropCycles: [{ id: 1, is_active: true, current_stage: 'veg' }],
      devices: [{ id: 1, status: 'online' }],
      actuators: [],
      sensors: [],
      readings: {},
      tasks: [],
      alerts: [],
      fertigationEvents: [],
    })
    expect(pulse.cells.length).toBeGreaterThanOrEqual(3)
    expect(pulse.cells.map((c) => c.id)).toContain('next_water')
    expect(pulse.cells.map((c) => c.id)).toContain('crops')
    expect(pulse.cells.map((c) => c.id)).toContain('edge')
  })
})
