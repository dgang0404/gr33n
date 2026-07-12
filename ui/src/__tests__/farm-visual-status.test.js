import { describe, it, expect } from 'vitest'
import {
  classifySensorHardwareState,
  computeZoneVisualStatus,
  defaultLayoutForZone,
  formatZoneTypeLabel,
  resolveZoneLayout,
  DEFAULT_ZONE_LAYOUTS_BY_NAME,
} from '../lib/farmVisualStatus.js'

const vegZone = { id: 1, name: 'Veg Room', zone_type: 'indoor' }
const flowerZone = { id: 2, name: 'Flower Room', zone_type: 'indoor' }
const herbZone = { id: 5, name: 'Herb & Greens Room', zone_type: 'indoor' }
const pepperZone = { id: 6, name: 'Outdoor Pepper Bed', zone_type: 'outdoor' }

describe('Phase 166 WS1 — farmVisualStatus', () => {
  it('classifies three hardware states', () => {
    expect(classifySensorHardwareState(
      { id: 1, sensor_type: 'humidity', alert_threshold_low: 40, alert_threshold_high: 65 },
      { value_raw: 55, is_valid: true, reading_time: new Date().toISOString() },
    )).toBe('healthy')

    expect(classifySensorHardwareState(
      { id: 2, sensor_type: 'humidity', alert_threshold_low: 40, alert_threshold_high: 65 },
      { value_raw: 72.4, is_valid: true, reading_time: new Date().toISOString() },
    )).toBe('attention')

    expect(classifySensorHardwareState(
      { id: 3, sensor_type: 'temperature' },
      null,
    )).toBe('not_set_up')
  })

  it('rolls up Veg Room as healthy with growing plants', () => {
    const status = computeZoneVisualStatus({
      zone: vegZone,
      sensors: [
        { id: 10, zone_id: 1, sensor_type: 'temperature', alert_threshold_low: 18, alert_threshold_high: 28 },
        { id: 11, zone_id: 1, sensor_type: 'humidity', alert_threshold_low: 40, alert_threshold_high: 70 },
      ],
      readings: {
        10: { sensor_id: 10, value_raw: 24, is_valid: true, reading_time: new Date().toISOString() },
        11: { sensor_id: 11, value_raw: 58, is_valid: true, reading_time: new Date().toISOString() },
      },
      cropCycles: [{
        zone_id: 1,
        is_active: true,
        batch_label: 'Anastasia Green',
        current_stage: 'late_veg',
      }],
      alerts: [],
      tasks: [],
      schedules: [],
      programs: [],
      actuators: [],
      fertigationEvents: [],
    })
    expect(status.health).toBe('ok')
    expect(status.plants.state).toBe('growing')
    expect(status.plants.stage).toBe('Veg stage')
    expect(status.sensors.state).toBe('healthy')
  })

  it('shows Flower Room humidity attention', () => {
    const status = computeZoneVisualStatus({
      zone: flowerZone,
      sensors: [
        { id: 20, zone_id: 2, sensor_type: 'humidity', alert_threshold_low: 40, alert_threshold_high: 65 },
      ],
      readings: {
        20: { sensor_id: 20, value_raw: 72.4, is_valid: true, reading_time: new Date().toISOString() },
      },
      alerts: [{
        is_read: false,
        is_acknowledged: false,
        severity: 'warning',
        title: 'Humidity high',
        triggering_event_source_type: 'sensor',
        triggering_event_source_id: 20,
      }],
      cropCycles: [{ zone_id: 2, is_active: true, batch_label: 'Zembla White', current_stage: 'early_flower' }],
      tasks: [],
      schedules: [],
      programs: [],
      actuators: [],
      fertigationEvents: [],
    })
    expect(status.health).toBe('warn')
    expect(status.sensors.summary).toBe('Humidity high')
  })

  it('shows unwired bed as not set up yet', () => {
    const status = computeZoneVisualStatus({
      zone: pepperZone,
      sensors: [{ id: 30, zone_id: 6, sensor_type: 'soil_moisture' }],
      readings: {},
      alerts: [],
      tasks: [],
      schedules: [],
      programs: [],
      actuators: [],
      cropCycles: [],
      fertigationEvents: [],
    })
    expect(status.sensors.summary).toBe('Not set up yet')
    expect(status.health).toBe('unconfigured')
  })

  it('shows gravity drip water line for Herb & Greens', () => {
    const status = computeZoneVisualStatus({
      zone: herbZone,
      sensors: [],
      readings: {},
      programs: [{
        id: 99,
        name: 'Herb Room Gravity Drip',
        target_zone_id: 5,
        is_active: true,
        irrigation_only: true,
        schedule_id: 40,
      }],
      schedules: [{
        id: 40,
        name: 'Water Herbs Gravity Drip Daily',
        cron_expression: '0 7 * * *',
        is_active: true,
      }],
      actuators: [{
        id: 7,
        zone_id: 5,
        name: 'Herb Room Gravity Drip Valve',
        actuator_type: 'drip',
      }],
      alerts: [],
      tasks: [],
      cropCycles: [],
      fertigationEvents: [],
    })
    expect(status.water.kind).toBe('gravity_drip')
    expect(status.water.label).toContain('Gravity drip')
    expect(status.water.nextRun).toBeTruthy()
  })

  it('ships default layouts for demo farm zone names', () => {
    expect(DEFAULT_ZONE_LAYOUTS_BY_NAME['Veg Room']).toBeTruthy()
    expect(defaultLayoutForZone({ name: 'Unknown' }, 2).x).toBeGreaterThan(0)
    const layout = resolveZoneLayout(vegZone, () => null, 0)
    expect(layout.x).toBe(DEFAULT_ZONE_LAYOUTS_BY_NAME['Veg Room'].x)
  })

  it('formats zone_type for farmer-facing tiles', () => {
    expect(formatZoneTypeLabel('greenhouse')).toBe('Greenhouse')
    expect(formatZoneTypeLabel('outdoor_bed')).toBe('Outdoor bed')
    expect(formatZoneTypeLabel('')).toBe('Grow area')
  })

  it('uses farmer-friendly empty plant copy', () => {
    const status = computeZoneVisualStatus({
      zone: pepperZone,
      sensors: [],
      readings: {},
      alerts: [],
      tasks: [],
      schedules: [],
      programs: [],
      actuators: [],
      cropCycles: [],
      fertigationEvents: [],
    })
    expect(status.plants.label).toBe('Empty — ready to plant')
  })
})
