import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import FarmZoneStack from '../components/FarmZoneStack.vue'
import { sortZonesForStack } from '../lib/zoneQuickActions.js'

vi.mock('../api', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    patch: vi.fn(),
    interceptors: { request: { use: vi.fn() }, response: { use: vi.fn() } },
  },
}))

describe('Phase 167 WS1 — FarmZoneStack', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('renders stacked zone cards on mobile layout', () => {
    const wrapper = mount(FarmZoneStack, {
      props: {
        zones: [
          { id: 1, name: 'Veg Room', zone_type: 'indoor' },
          { id: 2, name: 'Flower Room', zone_type: 'indoor' },
        ],
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
      global: { stubs: { FarmCanvasZoneTile: true, RouterLink: true } },
    })
    expect(wrapper.find('[data-test="farm-zone-stack"]').exists()).toBe(true)
    expect(wrapper.findAll('[data-test^="farm-zone-stack-card-"]')).toHaveLength(2)
  })

  it('emits select-zone when card tapped', async () => {
    const wrapper = mount(FarmZoneStack, {
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
      global: { stubs: { FarmCanvasZoneTile: true, RouterLink: true } },
    })
    await wrapper.find('[data-test="farm-zone-stack-card-1"]').trigger('click')
    expect(wrapper.emitted('select-zone')?.[0]?.[0]?.name).toBe('Veg Room')
  })

  it('sorts attention zones before healthy', () => {
    const zones = [
      { id: 1, name: 'A' },
      { id: 2, name: 'B' },
    ]
    const sorted = sortZonesForStack(zones, (z) => (
      z.id === 2
        ? { health: 'warn', attention: [{}], plants: { state: 'growing' } }
        : { health: 'ok', plants: { state: 'growing' } }
    ))
    expect(sorted[0].id).toBe(2)
  })
})
