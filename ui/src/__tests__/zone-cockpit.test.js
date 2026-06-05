import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { setActivePinia, createPinia } from 'pinia'
import ZoneTodayStrip from '../components/ZoneTodayStrip.vue'
import ZoneAdvancedHint from '../components/ZoneAdvancedHint.vue'
import GuardianStarterChips from '../components/GuardianStarterChips.vue'
import { computeZoneTodaySnapshot } from '../lib/zoneGrowSummary.js'
import { useCapabilitiesStore } from '../stores/capabilities'

vi.mock('../api', () => ({
  default: {
    get: vi.fn((url) => {
      if (url === '/capabilities') return Promise.resolve({ data: { ai_enabled: true } })
      return Promise.resolve({ data: {} })
    }),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
    interceptors: { request: { use: vi.fn() }, response: { use: vi.fn() } },
  },
}))

describe('Phase 40 WS8 — zone cockpit integration', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('Today strip renders chips from zone snapshot', () => {
    const chips = computeZoneTodaySnapshot({
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
          subject_rendered: 'Humidity high',
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
      zoneTasks: [{ id: 1, title: 'Check runoff', due_date: '2026-06-04' }],
    }).chips

    const wrapper = mount(ZoneTodayStrip, { props: { chips } })

    expect(wrapper.find('[data-test="zone-today-strip"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="zone-today-chip-next-schedule"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="zone-today-chip-open-alerts"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="zone-today-chip-queue"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('Today in this room')
    expect(wrapper.text()).toContain('Open alerts')
    expect(wrapper.text()).toContain('Queued commands')
  })

  it('Advanced hint points power users to farm-wide pages', () => {
    const wrapper = mount(ZoneAdvancedHint, {
      global: { stubs: { RouterLink: true } },
    })

    expect(wrapper.find('[data-test="zone-advanced-hint"]').exists()).toBe(true)
    expect(wrapper.text()).toMatch(/Power settings|Advanced/i)
    expect(wrapper.text()).not.toContain('setpoint')
  })

  it('Guardian starter chips render zone-context prompts', async () => {
    await useCapabilitiesStore().fetch()

    const wrapper = mount(GuardianStarterChips, {
      props: {
        starters: [
          {
            id: 'comfort',
            label: 'Why is humidity off?',
            message: 'Explain humidity vs target in Flower Room',
          },
        ],
      },
    })
    await flushPromises()

    expect(wrapper.find('[data-test="guardian-starter-chips"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="guardian-starter-comfort"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('Why is humidity off?')
  })
})
