import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { setActivePinia, createPinia } from 'pinia'
import GuardianSettingsAwakeningCard from '../components/GuardianSettingsAwakeningCard.vue'
import { useFarmContextStore } from '../stores/farmContext'
import { useCapabilitiesStore } from '../stores/capabilities'

vi.mock('../api', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    interceptors: { request: { use: vi.fn() }, response: { use: vi.fn() } },
  },
}))

import api from '../api'

function stubHealth(overrides = {}) {
  api.get.mockImplementation((url) => {
    if (url === '/v1/chat/health') {
      return Promise.resolve({
        data: {
          awakening: {
            state: 'ready',
            chat_model: 'phi3:mini',
            chat_model_loaded: true,
            profile: 'cpu_laptop',
            field_guide_chunks: 58,
            platform_doc_chunks: 12,
            rag_corpus_ok: true,
            ...overrides,
          },
        },
      })
    }
    return Promise.resolve({ data: {} })
  })
}

describe('Phase 129 WS8 — GuardianSettingsAwakeningCard', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    const caps = useCapabilitiesStore()
    caps.loaded = true
    caps.aiEnabled = true
    const farmContext = useFarmContextStore()
    farmContext.farmId = 1
  })

  it('hidden when AI is disabled (Lite mode)', async () => {
    useCapabilitiesStore().aiEnabled = false
    stubHealth()
    const wrapper = mount(GuardianSettingsAwakeningCard, {
      props: { isFarmAdmin: true },
    })
    await flushPromises()
    expect(wrapper.find('[data-test="settings-guardian-awakening"]').exists()).toBe(false)
    wrapper.unmount()
  })

  it('shows readiness state and corpus counts', async () => {
    stubHealth({ state: 'sleeping', chat_model_loaded: false, rag_corpus_ok: false })
    const wrapper = mount(GuardianSettingsAwakeningCard, {
      props: { isFarmAdmin: false },
    })
    await flushPromises()
    expect(wrapper.find('[data-test="settings-guardian-awakening"]').exists()).toBe(true)
    expect(wrapper.get('[data-test="settings-guardian-state"]').text()).toContain('Sleeping')
    expect(wrapper.get('[data-test="settings-guardian-field-chunks"]').text()).toContain('58')
    expect(wrapper.get('[data-test="settings-guardian-platform-chunks"]').text()).toContain('12')
    expect(wrapper.find('[data-test="settings-guardian-rag-warn"]').exists()).toBe(true)
    wrapper.unmount()
  })

  it('Awaken now posts warmup', async () => {
    stubHealth({ state: 'sleeping' })
    api.post.mockResolvedValue({ data: { state: 'stirring' } })
    const wrapper = mount(GuardianSettingsAwakeningCard, {
      props: { isFarmAdmin: true },
    })
    await flushPromises()
    await wrapper.get('[data-test="settings-guardian-awaken-btn"]').trigger('click')
    await flushPromises()
    expect(api.post).toHaveBeenCalledWith('/guardian/warmup', { mode: 'farm_counsel', farm_id: 1 })
    wrapper.unmount()
  })

  it('Phase 163 — Rest now posts dormant when a model is loaded', async () => {
    stubHealth({ state: 'ready', chat_model_loaded: true })
    api.post.mockResolvedValue({ data: { state: 'dormant' } })
    const wrapper = mount(GuardianSettingsAwakeningCard, {
      props: { isFarmAdmin: true },
    })
    await flushPromises()
    await wrapper.get('[data-test="settings-guardian-rest-btn"]').trigger('click')
    await flushPromises()
    expect(api.post).toHaveBeenCalledWith('/guardian/dormant', { mode: 'farm_counsel', farm_id: 1 })
    wrapper.unmount()
  })

  it('Phase 163 — Rest now is disabled when no chat model is loaded', async () => {
    stubHealth({ state: 'sleeping', chat_model_loaded: false })
    const wrapper = mount(GuardianSettingsAwakeningCard, {
      props: { isFarmAdmin: true },
    })
    await flushPromises()
    expect(wrapper.get('[data-test="settings-guardian-rest-btn"]').attributes('disabled')).toBeDefined()
    wrapper.unmount()
  })

  it('Phase 163 — shows Resting state label for dormant', async () => {
    stubHealth({ state: 'dormant', chat_model_loaded: false })
    const wrapper = mount(GuardianSettingsAwakeningCard, {
      props: { isFarmAdmin: true },
    })
    await flushPromises()
    expect(wrapper.get('[data-test="settings-guardian-state"]').text()).toContain('Resting')
    wrapper.unmount()
  })

  it('Phase 163 WS3 — shows auto-rest countdown when configured', async () => {
    stubHealth({ state: 'ready', chat_model_loaded: true, auto_dormant_minutes: 45, idle_until_dormant_sec: 1800 })
    const wrapper = mount(GuardianSettingsAwakeningCard, {
      props: { isFarmAdmin: true },
    })
    await flushPromises()
    const hint = wrapper.get('[data-test="settings-guardian-auto-dormant"]')
    expect(hint.text()).toContain('45')
    expect(hint.text()).toContain('30')
    wrapper.unmount()
  })
})
