import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { setActivePinia, createPinia } from 'pinia'

vi.mock('../api', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
    interceptors: { request: { use: vi.fn() }, response: { use: vi.fn() } },
  },
}))

import api from '../api'
import ZoneComfortTargets from '../components/ZoneComfortTargets.vue'

describe('Phase 40 WS2 — ZoneComfortTargets', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('renders comfort target editor for zone sensors', () => {
    const wrapper = mount(ZoneComfortTargets, {
      props: {
        need: 'air',
        zoneId: 3,
        farmId: 1,
        sensors: [{ id: 10, sensor_type: 'humidity', zone_id: 3 }],
        setpoints: [
          {
            id: 5,
            zone_id: 3,
            sensor_type: 'humidity',
            min_value: 40,
            ideal_value: 50,
            max_value: 60,
          },
        ],
      },
      global: { stubs: { RouterLink: true } },
    })

    expect(wrapper.find('[data-test="zone-comfort-targets"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="comfort-target-humidity"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('Humidity')
    expect(wrapper.text()).not.toContain('Setpoint')
  })

  it('POSTs new comfort target on save', async () => {
    api.post.mockResolvedValue({
      data: { id: 99, zone_id: 3, sensor_type: 'humidity', min_value: 35, ideal_value: 45, max_value: 55 },
    })

    const wrapper = mount(ZoneComfortTargets, {
      props: {
        need: 'air',
        zoneId: 3,
        farmId: 1,
        sensors: [{ id: 10, sensor_type: 'humidity', zone_id: 3 }],
        setpoints: [],
      },
      global: { stubs: { RouterLink: true } },
    })

    await wrapper.find('[data-test="add-comfort-target-humidity"]').trigger('click')
    await flushPromises()

    const form = wrapper.find('[data-test="comfort-target-humidity"]')
    const inputs = form.findAll('input[type="number"]')
    await inputs[0].setValue(35)
    await inputs[1].setValue(45)
    await inputs[2].setValue(55)
    await form.trigger('submit')
    await flushPromises()

    expect(api.post).toHaveBeenCalledWith('/farms/1/setpoints', expect.objectContaining({
      zone_id: 3,
      sensor_type: 'humidity',
      min_value: 35,
      ideal_value: 45,
      max_value: 55,
    }))
    expect(wrapper.emitted('updated')).toBeTruthy()
  })
})
