/**
 * Phase 180 WS1 — Help knowledge surfaces map.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import HelpKnowledgeSurfacesMap from '../components/HelpKnowledgeSurfacesMap.vue'
import SymptomGuide from '../views/SymptomGuide.vue'
import { WORKSPACES } from '../lib/workspaces.js'
import { uniqueCropKeys, uniqueCategories } from '../lib/symptomGuideFilters.js'

vi.mock('../api', () => ({
  default: {
    get: vi.fn(),
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

describe('Phase 180 WS1 — help knowledge surfaces map', () => {
  it('renders four surface cards with correct links', async () => {
    const wrapper = mount(HelpKnowledgeSurfacesMap, {
      global: { plugins: [router] },
    })

    expect(wrapper.find('[data-test="help-knowledge-surfaces-map"]').exists()).toBe(true)

    const guide = wrapper.find('[data-test="help-surface-card-guide"]')
    const knowledge = wrapper.find('[data-test="help-surface-card-knowledge"]')
    const catalog = wrapper.find('[data-test="help-surface-card-catalog"]')
    const symptoms = wrapper.find('[data-test="help-surface-card-symptoms"]')

    expect(guide.exists()).toBe(true)
    expect(knowledge.exists()).toBe(true)
    expect(catalog.exists()).toBe(true)
    expect(symptoms.exists()).toBe(true)

    expect(guide.attributes('href')).toContain('/operator-guide')
    expect(guide.attributes('href')).toContain('tab=guide')
    expect(knowledge.attributes('href')).toContain('tab=knowledge')
    expect(catalog.attributes('href')).toContain('tab=catalog')
    expect(symptoms.attributes('href')).toContain('tab=symptoms')

    expect(guide.text()).toMatch(/Guide/i)
    expect(knowledge.text()).toMatch(/semantic search/i)
    expect(catalog.text()).toMatch(/Commons/i)
    expect(symptoms.text()).toMatch(/Symptom/i)
  })
})

describe('Phase 180 WS2 — symptoms tab + dropdowns', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    api.get.mockResolvedValue({
      data: {
        symptoms: [
          {
            id: 1,
            symptom_key: 'lettuce-tip-burn',
            display_name: 'Tip burn',
            crop_keys: ['lettuce'],
            categories: ['deficiency'],
            body_md: 'Edge necrosis on older leaves.',
          },
          {
            id: 2,
            symptom_key: 'tomato-blossom',
            display_name: 'Blossom end rot',
            crop_keys: ['tomato'],
            categories: ['deficiency'],
            body_md: 'Dark lesion on fruit bottom.',
          },
        ],
      },
    })
  })

  it('help workspace includes symptoms tab', () => {
    expect(WORKSPACES.help.tabs.map((t) => t.id)).toContain('symptoms')
    expect(WORKSPACES.help.absorbs['/symptom-guide']).toEqual({ tab: 'symptoms' })
  })

  it('filter helpers derive distinct crop and category values', () => {
    const rows = [
      { crop_keys: ['lettuce', 'Tomato'], categories: ['Deficiency'] },
      { crop_keys: ['tomato'], categories: ['pest'] },
    ]
    expect(uniqueCropKeys(rows)).toEqual(['lettuce', 'tomato', 'Tomato'])
    expect(uniqueCategories(rows)).toEqual(['Deficiency', 'pest'])
  })

  it('SymptomGuide renders dropdowns and applies deep-link query params', async () => {
    await router.push('/symptom-guide?crop_key=lettuce&category=deficiency')
    const wrapper = mount(SymptomGuide, {
      global: { plugins: [router] },
    })
    await flushPromises()

    expect(wrapper.find('[data-test="symptom-crop-select"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="symptom-category-select"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="symptom-guide-list"]').exists()).toBe(true)
    expect(api.get).toHaveBeenCalledWith('/commons/agronomy-symptoms', {
      params: { crop_key: 'lettuce', category: 'deficiency' },
    })
  })
})
