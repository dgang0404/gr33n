import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { setActivePinia, createPinia } from 'pinia'

vi.mock('../api', () => ({
  default: {
    patch: vi.fn(),
    interceptors: { request: { use: vi.fn() }, response: { use: vi.fn() } },
  },
}))

import api from '../api'
import ZoneAutomationPanel from '../components/ZoneAutomationPanel.vue'
import { useFarmStore } from '../stores/farm'

describe('Phase 40 WS3 — ZoneAutomationPanel', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('toggles rule active without leaving zone', async () => {
    api.patch.mockResolvedValue({
      data: { id: 7, is_active: false, name: 'AUTO Light ON 12/12 Flower' },
    })
    const store = useFarmStore()
    store.updateAutomationRuleActive = vi.fn().mockResolvedValue({
      id: 7,
      is_active: false,
    })

    const wrapper = mount(ZoneAutomationPanel, {
      props: {
        need: 'light',
        zoneId: 3,
        zoneName: 'Flower Room',
        sensors: [],
        rules: [
          {
            id: 7,
            name: 'AUTO Light ON 12/12 Flower',
            is_active: true,
            trigger_configuration: { target_zone: 'Flower Room', action: 'actuator_on' },
            conditions_jsonb: [],
          },
        ],
        schedules: [],
        lightingPrograms: [],
      },
      global: { stubs: { RouterLink: true } },
    })

    expect(wrapper.find('[data-test="zone-rule-7"]').exists()).toBe(true)
    await wrapper.find('[data-test="zone-rule-toggle-7"]').trigger('click')
    await flushPromises()

    expect(store.updateAutomationRuleActive).toHaveBeenCalledWith(7, false)
    expect(wrapper.emitted('rules-updated')).toBeTruthy()
  })
})
