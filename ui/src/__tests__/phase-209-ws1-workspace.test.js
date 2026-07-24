/**
 * Phase 209 WS1 — Natural farming workspace shell.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { WORKSPACES, resolveWorkspaceTab, workspaceByRoute } from '../lib/workspaces.js'
import { buildNavGroups } from '../lib/navGroups.js'
import router from '../router/index.js'

describe('Phase 209 WS1 — natural farming workspace', () => {
  it('declares three tabs; batch is default', () => {
    const tabs = WORKSPACES.naturalfarming.tabs.map((t) => t.id)
    expect(tabs).toEqual(['batch', 'library', 'recipes', 'manage'])
    expect(resolveWorkspaceTab('naturalfarming', undefined)).toBe('batch')
    expect(resolveWorkspaceTab('naturalfarming', 'start')).toBe('batch')
    expect(resolveWorkspaceTab('naturalfarming', 'bogus')).toBe('batch')
    expect(resolveWorkspaceTab('naturalfarming', 'recipes')).toBe('recipes')
  })

  it('registers /natural-farming route and workspace metadata', () => {
    const resolved = router.resolve({ path: '/natural-farming' })
    expect(resolved.name).toBe('natural-farming')
    const ws = workspaceByRoute('/natural-farming')
    expect(ws?.label).toBe('Natural farming')
    expect(ws?.route).toBe('/natural-farming')
  })

  it('deep-links ?tab=library on /natural-farming', () => {
    const resolved = router.resolve({ path: '/natural-farming', query: { tab: 'library' } })
    expect(resolved.name).toBe('natural-farming')
    expect(resolved.query.tab).toBe('library')
  })

  it('NaturalFarmingWorkspace hosts three operational panels', () => {
    const src = readFileSync(
      join(process.cwd(), 'src/views/workspaces/NaturalFarmingWorkspace.vue'),
      'utf8',
    )
    expect(src).toContain('workspace-id="naturalfarming"')
    expect(src).not.toContain('SwitchoverWizard')
    expect(src).toContain("activeTab === 'batch'")
    expect(src).toContain('MakeBatchPanel')
    expect(src).toContain("activeTab === 'library'")
    expect(src).toContain('RecipeLibraryPanel')
    expect(src).toContain("activeTab === 'recipes'")
    expect(src).toContain('RecipesApplyPanel')
    expect(src).toContain('FarmRowsPanel')
  })

  it('sidebar lists Natural farming under Grow & operate', () => {
    const grow = buildNavGroups().find((g) => g.label === 'Grow & operate')
    const idxZones = grow.items.findIndex((i) => i.to === '/zones')
    const idxNf = grow.items.findIndex((i) => i.to === '/natural-farming')
    const idxComfort = grow.items.findIndex((i) => i.to === '/comfort-targets')
    expect(idxNf).toBeGreaterThan(idxZones)
    expect(idxNf).toBeLessThan(idxComfort)
    expect(grow.items[idxNf].label).toBe('Natural farming')
  })
})
