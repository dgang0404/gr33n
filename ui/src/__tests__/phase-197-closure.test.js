/**
 * Phase 197 — session sidebar labels for multi-turn pending threads.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { setActivePinia, createPinia } from 'pinia'
import { createRouter, createMemoryHistory } from 'vue-router'
import { useFarmContextStore } from '../stores/farmContext'

const repoRoot = join(process.cwd(), '..')

const evalSession = {
  session_id: 'a28a9684-1111-1111-1111-111111111111',
  title: '',
  first_user_message: 'Create a task to refill calcium nitrate when stock is low.',
  turn_count: 4,
  any_grounded: true,
  last_turn_at: '2026-07-16T12:00:00Z',
}

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

const router = createRouter({
  history: createMemoryHistory(),
  routes: [{ path: '/chat', component: { template: '<div />' } }],
})

function stubPanelApis() {
  api.get.mockImplementation((url) => {
    if (url === '/capabilities') return Promise.resolve({ data: { ai_enabled: true } })
    if (url === '/v1/chat/sessions') return Promise.resolve({ data: { sessions: [evalSession] } })
    if (url === '/v1/chat/proposals') {
      return Promise.resolve({
        data: {
          proposals: [{
            proposal_id: 'p-task',
            session_id: evalSession.session_id,
            summary: 'Create task: Refill calcium nitrate',
            status: 'pending',
            tool: 'create_task',
            farm_id: 1,
          }],
          total: 1,
          limit: 50,
          offset: 0,
        },
      })
    }
    if (url === '/v1/chat/health') {
      return Promise.resolve({ data: { awakening: { state: 'ready', rag_corpus_ok: true } } })
    }
    if (url === '/guardian/models') {
      return Promise.resolve({ data: { available_models: [], server_default: 'phi3:mini' } })
    }
    return Promise.resolve({ data: {} })
  })
}

describe('Phase 197 — session sidebar pending labels closure', () => {
  it('guardianProposals store exposes pendingBySessionId getter', () => {
    const store = readFileSync(join(repoRoot, 'ui/src/stores/guardianProposals.js'), 'utf8')
    expect(store).toContain('pendingBySessionId')
  })

  it('session label helper prefers pending proposal summary', () => {
    const lib = readFileSync(join(repoRoot, 'ui/src/lib/guardianSessionLabel.js'), 'utf8')
    expect(lib).toContain('Pending:')
    expect(lib).toContain('baseSessionLabel')
  })
})

describe('GuardianChatPanel — session pending labels (Phase 197)', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    stubPanelApis()
    localStorage.setItem('gr33n_farm_id', '1')
  })

  it('shows pending summary label and chip in session sidebar', async () => {
    const farmContext = useFarmContextStore()
    farmContext.farmId = 1

    const wrapper = mount(GuardianChatPanel, {
      props: { layout: 'full' },
      global: { plugins: [router] },
    })
    await flushPromises()
    await flushPromises()

    const sidebar = wrapper.find('[data-test="chat-sessions"]')
    expect(sidebar.text()).toContain('Pending: Create task: Refill calcium nitrate')
    expect(sidebar.text()).toContain('4 turns')
    expect(wrapper.find('[data-test="session-pending-chip"]').exists()).toBe(true)
  })
})
