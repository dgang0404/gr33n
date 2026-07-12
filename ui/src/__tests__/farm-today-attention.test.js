import { describe, it, expect, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import FarmTodayAttentionStrip from '../components/FarmTodayAttentionStrip.vue'
import {
  listAttentionZones,
  zoneAttentionSummary,
  zoneNeedsAttention,
} from '../lib/zoneQuickActions.js'
import { computeZoneVisualStatus } from '../lib/farmVisualStatus.js'
import { buildTodayAttentionStarters } from '../lib/guardianStarters.js'

describe('Phase 169 — attention helpers', () => {
  const flowerZone = { id: 2, name: 'Flower Room', zone_type: 'indoor' }
  const vegZone = { id: 1, name: 'Veg Room', zone_type: 'indoor' }

  it('detects attention zones from visual status', () => {
    const status = computeZoneVisualStatus({
      zone: flowerZone,
      sensors: [{
        id: 10,
        zone_id: 2,
        sensor_type: 'humidity',
        alert_threshold_low: 40,
        alert_threshold_high: 65,
      }],
      readings: {
        10: {
          value_raw: 72.4,
          is_valid: true,
          reading_time: new Date().toISOString(),
        },
      },
      alerts: [{ id: 1, subject_rendered: 'Humidity high — Flower Room', is_read: false }],
      tasks: [],
      schedules: [],
      programs: [],
      actuators: [],
      cropCycles: [],
      fertigationEvents: [],
    })
    expect(zoneNeedsAttention(status)).toBe(true)
    expect(zoneAttentionSummary(status)).toContain('Humidity')
    const flagged = listAttentionZones([vegZone, flowerZone], (z) =>
      z.id === 2 ? status : computeZoneVisualStatus({
        zone: z,
        sensors: [],
        readings: {},
        alerts: [],
        tasks: [],
        schedules: [],
        programs: [],
        actuators: [],
        cropCycles: [],
        fertigationEvents: [],
      }),
    )
    expect(flagged).toHaveLength(1)
    expect(flagged[0].zone.name).toBe('Flower Room')
  })

  it('builds Guardian starters for a single flagged zone', () => {
    const status = { health: 'warn', sensors: { summary: 'Humidity high' }, attention: [{ label: 'Humidity high' }] }
    const starters = buildTodayAttentionStarters({
      zones: [flowerZone],
      getStatus: () => status,
      farmId: 1,
    })
    expect(starters).toHaveLength(1)
    expect(starters[0].label).toContain('Flower Room')
    expect(starters[0].message).toContain('Humidity')
  })

  it('builds triage starters when multiple zones are flagged', () => {
    const starters = buildTodayAttentionStarters({
      zones: [vegZone, flowerZone],
      getStatus: (z) => ({
        health: 'warn',
        attention: [{ label: `${z.name} issue` }],
        sensors: { summary: `${z.name} issue` },
      }),
      farmName: 'Demo Farm',
    })
    expect(starters.some((s) => s.id === 'attention-triage')).toBe(true)
    expect(starters.length).toBeLessThanOrEqual(3)
  })
})

describe('Phase 169 WS1 — FarmTodayAttentionStrip', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('renders chips for flagged zones and emits select-zone', async () => {
    const wrapper = mount(FarmTodayAttentionStrip, {
      props: {
        zones: [{ id: 2, name: 'Flower Room', zone_type: 'indoor' }],
        sensors: [{
          id: 10,
          zone_id: 2,
          sensor_type: 'humidity',
          alert_threshold_low: 40,
          alert_threshold_high: 65,
        }],
        readings: {
          10: {
            value_raw: 72.4,
            is_valid: true,
            reading_time: new Date().toISOString(),
          },
        },
        actuators: [],
        tasks: [],
        alerts: [{ id: 1, subject_rendered: 'Humidity high — Flower Room' }],
        schedules: [],
        programs: [],
        cropCycles: [],
        fertigationEvents: [],
      },
    })
    expect(wrapper.find('[data-test="farm-today-attention"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="farm-attention-chip-2"]').exists()).toBe(true)
    await wrapper.find('[data-test="farm-attention-chip-2"]').trigger('click')
    expect(wrapper.emitted('select-zone')).toHaveLength(1)
  })

  it('hides when no zones need attention', () => {
    const wrapper = mount(FarmTodayAttentionStrip, {
      props: {
        zones: [{ id: 1, name: 'Veg Room', zone_type: 'indoor' }],
        sensors: [],
        readings: {},
        actuators: [],
        tasks: [],
        alerts: [],
        schedules: [],
        programs: [],
        cropCycles: [],
        fertigationEvents: [],
      },
    })
    expect(wrapper.find('[data-test="farm-today-attention"]').exists()).toBe(false)
  })
})
