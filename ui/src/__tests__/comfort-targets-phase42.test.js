import { describe, it, expect } from 'vitest'
import {
  comfortBandStatus,
  buildZoneComfortBands,
  summarizeZoneComfortStatus,
  validateComfortBandPayload,
} from '../lib/comfortBand.js'
import { buildFarmComfortCards, filterComfortCardsByZone } from '../lib/farmComfortHub.js'
import { buildDailyCron } from '../lib/cronHumanize.js'
import { ruleSummary } from '../lib/ruleSummary.js'

describe('Phase 42 — comfortBand', () => {
  it('marks missing when no setpoint values', () => {
    expect(comfortBandStatus(null, { value_raw: 50 })).toBe('missing')
  })

  it('marks out_of_range below min', () => {
    expect(comfortBandStatus(
      { min_value: 40, ideal_value: 50, max_value: 60 },
      { value_raw: 30 },
    )).toBe('out_of_range')
  })

  it('marks ok when reading inside band', () => {
    expect(comfortBandStatus(
      { min_value: 40, ideal_value: 50, max_value: 60 },
      { value_raw: 55 },
    )).toBe('ok')
  })

  it('validates band payload ordering', () => {
    expect(validateComfortBandPayload({
      sensor_type: 'humidity',
      min_value: 60,
      ideal_value: 50,
      max_value: 40,
    })).toContain('Too low')
  })
})

describe('Phase 42 — farmComfortHub', () => {
  const zones = [{ id: 1, name: 'Flower Room' }]
  const sensors = [
    { id: 10, zone_id: 1, sensor_type: 'humidity' },
    { id: 11, zone_id: 1, sensor_type: 'temperature' },
  ]

  it('builds cards with band summary', () => {
    const cards = buildFarmComfortCards({
      zones,
      sensors,
      setpoints: [{ id: 1, zone_id: 1, sensor_type: 'humidity', min_value: 40, max_value: 60 }],
      readings: { 10: { value_raw: 55 } },
    })
    expect(cards).toHaveLength(1)
    expect(cards[0].status).toBe('missing')
    expect(cards[0].summaryLine).toContain('Humidity')
    expect(cards[0].summaryLine).toContain('Temperature')
  })

  it('filters cards by zone_id', () => {
    const cards = buildFarmComfortCards({
      zones: [...zones, { id: 2, name: 'Veg' }],
      sensors: [...sensors, { id: 20, zone_id: 2, sensor_type: 'humidity' }],
      setpoints: [],
      readings: {},
    })
    expect(filterComfortCardsByZone(cards, 1)).toHaveLength(1)
  })
})

describe('Phase 42 — cronHumanize buildDailyCron', () => {
  it('builds daily cron at hour and minute', () => {
    expect(buildDailyCron(6, 30)).toBe('30 6 * * *')
  })
})

describe('Phase 42 — ruleSummary', () => {
  it('renders a one-line farmer summary', () => {
    const text = ruleSummary(
      {
        name: 'Shade high temp',
        condition_logic: 'ALL',
        conditions_jsonb: { predicates: [{ sensor_id: 1, op: 'gt', value: 28 }] },
      },
      {
        sensors: [{ id: 1, name: 'Air temp' }],
        actuators: [{ id: 2, name: 'Shade motor' }],
        actions: [{ action_type: 'control_actuator', target_actuator_id: 2, action_command: 'deploy' }],
      },
    )
    expect(text).toContain('Air temp')
    expect(text).toContain('Shade motor')
  })
})
