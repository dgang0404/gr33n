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

const testRouter = createRouter({
  history: createMemoryHistory(),
  routes: [{ path: '/', component: { template: '<div/>' } }],
})

describe('GuardianChatPanel — proposals in transcript (Phase 29 WS4)', () => {
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

  it('attaches proposal cards from SSE done payload', async () => {
    const farmContext = useFarmContextStore()
    farmContext.farmId = 1

    const donePayload = {
      answer: 'I can acknowledge that alert for you.',
      session_id: 'sess-1',
      turn_index: 0,
      proposals: [{
        proposal_id: '550e8400-e29b-41d4-a716-446655440000',
        tool: 'ack_alert',
        args: { alert_id: 4 },
        summary: 'Acknowledge: Humidity high — Flower Room',
        expires_at: new Date(Date.now() + 300_000).toISOString(),
      }],
    }

    const encoder = new TextEncoder()
    const sseBody = [
      'event: delta\ndata: {"text":"ok"}\n\n',
      `event: done\ndata: ${JSON.stringify(donePayload)}\n\n`,
      'data: [DONE]\n\n',
    ].join('')
    const stream = new ReadableStream({
      start(controller) {
        controller.enqueue(encoder.encode(sseBody))
        controller.close()
      },
    })

    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, body: stream }))

    const wrapper = mount(GuardianChatPanel, {
      props: { layout: 'compact' },
      global: { plugins: [testRouter] },
    })
    await flushPromises()

    await wrapper.find('[data-test="chat-message-input"]').setValue('acknowledge the humidity alert')
    await wrapper.find('[data-test="chat-send-button"]').trigger('click')
    await flushPromises()
    await flushPromises()

    expect(wrapper.find('[data-test="chat-turn-proposals"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="guardian-proposal-card"]').exists()).toBe(true)

    vi.unstubAllGlobals()
  })
})
