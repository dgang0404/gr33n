/**
 * Phase 199 — Help workspace sticky chrome consolidation.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { createPinia, setActivePinia } from 'pinia'
import { createRouter, createMemoryHistory } from 'vue-router'
import HelpWorkspace from '../views/workspaces/HelpWorkspace.vue'

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
  routes: [{ path: '/operator-guide', component: HelpWorkspace }],
})

describe('Phase 199 — Help sticky chrome consolidation', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    api.get.mockResolvedValue({ data: { symptoms: [], ai_enabled: true } })
  })

  it('library section pills live inside WorkspaceShell subnav, not HelpLibraryHub', () => {
    const hub = readFileSync(join(repoRoot, 'ui/src/views/HelpLibraryHub.vue'), 'utf8')
    expect(hub).not.toContain('sticky')
    expect(hub).not.toContain('help-library-jump')
    expect(hub).toContain('scroll-mt-4')

    const shell = readFileSync(join(repoRoot, 'ui/src/components/WorkspaceShell.vue'), 'utf8')
    expect(shell).toContain('subnav-extra')
    expect(shell).toContain('unifiedHeader')
    expect(shell).toContain('workspace-shell-unified-chrome')
    expect(shell).toContain('workspace-shell__content flex-1 min-h-0 overflow-y-auto')

    const nav = readFileSync(join(repoRoot, 'ui/src/components/HelpLibrarySectionNav.vue'), 'utf8')
    expect(nav).toContain('data-test="help-library-jump"')
    expect(nav).not.toContain('sticky')
  })

  it('Help workspace renders section pills in subnav-extra on library tab', async () => {
    await router.push('/operator-guide?tab=library')
    const wrapper = mount(HelpWorkspace, { global: { plugins: [router] } })
    await flushPromises()

    expect(wrapper.find('[data-test="workspace-shell-unified-chrome"]').exists()).toBe(true)
    const subnav = wrapper.find('[data-test="workspace-shell-subnav-extra"]')
    expect(subnav.exists()).toBe(true)
    expect(subnav.find('[data-test="help-library-jump"]').exists()).toBe(true)
    expect(subnav.find('[data-test="help-library-jump-symptoms"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="help-library-hub"]').exists()).toBe(true)
  })

  it('hides library section pills on pi-setup tab', async () => {
    await router.push('/operator-guide?tab=pi-setup')
    const wrapper = mount(HelpWorkspace, { global: { plugins: [router] } })
    await flushPromises()

    expect(wrapper.find('[data-test="workspace-shell-subnav-extra"]').exists()).toBe(false)
    expect(wrapper.find('[data-test="help-library-jump"]').exists()).toBe(false)
  })
})
