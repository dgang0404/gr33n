// Phase 152 WS2 — clickable citation deep links. A citation with a resolved
// `route` renders as a router-link (with the sidebar nav-hint wiggle); one
// without still renders as plain text, matching pre-152 behavior.
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
import { navHint } from '../directives/navHint.js'
import { useGuardianChatStore } from '../stores/guardianChat'
import { useFarmContextStore } from '../stores/farmContext'
import { useCapabilitiesStore } from '../stores/capabilities'
import { useFarmStore } from '../stores/farm'

const ownerID = '00000000-0000-0000-0000-000000000001'

const testRouter = createRouter({
  history: createMemoryHistory(),
  routes: [
    { path: '/', component: { template: '<div/>' } },
    { path: '/zones/:id', component: { template: '<div data-test="zone-detail-page"/>' } },
    { path: '/crop-cycles/:id/summary', component: { template: '<div/>' } },
  ],
})

function stubChatApi() {
  api.get.mockImplementation((url) => {
    if (url === '/guardian/models') {
      return Promise.resolve({
        data: { server_default: 'phi3:mini', available_models: [{ name: 'phi3:mini', context_window: 131072 }] },
      })
    }
    if (url === '/v1/chat/sessions') return Promise.resolve({ data: { sessions: [] } })
    if (url === '/farms/1') {
      return Promise.resolve({ data: { id: 1, owner_user_id: ownerID, guardian_preferred_model: '' } })
    }
    if (url === '/farms/1/members') {
      return Promise.resolve({ data: [{ user_id: ownerID, role_in_farm: 'owner' }] })
    }
    return Promise.resolve({ data: {} })
  })
}

describe('GuardianChatPanel citation deep links', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    localStorage.setItem('gr33n_farm_id', '1')
    stubChatApi()
    useCapabilitiesStore().aiEnabled = true
    useFarmContextStore().farmId = 1
    useFarmStore().farm = { id: 1, owner_user_id: ownerID, guardian_preferred_model: '' }
  })

  it('renders a router-link for citations with a resolved route, plain text otherwise', async () => {
    const wrapper = mount(GuardianChatPanel, {
      props: { layout: 'full' },
      global: { plugins: [testRouter], directives: { 'nav-hint': navHint } },
    })
    await flushPromises()
    await flushPromises()

    useGuardianChatStore().appendTurn({
      turn_index: 0,
      user_message: 'How is the Flower run doing?',
      assistant_message: 'Stage is early_flower [1], per the fertigation plan [2].',
      llm_model: 'phi3:mini',
      grounded: true,
      context_count: 2,
      citations: [
        { ref: 1, chunk_id: 10, source_type: 'crop_cycle', source_id: 2, route: '/crop-cycles/2/summary', excerpt: 'crop_cycle: Flower run' },
        { ref: 2, chunk_id: 11, source_type: 'schedule', source_id: 10, excerpt: 'schedule: Water Early Flower Daily' },
      ],
    })
    await flushPromises()

    const links = wrapper.findAll('[data-test="chat-citation-link"]')
    expect(links).toHaveLength(1)
    expect(links[0].attributes('href')).toBe('/crop-cycles/2/summary')
    expect(links[0].text()).toContain('[1]')

    const items = wrapper.findAll('li')
    const scheduleItem = items.find((li) => li.text().includes('[2]'))
    expect(scheduleItem.find('[data-test="chat-citation-link"]').exists()).toBe(false)
  })
})
