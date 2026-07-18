/**
 * Phase 180 WS1 — Help knowledge surfaces map.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createRouter, createMemoryHistory } from 'vue-router'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import HelpKnowledgeSurfacesMap from '../components/HelpKnowledgeSurfacesMap.vue'
import SymptomGuide from '../views/SymptomGuide.vue'
import { WORKSPACES } from '../lib/workspaces.js'
import { uniqueCropKeys, uniqueCategories } from '../lib/symptomGuideFilters.js'

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
    { path: '/chat', component: { template: '<div/>' } },
  ],
})

describe('Phase 180 WS1 — help knowledge surfaces map', () => {
  it('renders three surface cards with correct links (Guide card removed)', async () => {
    const wrapper = mount(HelpKnowledgeSurfacesMap, {
      global: { plugins: [router] },
    })

    expect(wrapper.find('[data-test="help-knowledge-surfaces-map"]').exists()).toBe(true)

    const guide = wrapper.find('[data-test="help-surface-card-guide"]')
    const knowledge = wrapper.find('[data-test="help-surface-card-knowledge"]')
    const catalog = wrapper.find('[data-test="help-surface-card-catalog"]')
    const symptoms = wrapper.find('[data-test="help-surface-card-symptoms"]')

    expect(guide.exists()).toBe(false)
    expect(knowledge.exists()).toBe(true)
    expect(catalog.exists()).toBe(true)
    expect(symptoms.exists()).toBe(true)

    expect(knowledge.attributes('href')).toContain('/operator-guide')
    expect(knowledge.attributes('href')).toContain('tab=knowledge')
    expect(catalog.attributes('href')).toContain('tab=catalog')
    expect(symptoms.attributes('href')).toContain('tab=symptoms')

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

  it('help workspace includes symptoms as its own tab', () => {
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
    await router.push('/operator-guide?tab=symptoms&crop_key=lettuce&category=deficiency')
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

describe('Phase 180 WS3 — knowledge search simplification', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    api.get.mockResolvedValue({ data: { ai_enabled: true } })
    api.post.mockResolvedValue({ data: { results: [] } })
  })

  it('hides advanced filters by default and shows semantic hint', async () => {
    const FarmKnowledge = (await import('../views/FarmKnowledge.vue')).default
    const { useFarmContextStore } = await import('../stores/farmContext')
    const { useCapabilitiesStore } = await import('../stores/capabilities')
    useFarmContextStore().farmId = 1
    useCapabilitiesStore().loaded = true
    useCapabilitiesStore().isLite = false

    const wrapper = mount(FarmKnowledge, {
      props: { embedded: true },
      global: { plugins: [router] },
    })
    await flushPromises()

    expect(wrapper.find('[data-test="farm-knowledge-semantic-hint"]').text()).toMatch(/plain language/i)
    expect(wrapper.find('[data-test="farm-knowledge-advanced"]').exists()).toBe(false)
    expect(wrapper.find('[data-test="farm-knowledge-examples"]').exists()).toBe(true)

    await wrapper.find('[data-test="farm-knowledge-advanced-toggle"]').trigger('click')
    expect(wrapper.find('[data-test="farm-knowledge-advanced"]').exists()).toBe(true)
  })
})

describe('Phase 180 WS4 — field guide browse list', () => {
  const sampleGuides = [
    {
      id: 1,
      slug: 'crop-lettuce-nutrition',
      title: 'Lettuce nutrition',
      crop_key: 'lettuce',
      guide_kind: 'nutrition',
      safety_tier: 'informational',
      catalog_version: 1,
      sort_order: 10,
    },
    {
      id: 2,
      slug: 'crop-tomato-nutrition',
      title: 'Tomato nutrition',
      crop_key: 'tomato',
      guide_kind: 'nutrition',
      safety_tier: 'informational',
      catalog_version: 1,
      sort_order: 20,
    },
  ]

  beforeEach(() => {
    vi.clearAllMocks()
    api.get.mockImplementation((url) => {
      if (url === '/commons/agronomy-field-guides') {
        return Promise.resolve({ data: sampleGuides })
      }
      if (url === '/commons/agronomy-field-guides/crop-lettuce-nutrition') {
        return Promise.resolve({
          data: {
            ...sampleGuides[0],
            body_md: '## Lettuce\nFeed EC 1.2–1.6.',
          },
        })
      }
      return Promise.resolve({ data: {} })
    })
  })

  it('FieldGuideBrowse lists guides and filters by crop', async () => {
    const FieldGuideBrowse = (await import('../components/FieldGuideBrowse.vue')).default
    const wrapper = mount(FieldGuideBrowse, {
      global: { plugins: [router] },
    })
    await flushPromises()

    expect(wrapper.find('[data-test="field-guide-browse"]').exists()).toBe(true)
    expect(wrapper.findAll('[data-test^="field-guide-row-"]')).toHaveLength(2)

    await wrapper.find('[data-test="field-guide-crop-filter"]').setValue('lettuce')
    expect(wrapper.findAll('[data-test^="field-guide-row-"]')).toHaveLength(1)
    expect(wrapper.find('[data-test="field-guide-row-crop-lettuce-nutrition"]').exists()).toBe(true)
  })

  it('selecting a guide loads detail and open-indexed-doc action', async () => {
    await router.push('/operator-guide?tab=knowledge')
    const FieldGuideBrowse = (await import('../components/FieldGuideBrowse.vue')).default
    const wrapper = mount(FieldGuideBrowse, {
      global: { plugins: [router] },
    })
    await flushPromises()

    await wrapper.find('[data-test="field-guide-row-crop-lettuce-nutrition"]').trigger('click')
    await flushPromises()

    expect(wrapper.find('[data-test="field-guide-detail"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="field-guide-body"]').text()).toMatch(/Lettuce/)

    await wrapper.find('[data-test="field-guide-open-doc"]').trigger('click')
    await flushPromises()
    expect(router.currentRoute.value.path).toBe('/operator-guide')
    expect(router.currentRoute.value.query.cited_doc).toBe('field-guides/crop-lettuce-nutrition.md')
  })

  it('FarmKnowledge embeds field guide browse section', async () => {
    setActivePinia(createPinia())
    const FarmKnowledge = (await import('../views/FarmKnowledge.vue')).default
    const { useFarmContextStore } = await import('../stores/farmContext')
    const { useCapabilitiesStore } = await import('../stores/capabilities')
    useFarmContextStore().farmId = 1
    useCapabilitiesStore().loaded = true
    useCapabilitiesStore().isLite = false

    api.get.mockImplementation((url) => {
      if (url === '/commons/agronomy-field-guides') {
        return Promise.resolve({ data: sampleGuides })
      }
      if (url.includes('/ai/status') || url.includes('/capabilities')) {
        return Promise.resolve({ data: { ai_enabled: true } })
      }
      return Promise.resolve({ data: {} })
    })

    const wrapper = mount(FarmKnowledge, {
      props: { embedded: true },
      global: { plugins: [router] },
    })
    await flushPromises()

    expect(wrapper.find('[data-test="field-guide-browse"]').exists()).toBe(true)
  })
})

describe('Phase 180 WS5 — citation doc view round-trip', () => {
  const sampleChunks = [
    {
      id: 101,
      chunk_index: 0,
      content_text: 'field_guide\ndoc_path: field-guides/crop-lettuce-nutrition.md\n\n## Lettuce\nFeed EC 1.2–1.6.',
    },
    {
      id: 102,
      chunk_index: 1,
      content_text: 'field_guide\ndoc_path: field-guides/crop-lettuce-nutrition.md\n\nWatch for tip burn in summer.',
    },
  ]

  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    api.get.mockImplementation((url, config) => {
      if (url.includes('/rag/docs')) {
        return Promise.resolve({
          data: {
            doc_path: config?.params?.doc_path,
            chunks: sampleChunks,
          },
        })
      }
      if (url === '/commons/agronomy-field-guides/crop-lettuce-nutrition') {
        return Promise.resolve({ data: { title: 'Lettuce nutrition' } })
      }
      return Promise.resolve({ data: {} })
    })
  })

  it('CitationDocView renders chunks and highlights cited section', async () => {
    const CitationDocView = (await import('../components/CitationDocView.vue')).default
    const { useFarmContextStore } = await import('../stores/farmContext')
    useFarmContextStore().farmId = 1

    const wrapper = mount(CitationDocView, {
      props: {
        docPath: 'field-guides/crop-lettuce-nutrition.md',
        docType: 'field_guide',
        highlightChunkId: 102,
      },
      global: { plugins: [router] },
    })
    await flushPromises()

    expect(wrapper.find('[data-test="citation-doc-view"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="citation-doc-title"]').text()).toBe('Lettuce nutrition')
    expect(wrapper.findAll('[data-test^="citation-doc-chunk-"]')).toHaveLength(2)
    expect(wrapper.find('[data-test="citation-doc-chunk-highlight"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="citation-doc-chunk-highlight"]').text()).toMatch(/tip burn/i)
  })

  it('CitationDocView Ask Guardian opens slide-out drawer instead of navigating to /chat', async () => {
    const CitationDocView = (await import('../components/CitationDocView.vue')).default
    const { useFarmContextStore } = await import('../stores/farmContext')
    const { useGuardianPanelStore } = await import('../stores/guardianPanel')
    useFarmContextStore().farmId = 1
    const panel = useGuardianPanelStore()

    const wrapper = mount(CitationDocView, {
      props: {
        docPath: 'field-guides/crop-lettuce-nutrition.md',
        docType: 'field_guide',
      },
      global: { plugins: [router] },
    })
    await flushPromises()

    await wrapper.find('[data-test="citation-doc-ask-guardian"]').trigger('click')
    expect(panel.open).toBe(true)
    expect(panel.drawerTab).toBe('chat')
    expect(panel.prefilledMessage).toMatch(/Lettuce nutrition/)
    expect(router.currentRoute.value.path).not.toBe('/chat')
  })

  it('FarmKnowledge shows doc view instead of cited-doc banner when deep-linked', async () => {
    await router.push('/operator-guide?tab=knowledge&cited_doc=field-guides/crop-lettuce-nutrition.md&cited_chunk=102')
    const FarmKnowledge = (await import('../views/FarmKnowledge.vue')).default
    const { useFarmContextStore } = await import('../stores/farmContext')
    const { useCapabilitiesStore } = await import('../stores/capabilities')
    useFarmContextStore().farmId = 1
    useCapabilitiesStore().loaded = true
    useCapabilitiesStore().isLite = false

    const wrapper = mount(FarmKnowledge, {
      props: { embedded: true },
      global: { plugins: [router] },
    })
    await flushPromises()

    expect(wrapper.find('[data-test="citation-doc-view"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="farm-knowledge-cited-doc"]').exists()).toBe(false)
  })

  it('citationDoc helpers strip ingest headers and build Guardian prefill', async () => {
    const { chunkDisplayText, guardianDocPrefill } = await import('../lib/citationDoc.js')
    const text = chunkDisplayText(sampleChunks[0].content_text)
    expect(text).toMatch(/^## Lettuce/)
    expect(text).not.toMatch(/doc_path:/)
    expect(guardianDocPrefill('Lettuce nutrition')).toMatch(/Lettuce nutrition/)
  })
})

describe('Phase 180 WS6 — closure docs and nav', () => {
  const repoRoot = join(process.cwd(), '..')
  const repoDocs = join(repoRoot, 'docs')

  it('SymptomGuide dropdowns populate from mocked catalog API', async () => {
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
            body_md: 'Edge necrosis.',
          },
          {
            id: 2,
            symptom_key: 'tomato-blossom',
            display_name: 'Blossom end rot',
            crop_keys: ['tomato'],
            categories: ['pest'],
            body_md: 'Fruit lesion.',
          },
        ],
      },
    })

    await router.push('/operator-guide?tab=symptoms')
    const wrapper = mount(SymptomGuide, {
      props: { embedded: true },
      global: { plugins: [router] },
    })
    await flushPromises()

    const cropSelect = wrapper.find('[data-test="symptom-crop-select"]')
    const categorySelect = wrapper.find('[data-test="symptom-category-select"]')
    expect(cropSelect.findAll('option')).toHaveLength(3)
    expect(categorySelect.findAll('option')).toHaveLength(3)
    expect(cropSelect.text()).toMatch(/lettuce/)
    expect(cropSelect.text()).toMatch(/tomato/)
    expect(categorySelect.text()).toMatch(/deficiency/)
    expect(categorySelect.text()).toMatch(/pest/)
  })

  it('OperatorGuide embeds the knowledge surfaces map', async () => {
    const OperatorGuide = (await import('../views/OperatorGuide.vue')).default
    const wrapper = mount(OperatorGuide, {
      props: { embedded: true },
      global: { plugins: [router] },
    })
    await flushPromises()
    expect(wrapper.find('[data-test="help-knowledge-surfaces-map"]').exists()).toBe(true)
  })

  it('plan and operator-tour document Phase 180 shipped', () => {
    const plan = readFileSync(
      join(repoDocs, 'plans/archive/phase_180_knowledge_surfaces_discoverability.plan.md'),
      'utf8',
    )
    const tour = readFileSync(join(repoDocs, 'operator-tour.md'), 'utf8')
    const routes = readFileSync(join(repoRoot, 'cmd/api/routes.go'), 'utf8')

    expect(plan).toContain('**Status:** shipped')
    expect(plan).toMatch(/- \[x\] Symptom guide reachable/)
    expect(tour).toMatch(/7m\. Help knowledge surfaces \(Phase 180/i)
    expect(tour).toMatch(/What lives where/)
    expect(tour).toMatch(/tab=library/)
    expect(tour).toMatch(/section=symptoms/)
    expect(tour).toMatch(/semantic search/i)
    expect(routes).toContain('/farms/{id}/rag/docs')
  })
})
