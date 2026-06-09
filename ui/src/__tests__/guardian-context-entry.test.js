import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { setActivePinia, createPinia } from 'pinia'

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
import AskGuardianButton from '../components/AskGuardianButton.vue'
import Alerts from '../views/Alerts.vue'
import GuardianChatPanel from '../components/GuardianChatPanel.vue'
import { useGuardianPanelStore } from '../stores/guardianPanel'
import { useCapabilitiesStore } from '../stores/capabilities'
import { useFarmContextStore } from '../stores/farmContext'
import { useFarmStore } from '../stores/farm'

const testRouter = createRouter({
  history: createMemoryHistory(),
  routes: [
    { path: '/', name: 'dashboard', component: { template: '<div/>' } },
    { path: '/plants', name: 'plants', component: { template: '<div/>' } },
  ],
})

function mountChatPanel() {
  return mount(GuardianChatPanel, {
    props: { layout: 'compact' },
    global: { plugins: [testRouter] },
  })
}

function stubCapabilities() {
  api.get.mockImplementation((url) => {
    if (url === '/capabilities') return Promise.resolve({ data: { ai_enabled: true } })
    return Promise.resolve({ data: {} })
  })
}

describe('Phase 29 WS6 — AskGuardianButton', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    stubCapabilities()
  })

  it('opens drawer with prefill and contextRef', async () => {
    await useCapabilitiesStore().fetch()
    const panel = useGuardianPanelStore()

    const wrapper = mount(AskGuardianButton, {
      props: {
        prefilledMessage: 'Explain alert #42 and suggest next steps',
        contextRef: { type: 'alert', id: 42 },
      },
    })

    await wrapper.find('[data-test="ask-guardian-button"]').trigger('click')

    expect(panel.open).toBe(true)
    expect(panel.prefilledMessage).toBe('Explain alert #42 and suggest next steps')
    expect(panel.contextRef).toEqual({ type: 'alert', id: 42 })
  })

  it('hides when AI is disabled', async () => {
    api.get.mockImplementation((url) => {
      if (url === '/capabilities') return Promise.resolve({ data: { ai_enabled: false } })
      return Promise.resolve({ data: {} })
    })
    await useCapabilitiesStore().fetch()

    const wrapper = mount(AskGuardianButton, {
      props: {
        prefilledMessage: 'hello',
        contextRef: { type: 'alert', id: 1 },
      },
    })

    expect(wrapper.find('[data-test="ask-guardian-button"]').exists()).toBe(false)
  })
})

describe('Phase 29 WS6 — Alerts contextual entry', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    stubCapabilities()
    localStorage.setItem('gr33n_farm_id', '1')

    const farmContext = useFarmContextStore()
    farmContext.farmId = 1

    const farmStore = useFarmStore()
    farmStore.alerts = [
      {
        id: 99,
        is_read: false,
        is_acknowledged: false,
        severity: 'high',
        subject_rendered: 'Humidity high',
        message_text_rendered: '72% RH in Flower Room',
        created_at: new Date().toISOString(),
      },
      {
        id: 100,
        is_read: true,
        is_acknowledged: true,
        severity: 'low',
        subject_rendered: 'Already read',
        message_text_rendered: 'noop',
        created_at: new Date().toISOString(),
      },
    ]
    farmStore.loadAlerts = vi.fn().mockResolvedValue(farmStore.alerts)
    farmStore.countUnreadAlerts = vi.fn().mockResolvedValue(1)
    farmStore.loadTasks = vi.fn().mockResolvedValue([])
  })

  it('shows Ask Guardian on unread alert rows only', async () => {
    await useCapabilitiesStore().fetch()
    const wrapper = mount(Alerts)
    await flushPromises()

    const buttons = wrapper.findAll('[data-test="ask-guardian-button"]')
    expect(buttons).toHaveLength(1)
  })

  it('clicking Ask Guardian on alert opens drawer with alert context', async () => {
    await useCapabilitiesStore().fetch()
    const panel = useGuardianPanelStore()
    const wrapper = mount(Alerts)
    await flushPromises()

    await wrapper.find('[data-test="ask-guardian-button"]').trigger('click')

    expect(panel.open).toBe(true)
    expect(panel.contextRef).toEqual({ type: 'alert', id: 99 })
    expect(panel.prefilledMessage).toContain('alert #99')
  })
})

