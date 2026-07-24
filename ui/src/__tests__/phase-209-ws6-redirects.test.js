/**
 * Phase 209 WS6 — redirects and nav wiring.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { buildLegacyRedirectRoutes, WORKSPACES } from '../lib/workspaces.js'
import { naturalFarmingTabRoute, redirectLegacyInventory } from '../lib/workspaceRoutes.js'
import { relatedTo } from '../lib/navRelations.js'

const fertigation = readFileSync(join(process.cwd(), 'src/views/Fertigation.vue'), 'utf8')

describe('Phase 209 WS6 — redirects & nav', () => {
  it('legacy /inventory maps to natural-farming recipes or Money supplies', () => {
    expect(WORKSPACES.naturalfarming.absorbs?.['/inventory']).toEqual({ tab: 'recipes' })
    expect(WORKSPACES.money.absorbs?.['/inventory']).toBeUndefined()

    const entry = buildLegacyRedirectRoutes().find((r) => r.path === '/inventory')
    expect(entry).toBeTruthy()
    expect(redirectLegacyInventory({ query: {} }).path).toBe('/natural-farming')
    expect(redirectLegacyInventory({ query: { inv: 'batches' } }).path).toBe('/money')
    expect(redirectLegacyInventory({ query: { inv: 'definitions' } }).path).toBe('/money')
  })

  it('Fertigation inventory links target Money supplies and apply recipes', () => {
    expect(fertigation).toContain('naturalFarmingTabRoute')
    expect(fertigation).toContain('naturalFarmingManageRoute')
    expect(fertigation).toContain('batchStockLink')
    expect(fertigation).toContain('recipeLink')
    expect(fertigation).toContain('Inventory batches')
  })

  it('nav relations include natural-farming for /inventory', () => {
    expect(relatedTo('/inventory')).toContain('/natural-farming')
    expect(relatedTo('/inventory')).toContain('/money')
  })

  it('naturalFarmingTabRoute builds studio deep links', () => {
    expect(naturalFarmingTabRoute('recipes', { recipe: 7 })).toEqual({
      path: '/natural-farming',
      query: { tab: 'recipes', recipe: '7' },
    })
    expect(naturalFarmingTabRoute('stock', { batchId: 12 })).toEqual({
      path: '/money',
      query: { tab: 'supplies', batch_id: '12' },
    })
  })

  it('money workspace has no advanced inventory tab', () => {
    const tab = WORKSPACES.money.tabs.find((t) => t.id === 'inventory')
    expect(tab).toBeUndefined()
    expect(WORKSPACES.naturalfarming.tabs.some((t) => t.id === 'manage')).toBe(false)
  })
})
