/**
 * Phase 179 — Guardian chat status consolidation closure.
 */
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createRouter, createMemoryHistory } from 'vue-router'
import GuardianAwakeningPanel from '../components/GuardianAwakeningPanel.vue'
import { useGuardianReadinessStore } from '../stores/guardianReadiness'

vi.mock('../api', () => ({
  default: {
    get: vi.fn(() => Promise.resolve({ data: { awakening: { state: 'busy' } } })),
    post: vi.fn(),
    interceptors: { request: { use: vi.fn() }, response: { use: vi.fn() } },
  },
  isUnauthorizedError: () => false,
  resetUnauthorizedGate: () => {},
}))

describe('Phase 179 — awakening busy suppression', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    const readiness = useGuardianReadinessStore()
    readiness.awakening = { state: 'busy', messages: [] }
    readiness.loaded = true
  })

  it('hides awakening panel when suppressBusy and state is busy', () => {
    const wrapper = mount(GuardianAwakeningPanel, {
      props: { farmId: 1, suppressBusy: true },
    })
    expect(wrapper.find('[data-test="guardian-awakening-panel"]').exists()).toBe(false)
  })

  it('shows awakening panel when suppressBusy is false', () => {
    const wrapper = mount(GuardianAwakeningPanel, {
      props: { farmId: 1, suppressBusy: false },
    })
    expect(wrapper.find('[data-test="guardian-awakening-headline"]').text()).toContain('answering')
  })
})

describe('Phase 179 — grounded block during local stream', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('does not show composer busy block when streaming locally', async () => {
    const { default: GuardianChatPanel } = await import('../components/GuardianChatPanel.vue')
    const api = (await import('../api')).default
    api.get.mockImplementation((url) => {
      if (url === '/guardian/models') {
        return Promise.resolve({
          data: {
            server_default: 'phi3:mini',
            available_models: [{ name: 'phi3:mini', context_window: 131072, capabilities: ['completion'] }],
          },
        })
      }
      if (url === '/v1/chat/sessions') return Promise.resolve({ data: { sessions: [] } })
      if (url === '/v1/chat/health') {
        return Promise.resolve({ data: { awakening: { state: 'busy' } } })
      }
      return Promise.resolve({ data: {} })
    })

    const ownerID = '00000000-0000-0000-0000-000000000001'
    localStorage.setItem('gr33n_farm_id', '1')
    const { useCapabilitiesStore } = await import('../stores/capabilities')
    const { useFarmContextStore } = await import('../stores/farmContext')
    const { useFarmStore } = await import('../stores/farm')
    const { useGuardianChatStore } = await import('../stores/guardianChat')
    const { useGuardianReadinessStore } = await import('../stores/guardianReadiness')

    useCapabilitiesStore().aiEnabled = true
    useFarmContextStore().farmId = 1
    useFarmStore().farm = { id: 1, owner_user_id: ownerID, guardian_preferred_model: 'phi3:mini' }
    useGuardianReadinessStore().awakening = { state: 'busy' }
    useGuardianReadinessStore().loaded = true
    useGuardianReadinessStore().mode = 'farm_counsel'
    useGuardianChatStore().streaming = true

    const router = createRouter({
      history: createMemoryHistory(),
      routes: [{ path: '/', component: { template: '<div/>' } }],
    })

    const wrapper = mount(GuardianChatPanel, {
      props: { layout: 'full' },
      global: { plugins: [router] },
    })
    await flushPromises()
    await flushPromises()

    expect(wrapper.find('[data-test="chat-grounded-model-block"]').exists()).toBe(false)
    expect(wrapper.find('[data-test="chat-streaming-row"]').exists()).toBe(true)
  })
})
