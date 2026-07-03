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
          DeviceCommandQueue: true,
          ActuatorPulseControl: true,
        },
      },
    })
    expect(wrapper.find('[data-test="device-config-sync-badge"]').exists()).toBe(false)
  })
})