describe('Phase 29 WS6 — GuardianChatPanel context_ref POST', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    api.get.mockImplementation((url) => {
      if (url === '/capabilities') return Promise.resolve({ data: { ai_enabled: true } })
      if (url === '/v1/chat/sessions') return Promise.resolve({ data: { sessions: [] } })
      return Promise.resolve({ data: {} })
    })
    localStorage.setItem('gr33n_farm_id', '1')
  })

  it('includes context_ref in POST /v1/chat when store has contextRef', async () => {
    const farmContext = useFarmContextStore()
    farmContext.farmId = 1
    farmContext.farms = [{ id: 1, name: 'Demo Farm' }]

    const panel = useGuardianPanelStore()
    panel.openDrawer({
      prefilledMessage: 'Explain alert #42',
      contextRef: { type: 'alert', id: 42 },
    })

    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      body: {
        getReader: () => ({
          read: async () => ({ done: true, value: undefined }),
        }),
      },
    })
    vi.stubGlobal('fetch', fetchMock)

    const wrapper = mountChatPanel()
    await flushPromises()

    await wrapper.find('[data-test="chat-message-input"]').setValue('Explain alert #42')
    await wrapper.find('[data-test="chat-send-button"]').trigger('click')
    await flushPromises()

    const [, opts] = fetchMock.mock.calls[0]
    const body = JSON.parse(opts.body)
    expect(body.context_ref).toEqual({ type: 'alert', id: 42 })
    expect(body.farm_id).toBe(1)

    vi.unstubAllGlobals()
  })

  it('includes route context_ref when no Ask Guardian entity ref is set', async () => {
    const farmContext = useFarmContextStore()
    farmContext.farmId = 1
    farmContext.farms = [{ id: 1, name: 'Demo Farm' }]

    const panel = useGuardianPanelStore()
    panel.setRouteFromRouter({ path: '/fertigation', meta: {} })

    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      body: {
        getReader: () => ({
          read: async () => ({ done: true, value: undefined }),
        }),
      },
    })
    vi.stubGlobal('fetch', fetchMock)

    const wrapper = mountChatPanel()
    await flushPromises()

    await wrapper.find('[data-test="chat-message-input"]').setValue('How do I add a program?')
    await wrapper.find('[data-test="chat-send-button"]').trigger('click')
    await flushPromises()

    const [, opts] = fetchMock.mock.calls[0]
    const body = JSON.parse(opts.body)
    expect(body.context_ref).toEqual({
      type: 'route',
      path: '/fertigation',
      name: 'Feeding (technical)',
    })

    vi.unstubAllGlobals()
  })

  it('prefers Ask Guardian entity contextRef over routeRef', async () => {
    const farmContext = useFarmContextStore()
    farmContext.farmId = 1
    farmContext.farms = [{ id: 1, name: 'Demo Farm' }]

    const panel = useGuardianPanelStore()
    panel.setRouteFromRouter({ path: '/fertigation', meta: {} })
    panel.openDrawer({
      prefilledMessage: 'Explain alert #42',
      contextRef: { type: 'alert', id: 42 },
    })

    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      body: {
        getReader: () => ({
          read: async () => ({ done: true, value: undefined }),
        }),
      },
    })
    vi.stubGlobal('fetch', fetchMock)

    const wrapper = mountChatPanel()
    await flushPromises()

    await wrapper.find('[data-test="chat-message-input"]').setValue('Explain alert #42')
    await wrapper.find('[data-test="chat-send-button"]').trigger('click')
    await flushPromises()

    const [, opts] = fetchMock.mock.calls[0]
    const body = JSON.parse(opts.body)
    expect(body.context_ref).toEqual({ type: 'alert', id: 42 })

    vi.unstubAllGlobals()
  })
})
