import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { setActivePinia, createPinia } from 'pinia'
import ZoneWaterGrowStory from '../components/ZoneWaterGrowStory.vue'

vi.mock('../api', () => ({
  default: {
    get: vi.fn(() => Promise.resolve({ data: { queue_depth: 0, mix_required: false } })),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
    interceptors: { request: { use: vi.fn() }, response: { use: vi.fn() } },
  },
}))

describe('Phase 47 WS2 — zone Water primary surface', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })
  const baseProps = {
    zoneId: 3,
    farmId: 1,
    programs: [{ id: 10, name: 'Flower FFJ Program' }],
    schedules: [{
      id: 20,
      name: 'Water Early Flower Daily',
      cron_expression: '0 8 * * *',
      is_active: true,
    }],
    fertigationEvents: [{
      id: 2,
      zone_id: 3,
      applied_at: '2026-06-04T08:00:00Z',
      volume_applied_liters: 0.9,
      ec_after_mscm: 2.1,
      program_id: 10,
    }],
    actuators: [],
    ecTargets: [{ id: 5, ec_min_mscm: 1.1, ec_max_mscm: 1.3 }],
    reservoirs: [{ id: 1, name: 'Flower tank', status: 'ready' }],
  }

  it('shows feeding plan card and status line without fertigation as primary CTA', () => {
    const wrapper = mount(ZoneWaterGrowStory, {
      props: {
        ...baseProps,
        activeProgram: {
          id: 10,
          name: 'Flower FFJ Program',
          schedule_id: 20,
          total_volume_liters: 0.3,
          ec_target_id: 5,
          reservoir_id: 1,
        },
      },
      global: {
        stubs: {
          RouterLink: { props: ['to'], template: '<a :data-to="JSON.stringify(to)"><slot /></a>' },
          ActuatorPulseControl: true,
          EmptyStateHint: true,
        },
      },
    })

    expect(wrapper.find('[data-test="feeding-status-line"]').text()).toContain('Next feed:')
    expect(wrapper.find('[data-test="feeding-plan-card"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="feeding-advanced-link"]').text()).toContain('Advanced feeding')
    expect(wrapper.find('[data-test="grow-story-run-now"]').text()).toContain('Run feed now')
    const historyTo = JSON.parse(wrapper.find('[data-test="feeding-history-link"]').attributes('data-to'))
    expect(historyTo.path).toBe('/feeding')
    expect(historyTo.query.zone_id).toBe('3')
  })

  it('shows water-only badge and hides preview mix for irrigation_only', () => {
    const wrapper = mount(ZoneWaterGrowStory, {
      props: {
        ...baseProps,
        activeProgram: {
          id: 11,
          name: 'Plain irrigation',
          irrigation_only: true,
          total_volume_liters: 1,
        },
      },
      global: {
        stubs: {
          RouterLink: { props: ['to'], template: '<a :data-to="JSON.stringify(to)"><slot /></a>' },
          ActuatorPulseControl: true,
          EmptyStateHint: true,
        },
      },
    })

    expect(wrapper.find('[data-test="feeding-water-only-badge"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="grow-story-preview-mix"]').exists()).toBe(false)
  })
})
