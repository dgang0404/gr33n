/**
 * Phase 182 — Guardian quick UX wins closure.
 */
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createRouter, createMemoryHistory } from 'vue-router'
import GuardianRequestsInbox from '../components/GuardianRequestsInbox.vue'
import { useFarmContextStore } from '../stores/farmContext'
import { useGuardianProposalsStore } from '../stores/guardianProposals'
import { isUnauthorizedError, resetUnauthorizedGate } from '../api/index.js'

vi.mock('../api', async (importOriginal) => {
  const actual = await importOriginal()
  return {
    ...actual,
    default: {
      get: vi.fn(),
      post: vi.fn(),
      interceptors: { request: { use: vi.fn() }, response: { use: vi.fn() } },
    },
  }
})

vi.mock('../composables/useFarmOperate', () => ({
  useFarmOperate: () => ({ canOperate: { value: true }, loading: { value: false } }),
}))

import api from '../api'

const router = createRouter({
  history: createMemoryHistory(),
  routes: [{ path: '/chat', component: { template: '<div/>' } }],
})

describe('Phase 182 — api 401 helpers', () => {
  it('isUnauthorizedError detects 401', () => {
    expect(isUnauthorizedError({ response: { status: 401 } })).toBe(true)
    expect(isUnauthorizedError({ response: { status: 500 } })).toBe(false)
  })
})

describe('Phase 182 — pending inbox scroll + sort', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('shows sticky count and scroll container when proposals exist', async () => {
    api.get.mockResolvedValue({
      data: {
        proposals: [
          {
            proposal_id: 'p-old',
            tool: 'ack_alert',
            args: {},
            summary: 'Older',
            risk_tier: 'low',
            expires_at: new Date(Date.now() + 60000).toISOString(),
            created_at: '2026-07-01T10:00:00Z',
            farm_id: 1,
            status: 'pending',
          },
          {
            proposal_id: 'p-new',
            tool: 'create_task',
            args: {},
            summary: 'Newer',
            risk_tier: 'medium',
            expires_at: new Date(Date.now() + 60000).toISOString(),
            created_at: '2026-07-13T10:00:00Z',
            farm_id: 1,
            status: 'pending',
          },
        ],
        total: 2,
        limit: 50,
        offset: 0,
      },
    })

    useFarmContextStore().farmId = 1
    const wrapper = mount(GuardianRequestsInbox, { global: { plugins: [router] } })
    await flushPromises()

    expect(wrapper.find('[data-test="guardian-inbox-count"]').text()).toContain('2 requests')
    expect(wrapper.find('[data-test="guardian-inbox-scroll"]').exists()).toBe(true)

    const store = useGuardianProposalsStore()
    expect(store.proposals[0].proposal_id).toBe('p-new')
    expect(store.proposals[1].proposal_id).toBe('p-old')
  })
})

describe('Phase 182 — refine hint', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    resetUnauthorizedGate()
  })

  it('shows hint when message contains Correction:', async () => {
    const { default: GuardianChatPanel } = await import('../components/GuardianChatPanel.vue')
    vi.mocked(api.get).mockImplementation((url) => {
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
        return Promise.resolve({ data: { awakening: { state: 'ready' } } })
      }
      return Promise.resolve({ data: {} })
    })

    const ownerID = '00000000-0000-0000-0000-000000000001'
    localStorage.setItem('gr33n_farm_id', '1')
    const { useCapabilitiesStore } = await import('../stores/capabilities')
    const { useFarmStore } = await import('../stores/farm')
    useCapabilitiesStore().aiEnabled = true
    useFarmContextStore().farmId = 1
    useFarmStore().farm = { id: 1, owner_user_id: ownerID }

    const wrapper = mount(GuardianChatPanel, {
      props: { layout: 'full' },
      global: { plugins: [router] },
    })
    await flushPromises()
    await flushPromises()

    const textarea = wrapper.find('[data-test="chat-message-input"]')
    await textarea.setValue('Set feeding plan volume to 0.3L\n\nCorrection: ')
    await flushPromises()

    expect(wrapper.find('[data-test="chat-refine-hint"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="chat-refine-hint"]').text()).toContain('same session')
  })
})
