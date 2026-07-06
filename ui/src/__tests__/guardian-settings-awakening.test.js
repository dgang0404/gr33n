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
})
