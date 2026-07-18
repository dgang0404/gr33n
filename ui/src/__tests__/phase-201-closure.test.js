/**
 * Phase 201 — Help knowledge surfaces unification (Search + Import tabs).
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createRouter, createMemoryHistory } from 'vue-router'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import HelpWorkspace from '../views/workspaces/HelpWorkspace.vue'
import { WORKSPACES, buildLegacyRedirectRoutes } from '../lib/workspaces.js'

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

const repoRoot = join(process.cwd(), '..')

const router = createRouter({
  history: createMemoryHistory(),
  routes: [
    { path: '/operator-guide', component: HelpWorkspace },
    ...buildLegacyRedirectRoutes(),
  ],
})

describe('Phase 201 — Help knowledge surfaces unification', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    api.get.mockResolvedValue({ data: { symptoms: [], ai_enabled: true } })
  })

  it('Help workspace has Search and Import tabs', () => {
    expect(WORKSPACES.help.tabs.map((t) => t.id)).toEqual([
      'library', 'pi-setup', 'knowledge', 'symptoms', 'catalog',
    ])
    expect(WORKSPACES.help.absorbs['/farm-knowledge']).toEqual({ tab: 'knowledge' })
    expect(WORKSPACES.help.absorbs['/catalog']).toEqual({ tab: 'catalog' })
  })

  it('router has no standalone /farm-knowledge or /catalog pages', () => {
    const routerSrc = readFileSync(join(repoRoot, 'ui/src/router/index.js'), 'utf8')
    expect(routerSrc).not.toMatch(/path: '\/farm-knowledge'/)
    expect(routerSrc).not.toMatch(/path: '\/catalog'/)
  })

  it('/farm-knowledge redirects to Help Search tab', async () => {
    await router.push('/farm-knowledge?cited_doc=field-guides/foo.md')
    await flushPromises()
    expect(router.currentRoute.value.path).toBe('/operator-guide')
    expect(router.currentRoute.value.query.tab).toBe('knowledge')
    expect(router.currentRoute.value.query.cited_doc).toBe('field-guides/foo.md')
  })

  it('/catalog redirects to Help Import tab', async () => {
    await router.push('/catalog')
    await flushPromises()
    expect(router.currentRoute.value.path).toBe('/operator-guide')
    expect(router.currentRoute.value.query.tab).toBe('catalog')
  })

  it('HelpLibraryHub is how-to only (no embedded knowledge/catalog)', () => {
    const hub = readFileSync(join(repoRoot, 'ui/src/views/HelpLibraryHub.vue'), 'utf8')
    expect(hub).not.toContain('FarmKnowledge')
    expect(hub).not.toContain('CommonsCatalog')
    expect(hub).not.toContain('help-library-section-knowledge')
    expect(hub).not.toContain('help-library-section-catalog')
  })

  it('renders FarmKnowledge on knowledge tab', async () => {
    const { useFarmContextStore } = await import('../stores/farmContext.js')
    const { useCapabilitiesStore } = await import('../stores/capabilities.js')
    useFarmContextStore().farmId = 1
    useCapabilitiesStore().loaded = true
    useCapabilitiesStore().isLite = false

    await router.push('/operator-guide?tab=knowledge')
    const wrapper = mount(HelpWorkspace, { global: { plugins: [router] } })
    await flushPromises()
    expect(wrapper.find('[data-test="farm-knowledge-search"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="help-library-hub"]').exists()).toBe(false)
  })

  it('renders CommonsCatalog on catalog tab', async () => {
    await router.push('/operator-guide?tab=catalog')
    const wrapper = mount(HelpWorkspace, { global: { plugins: [router] } })
    await flushPromises()
    expect(wrapper.text()).toMatch(/Browse Catalog|Farm Imports/i)
    expect(wrapper.find('[data-test="help-library-hub"]').exists()).toBe(false)
  })

  it('legacy ?section=knowledge redirects to knowledge tab', async () => {
    await router.push('/operator-guide?tab=library&section=knowledge')
    await flushPromises()
    expect(router.currentRoute.value.query.tab).toBe('knowledge')
    expect(router.currentRoute.value.query.section).toBeUndefined()
  })
})
