import { describe, it, expect } from 'vitest'
import {
  computeZoneTodaySnapshot,
  filterZoneAlerts,
  ruleAppliesToZone,
} from '../lib/zoneGrowSummary.js'
import { humanizeCron } from '../lib/cronHumanize.js'

describe('Phase 40 WS1 — zone grow summary', () => {
  it('humanizes daily cron for farmers', () => {
    expect(humanizeCron('0 6 * * *')).toBe('Every day at 6 AM')
    expect(humanizeCron('0 18 * * *')).toBe('Every day at 6 PM')
  })

  it('filters unread alerts for zone sensors', () => {
    const sensors = [{ id: 10, sensor_type: 'humidity' }]
    const alerts = [
      {
        id: 1,
        is_read: false,
        is_acknowledged: false,
        triggering_event_source_type: 'sensor',
        triggering_event_source_id: 10,
        subject_rendered: 'Humidity high',
      },
      {
        id: 2,
        is_read: false,
        is_acknowledged: false,
        triggering_event_source_type: 'input_batch',
        triggering_event_source_id: 99,
        subject_rendered: 'Farm-wide stock',
      },
    ]
    expect(filterZoneAlerts(alerts, sensors, 'Flower Room')).toHaveLength(1)
    expect(filterZoneAlerts(alerts, sensors, 'Flower Room')[0].id).toBe(1)
  })

  it('counts zone rules by zone_id in trigger_configuration', () => {
    const zoneId = 3
    const rules = [
      { id: 1, is_active: true, trigger_configuration: { zone_id: 3 }, conditions_jsonb: [] },
      { id: 2, is_active: true, trigger_configuration: { zone_id: 99 }, conditions_jsonb: [] },
    ]
    expect(ruleAppliesToZone(rules[0], zoneId, 'Flower Room', new Set())).toBe(true)
    expect(ruleAppliesToZone(rules[1], zoneId, 'Flower Room', new Set())).toBe(false)
  })

  it('builds Today chips with schedule, alerts, and queue', () => {
    const snap = computeZoneTodaySnapshot({
      zone: { id: 3, name: 'Flower Room' },
      sensors: [{ id: 10, sensor_type: 'humidity' }],
      devices: [{ id: 1, status: 'online' }],
      alerts: [
        {
          id: 5,
          is_read: false,
          is_acknowledged: false,
          triggering_event_source_type: 'sensor',
          triggering_event_source_id: 10,
          subject_rendered: 'Humidity high — Flower Room',
        },
      ],
      rules: [{ id: 1, is_active: true, trigger_configuration: { zone_id: 3 }, conditions_jsonb: [] }],
      schedules: [
        {
          id: 20,
          name: 'Water Early Flower Daily',
          cron_expression: '0 8 * * *',
          is_active: true,
          description: 'Zone: Flower Room.',
        },
      ],
      activeProgram: { schedule_id: 20 },
      lightingPrograms: [],
      setpoints: [],
      queueDepth: 2,
      zoneTasks: [],
    })

    const ids = snap.chips.map((c) => c.id)
    expect(ids).toContain('next-schedule')
    expect(ids).toContain('open-alerts')
    expect(ids).toContain('queue')
    const alertChip = snap.chips.find((c) => c.id === 'open-alerts')
    expect(alertChip.value).toBe('1')
    expect(snap.chips.find((c) => c.id === 'queue').value).toBe('2')
  })
})
