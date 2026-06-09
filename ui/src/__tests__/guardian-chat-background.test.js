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
import { useGuardianChatStore } from '../stores/guardianChat'

const testRouter = createRouter({
  history: createMemoryHistory(),
  routes: [{ path: '/', component: { template: '<div/>' } }],
})

function mountChatPanel() {
  return mount(GuardianChatPanel, {
    props: { layout: 'compact' },
    global: { plugins: [testRouter] },
  })
}

describe('Phase 37 WS9 — Guardian chat background stream', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    localStorage.setItem('gr33n_farm_id', '1')

    api.get.mockImplementation((url) => {
      if (url === '/v1/chat/sessions') return Promise.resolve({ data: { sessions: [] } })
      if (url === '/farms/1/members') {
        return Promise.resolve({
          data: [{ user_id: '00000000-0000-0000-0000-000000000001', role_in_farm: 'owner' }],
        })
      }
      return Promise.resolve({ data: {} })
    })
  })

  it('keeps streaming in the store after panel unmount', async () => {
    const farmContext = useFarmContextStore()
    farmContext.farmId = 1
    const chatStore = useGuardianChatStore()

    const encoder = new TextEncoder()
    const stream = new ReadableStream({
      start(controller) {
        controller.enqueue(encoder.encode('event: delta\ndata: {"text":"partial "}\n\n'))
      },
    })

    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, body: stream }))

    const wrapper = mountChatPanel()
    await flushPromises()

    await wrapper.find('[data-test="chat-message-input"]').setValue('wire the relay')
    void wrapper.find('[data-test="chat-send-button"]').trigger('click')
    await flushPromises()
    await flushPromises()

    expect(chatStore.streaming).toBe(true)
    expect(chatStore.streamingText).toContain('partial')

    wrapper.unmount()
    expect(chatStore.streaming).toBe(true)
    expect(chatStore.streamingText).toContain('partial')

    chatStore.cancelStream()
    expect(chatStore.streaming).toBe(false)

    vi.unstubAllGlobals()
  })

  it('remount shows accumulated streamingText from the store', async () => {
    const farmContext = useFarmContextStore()
    farmContext.farmId = 1
    const chatStore = useGuardianChatStore()
    chatStore.streaming = true
    chatStore.streamingText = 'still thinking'

    const wrapper = mountChatPanel()
    await flushPromises()
    expect(wrapper.find('[data-test="chat-streaming-row"]').text()).toContain('still thinking')
    wrapper.unmount()
  })
})
