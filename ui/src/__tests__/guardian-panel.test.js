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
import App from '../App.vue'
import GuardianDrawer from '../components/GuardianDrawer.vue'
import GuardianEdgeTab from '../components/GuardianEdgeTab.vue'
import GuardianChatPanel from '../components/GuardianChatPanel.vue'
import { useGuardianPanelStore } from '../stores/guardianPanel'
import { useFarmContextStore } from '../stores/farmContext'
import { useCapabilitiesStore } from '../stores/capabilities'

const router = createRouter({
  history: createMemoryHistory(),
  routes: [
    { path: '/', component: { template: '<div data-test="zones-page">Zones</div>' } },
    { path: '/chat', component: { template: '<div>Chat</div>' } },
  ],
})

function stubCapabilities() {
  api.get.mockImplementation((url) => {
    if (url === '/capabilities') return Promise.resolve({ data: { ai_enabled: true } })
    if (url === '/v1/chat/sessions') return Promise.resolve({ data: { sessions: [] } })
    if (url === '/farms') return Promise.resolve({ data: [{ id: 1, name: 'Demo Farm' }] })
    if (url === '/health') return Promise.resolve({ data: { ok: true } })
    return Promise.resolve({ data: {} })
  })
}

describe('Phase 29 WS1 — guardian panel store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    localStorage.setItem('gr33n_token', 'test-token')
    localStorage.setItem('gr33n_farm_id', '1')
  })

  it('toggle opens and closes the drawer', () => {
    const panel = useGuardianPanelStore()
    expect(panel.open).toBe(false)
    panel.toggle()
    expect(panel.open).toBe(true)
    panel.close()
    expect(panel.open).toBe(false)
  })

  it('openDrawer sets prefill, contextRef, and session id', () => {
    const panel = useGuardianPanelStore()
    panel.openDrawer({
      prefilledMessage: 'Explain alert #42',
      contextRef: { type: 'alert', id: 42 },
      activeSessionId: 'sess-abc',
    })
    expect(panel.open).toBe(true)
    expect(panel.prefilledMessage).toBe('Explain alert #42')
    expect(panel.contextRef).toEqual({ type: 'alert', id: 42 })
    expect(panel.activeSessionId).toBe('sess-abc')
  })
})

describe('Phase 29 WS1 — GuardianDrawer', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    stubCapabilities()
  })

  it('renders teleported drawer with compact session picker when open', async () => {
    const panel = useGuardianPanelStore()
    panel.open = true

    const wrapper = mount(GuardianDrawer, {
      global: { plugins: [router] },
      attachTo: document.body,
    })
    await flushPromises()

    expect(document.body.querySelector('[data-test="guardian-drawer"]')).not.toBeNull()
    expect(document.body.querySelector('[data-test="chat-sessions-compact"]')).not.toBeNull()
    wrapper.unmount()
    document.body.innerHTML = ''
  })
})

describe('Phase 29 WS1 — GuardianChatPanel farm context', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    stubCapabilities()
    localStorage.setItem('gr33n_farm_id', '1')
  })

  it('includes farm_id in POST /v1/chat when farm is selected and context is on', async () => {
    const farmContext = useFarmContextStore()
    farmContext.farmId = 1
    farmContext.farms = [{ id: 1, name: 'Demo Farm' }]

    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      body: {
        getReader: () => ({
          read: async () => ({ done: true, value: undefined }),
        }),
      },
    })
    vi.stubGlobal('fetch', fetchMock)

    const wrapper = mount(GuardianChatPanel, { props: { layout: 'compact' } })
    await flushPromises()

    await wrapper.find('[data-test="chat-message-input"]').setValue('what is the humidity trend?')
    const farmCheckbox = wrapper.find('[data-test="chat-use-farm-context"]')
    expect(farmCheckbox.element.checked).toBe(true)

    await wrapper.find('[data-test="chat-send-button"]').trigger('click')
    await flushPromises()

    expect(fetchMock).toHaveBeenCalled()
    const [, opts] = fetchMock.mock.calls[0]
    const body = JSON.parse(opts.body)
    expect(body.farm_id).toBe(1)
    expect(body.message).toBe('what is the humidity trend?')

    vi.unstubAllGlobals()
  })
})

describe('Phase 29 WS1 — drawer on any route', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    stubCapabilities()
    localStorage.setItem('gr33n_token', 'test-token')
    vi.stubGlobal('EventSource', class {
      addEventListener() {}
      close() {}
    })
  })

  it('App keeps Zones visible when drawer opens and closes', async () => {
    await router.push('/')
    const wrapper = mount(App, {
      global: { plugins: [router] },
      attachTo: document.body,
    })
    await flushPromises()
    await flushPromises()

    expect(wrapper.find('[data-test="zones-page"]').exists()).toBe(true)

    const panel = useGuardianPanelStore()
    panel.toggle()
    await flushPromises()
    expect(document.body.querySelector('[data-test="guardian-drawer"]')).not.toBeNull()
    expect(wrapper.find('[data-test="zones-page"]').exists()).toBe(true)

    panel.close()
    await flushPromises()
    expect(document.body.querySelector('[data-test="guardian-drawer"]')).toBeNull()
    expect(wrapper.find('[data-test="zones-page"]').exists()).toBe(true)

    wrapper.unmount()
    vi.unstubAllGlobals()
  })
})

describe('Phase 29 WS1 — GuardianEdgeTab', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    stubCapabilities()
    document.body.innerHTML = ''
  })

  it('renders edge tab when AI is enabled and drawer is closed', async () => {
    await useCapabilitiesStore().fetch()
    const wrapper = mount(GuardianEdgeTab, { attachTo: document.body })
    await flushPromises()
    expect(document.body.querySelector('[data-test="guardian-edge-tab"]')).not.toBeNull()
    wrapper.unmount()
  })

  it('hides edge tab while drawer is open', async () => {
    await useCapabilitiesStore().fetch()
    useGuardianPanelStore().openDrawer()
    const wrapper = mount(GuardianEdgeTab, { attachTo: document.body })
    await flushPromises()
    expect(document.body.querySelector('[data-test="guardian-edge-tab"]')).toBeNull()
    wrapper.unmount()
  })
})
