import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { setActivePinia, createPinia } from 'pinia'
import GuardianSettingsCorpusCard from '../components/GuardianSettingsCorpusCard.vue'
import { useFarmContextStore } from '../stores/farmContext'
import { useCapabilitiesStore } from '../stores/capabilities'
import { useGuardianReadinessStore } from '../stores/guardianReadiness'

vi.mock('../api', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    interceptors: { request: { use: vi.fn() }, response: { use: vi.fn() } },
  },
}))

import api from '../api'

function stubCorpusHealth() {
  api.get.mockImplementation((url) => {
    if (url === '/v1/chat/health') {
      return Promise.resolve({
        data: {
          awakening: {
            state: 'ready',
            corpus: {
              field_guide_chunks: 58,
              field_guide_last_ingested_at: '2026-07-04T10:00:00Z',
              field_guide_freshness: 'aging',
              platform_doc_chunks: 12,
              platform_last_ingested_at: '2026-07-01T08:00:00Z',
              platform_freshness: 'aging',
              operational_chunks: 240,
              operational_last_ingested_at: '2026-06-10T12:00:00Z',
              operational_freshness: 'stale',
              staleness: 'operational_stale',
            },
          },
        },
      })
    }
    if (String(url).includes('/guardian/reingest/status')) {
      return Promise.resolve({ data: { status: 'idle' } })
    }
    return Promise.resolve({ data: {} })
  })
}

describe('Phase 135 — GuardianSettingsCorpusCard', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    useCapabilitiesStore().aiEnabled = true
    useFarmContextStore().farmId = 1
    useGuardianReadinessStore().loaded = true
    stubCorpusHealth()
  })

  it('renders corpus table with staleness warning', async () => {
    const wrapper = mount(GuardianSettingsCorpusCard, { props: { isFarmAdmin: true } })
    await flushPromises()
    expect(wrapper.find('[data-test="settings-guardian-corpus"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="settings-guardian-corpus-warn"]').text()).toMatch(/stale/i)
    expect(wrapper.get('[data-test="settings-corpus-chunks-operational"]').text()).toContain('240')
    expect(wrapper.find('[data-test="settings-corpus-reingest-field_guides"]').exists()).toBe(true)
    wrapper.unmount()
  })

  it('posts re-ingest for admin', async () => {
    api.post.mockResolvedValue({ data: { status: 'running', scope: 'field_guides' } })
    api.get.mockImplementation((url) => {
      if (String(url).includes('/guardian/reingest/status')) {
        return Promise.resolve({ data: { status: 'running', scope: 'field_guides' } })
      }
      return stubCorpusHealth().then ? stubCorpusHealth() : Promise.resolve({ data: {} })
    })
    stubCorpusHealth()
    const wrapper = mount(GuardianSettingsCorpusCard, { props: { isFarmAdmin: true } })
    await flushPromises()
    await wrapper.get('[data-test="settings-corpus-reingest-field_guides"]').trigger('click')
    await flushPromises()
    expect(api.post).toHaveBeenCalledWith('/farms/1/guardian/reingest', { scope: 'field_guides' })
    wrapper.unmount()
  })
})
