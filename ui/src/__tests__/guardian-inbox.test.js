/**
 * Phase 30 WS1 — Guardian pending requests inbox.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createRouter, createMemoryHistory } from 'vue-router'
import GuardianRequestsInbox from '../components/GuardianRequestsInbox.vue'
import { useFarmContextStore } from '../stores/farmContext'
import { useGuardianProposalsStore } from '../stores/guardianProposals'

vi.mock('../api', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
  },
}))

vi.mock('../composables/useFarmOperate', () => ({
  useFarmOperate: () => ({ canOperate: { value: true }, loading: { value: false } }),
}))

import api from '../api'

const router = createRouter({
  history: createMemoryHistory(),
  routes: [{ path: '/guardian/requests', component: { template: '<div/>' } }],
})

describe('Phase 30 WS1 — guardian inbox', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('loads pending proposals for selected farm', async () => {
    api.get.mockResolvedValue({
      data: {
        proposals: [
          {
            proposal_id: 'p-1',
            tool: 'ack_alert',
            args: { alert_id: 9 },
            summary: 'Acknowledge: humidity',
            expires_at: new Date(Date.now() + 60000).toISOString(),
            created_at: new Date().toISOString(),
            farm_id: 1,
            status: 'pending',
          },
        ],
        total: 1,
        limit: 50,
        offset: 0,
      },
    })

    const farmContext = useFarmContextStore()
    farmContext.farmId = 1

    const wrapper = mount(GuardianRequestsInbox, {
      global: { plugins: [router] },
    })
    await flushPromises()

    expect(api.get).toHaveBeenCalledWith('/v1/chat/proposals', {
      params: { farm_id: 1, status: 'pending', limit: 50, offset: 0 },
    })
    expect(wrapper.find('[data-test="guardian-inbox-list"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="guardian-proposal-card"]').exists()).toBe(true)
  })

  it('shows empty state when no pending proposals', async () => {
    api.get.mockResolvedValue({
      data: { proposals: [], total: 0, limit: 50, offset: 0 },
    })
    const farmContext = useFarmContextStore()
    farmContext.farmId = 1

    const wrapper = mount(GuardianRequestsInbox, {
      global: { plugins: [router] },
    })
    await flushPromises()

    expect(wrapper.find('[data-test="guardian-inbox-empty"]').exists()).toBe(true)
  })

  it('store refreshPendingCount uses total from API', async () => {
    api.get.mockResolvedValue({ data: { proposals: [], total: 3, limit: 1, offset: 0 } })
    const store = useGuardianProposalsStore()
    await store.refreshPendingCount(1)
    expect(store.pendingCount).toBe(3)
  })
})
