import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createRouter, createMemoryHistory } from 'vue-router'

vi.mock('../api', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    patch: vi.fn(),
    delete: vi.fn(),
    interceptors: { request: { use: vi.fn() }, response: { use: vi.fn() } },
  },
}))

import api from '../api'
import GuardianChatPanel from '../components/GuardianChatPanel.vue'
import { useFarmContextStore } from '../stores/farmContext'
import { useCapabilitiesStore } from '../stores/capabilities'
import { useFarmStore } from '../stores/farm'

const ownerID = '00000000-0000-0000-0000-000000000001'

const testRouter = createRouter({
  history: createMemoryHistory(),
  routes: [{ path: '/', component: { template: '<div/>' } }],
})

function stubChatApi({ serverDefault = 'tinyllama' } = {}) {
  api.get.mockImplementation((url) => {
    if (url === '/guardian/models') {
      return Promise.resolve({
        data: {
          server_default: serverDefault,
          available_models: [
            { name: 'tinyllama', context_window: 2048, capabilities: ['completion'] },
            { name: 'phi3:mini', context_window: 131072, capabilities: ['completion'] },
          ],
        },
      })
    }
    if (url === '/v1/chat/sessions') return Promise.resolve({ data: { sessions: [] } })
    if (url === '/farms/1') {
      return Promise.resolve({
        data: { id: 1, owner_user_id: ownerID, guardian_preferred_model: '' },
      })
    }
    if (url === '/farms/1/members') {
      return Promise.resolve({
        data: [{ user_id: ownerID, role_in_farm: 'owner' }],
      })
    }
    return Promise.resolve({ data: {} })
  })
}

describe('GuardianChatPanel grounded model gate', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    localStorage.setItem('gr33n_token', 'test-token')
    localStorage.setItem('gr33n_farm_id', '1')
    stubChatApi()
    const caps = useCapabilitiesStore()
    caps.aiEnabled = true
    useFarmContextStore().farmId = 1
    useFarmStore().farm = { id: 1, owner_user_id: ownerID, guardian_preferred_model: '' }
  })

  it('blocks send when farm context on and server default is tinyllama', async () => {
    const wrapper = mount(GuardianChatPanel, {
      props: { layout: 'full' },
      global: { plugins: [testRouter] },
    })
    await flushPromises()
    await flushPromises()

    expect(wrapper.find('[data-test="chat-grounded-model-block"]').exists()).toBe(false)
    const sessionSelect = wrapper.find('[data-test="guardian-session-model"]')
    expect(sessionSelect.exists()).toBe(true)
    expect(sessionSelect.element.value).toBe('phi3:mini')
    const optionValues = sessionSelect.findAll('option').map((o) => o.element.value)
    expect(optionValues).not.toContain('tinyllama')

    const send = wrapper.find('[data-test="chat-send-button"]')
    await wrapper.find('[data-test="chat-message-input"]').setValue('hello')
    expect(send.attributes('disabled')).toBeUndefined()
  })

  it('enables send after turning off farm context', async () => {
    const wrapper = mount(GuardianChatPanel, {
      props: { layout: 'full' },
      global: { plugins: [testRouter] },
    })
    await flushPromises()
    await flushPromises()

    await wrapper.find('[data-test="guardian-mode-quick"]').trigger('click')
    await flushPromises()

    expect(wrapper.find('[data-test="chat-grounded-model-block"]').exists()).toBe(false)
    const sessionSelect = wrapper.find('[data-test="guardian-session-model"]')
    expect(sessionSelect.exists()).toBe(true)
    await sessionSelect.setValue('phi3:mini')
    await flushPromises()
    expect(sessionSelect.element.value).toBe('phi3:mini')

    await wrapper.find('[data-test="chat-message-input"]').setValue('hello')
    const send = wrapper.find('[data-test="chat-send-button"]')
    expect(send.attributes('disabled')).toBeUndefined()
  })
})
