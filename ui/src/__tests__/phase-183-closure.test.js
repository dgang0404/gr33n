/**
 * Phase 183 — contextual knowledge links, Help Library hub, task revise matchers.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createRouter, createMemoryHistory } from 'vue-router'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import SymptomCropLink from '../components/SymptomCropLink.vue'
import HelpLibraryHub from '../views/HelpLibraryHub.vue'
import { WORKSPACES } from '../lib/workspaces.js'
import { symptomGuideRoute } from '../lib/symptomGuideLink.js'

vi.mock('../api', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    interceptors: { request: { use: vi.fn() }, response: { use: vi.fn() } },
  },
  isUnauthorizedError: () => false,
  resetUnauthorizedGate: () => {},
}))

import api from '../api'

const router = createRouter({
  history: createMemoryHistory(),
  routes: [
    { path: '/operator-guide', component: { template: '<div/>' } },
    { path: '/symptom-guide', component: { template: '<div/>' } },
  ],
})

describe('Phase 183 WS1 — contextual symptom links', () => {
  it('symptomGuideRoute returns null without crop_key', () => {
    expect(symptomGuideRoute('')).toBeNull()
    expect(symptomGuideRoute(null)).toBeNull()
  })

  it('symptomGuideRoute builds pre-filtered symptom guide path', () => {
    expect(symptomGuideRoute('lettuce')).toEqual({
      path: '/symptom-guide',
      query: { crop_key: 'lettuce' },
    })
  })

  it('SymptomCropLink renders only when crop_key is known', async () => {
    const withCrop = mount(SymptomCropLink, {
      props: { cropKey: 'tomato' },
      global: { plugins: [router] },
    })
    expect(withCrop.find('[data-test="symptom-crop-link"]').exists()).toBe(true)
    expect(withCrop.find('[data-test="symptom-crop-link"]').text()).toMatch(/tomato/i)

    const without = mount(SymptomCropLink, {
      props: { cropKey: '' },
      global: { plugins: [router] },
    })
    expect(without.find('[data-test="symptom-crop-link"]').exists()).toBe(false)
  })
})

describe('Phase 183 WS2 — Help Library hub', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    api.get.mockResolvedValue({ data: { symptoms: [], ai_enabled: true } })
  })

  it('help workspace uses Library tab instead of four equal tabs', () => {
    expect(WORKSPACES.help.tabs.map((t) => t.id)).toEqual(['library', 'pi-setup'])
    expect(WORKSPACES.help.absorbs['/farm-knowledge']).toEqual({ tab: 'library', section: 'knowledge' })
    expect(WORKSPACES.help.absorbs['/symptom-guide']).toEqual({ tab: 'library', section: 'symptoms' })
  })

  it('HelpLibraryHub renders four library sections', async () => {
    await router.push('/operator-guide?tab=library')
    const wrapper = mount(HelpLibraryHub, {
      global: { plugins: [router] },
    })
    await flushPromises()

    expect(wrapper.find('[data-test="help-library-hub"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="help-library-section-guide"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="help-library-section-knowledge"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="help-library-section-symptoms"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="help-library-section-catalog"]').exists()).toBe(true)
  })

  it('legacy tab=knowledge deep link resolves to library hub', async () => {
    await router.push('/operator-guide?tab=knowledge&cited_doc=field-guides/crop-lettuce-nutrition.md')
    const wrapper = mount(HelpLibraryHub, {
      global: { plugins: [router] },
    })
    await flushPromises()
    expect(wrapper.find('[data-test="help-library-section-knowledge"]').exists()).toBe(true)
  })
})

describe('Phase 183 WS4 — closure', () => {
  const repoRoot = join(process.cwd(), '..')

  it('Go ships create_task revise matchers', () => {
    const revise = readFileSync(join(repoRoot, 'internal/farmguardian/proposals_revise.go'), 'utf8')
    expect(revise).toContain('create_task", "create_task_from_alert"')
    expect(revise).toContain('parseTaskTitleRevision')
    const tests = readFileSync(join(repoRoot, 'internal/farmguardian/proposals_revise_test.go'), 'utf8')
    expect(tests).toContain('TestApplyRevisionDeltas_CreateTaskTitleCallIt')
  })

  it('citation routes use library sections', () => {
    const route = readFileSync(join(repoRoot, 'internal/farmguardian/citation_route.go'), 'utf8')
    expect(route).toContain('tab=library')
    expect(route).toContain('section=knowledge')
  })
})
