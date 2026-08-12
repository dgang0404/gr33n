import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { setActivePinia, createPinia } from 'pinia'
import ActuatorCard from '../components/ActuatorCard.vue'

vi.mock('../api', () => ({
  default: {
    get: vi.fn().mockResolvedValue({ data: [] }),
    post: vi.fn(),
    interceptors: { request: { use: vi.fn() }, response: { use: vi.fn() } },
  },
}))

describe('Phase 117 — ActuatorCard sync badge', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.useRealTimers()
  })

  it('shows config sync badge for platform-sync devices', () => {
    const wrapper = mount(ActuatorCard, {
      props: {
        device: {
          id: 1,
          name: 'Demo Relay',
          device_type: 'irrigation',
          zone_id: 2,
          device_uid: 'pi-demo-01',
          config_version: 3,
          config: { last_config_fetch_at: new Date().toISOString() },
          status: 'online',
        },
      },
      global: {
        stubs: {
          DeviceApiKeyPanel: true,
          DeviceIPHistoryPanel: true,
          DeviceCommandQueue: true,
          ActuatorPulseControl: true,
        },
      },
    })
    const badge = wrapper.find('[data-test="device-config-sync-badge"]')
    expect(badge.exists()).toBe(true)
    expect(badge.text()).toMatch(/Config synced|Never fetched/)
  })

  it('hides sync badge when device is not on platform sync', () => {
    const wrapper = mount(ActuatorCard, {
      props: {
        device: {
          id: 2,
          name: 'Local YAML Pi',
          device_type: 'fan',
          zone_id: 1,
          config_version: 0,
        },
      },
      global: {
        stubs: {
          DeviceApiKeyPanel: true,
          DeviceIPHistoryPanel: true,
          DeviceCommandQueue: true,
          ActuatorPulseControl: true,
        },
      },
    })
    expect(wrapper.find('[data-test="device-config-sync-badge"]').exists()).toBe(false)
  })
})

describe('Phase 214 — ActuatorCard IP history', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.useRealTimers()
  })

  it('shows current IP and toggle when device has an ip_address', async () => {
    const wrapper = mount(ActuatorCard, {
      props: {
        device: {
          id: 3,
          name: 'Veg Room Pi',
          device_type: 'raspberry_pi_edge',
          zone_id: 1,
          ip_address: '192.168.1.246',
          status: 'online',
        },
      },
      global: {
        stubs: {
          DeviceApiKeyPanel: true,
          DeviceIPHistoryPanel: true,
          DeviceCommandQueue: true,
          ActuatorPulseControl: true,
        },
      },
    })
    expect(wrapper.find('[data-test="device-current-ip"]').text()).toBe('192.168.1.246')
    const toggle = wrapper.find('[data-test="device-ip-history-toggle"]')
    expect(toggle.exists()).toBe(true)
    expect(wrapper.findComponent({ name: 'DeviceIPHistoryPanel' }).exists()).toBe(false)

    await toggle.trigger('click')
    expect(wrapper.find('[data-test="device-ip-history-toggle"]').text()).toBe('Hide IP history')
  })

  it('hides IP line and toggle when device has no ip_address', () => {
    const wrapper = mount(ActuatorCard, {
      props: {
        device: {
          id: 4,
          name: 'Local YAML Pi',
          device_type: 'fan',
          zone_id: 1,
        },
      },
      global: {
        stubs: {
          DeviceApiKeyPanel: true,
          DeviceIPHistoryPanel: true,
          DeviceCommandQueue: true,
          ActuatorPulseControl: true,
        },
      },
    })
    expect(wrapper.find('[data-test="device-current-ip"]').exists()).toBe(false)
    expect(wrapper.find('[data-test="device-ip-history-toggle"]').exists()).toBe(false)
  })
})
