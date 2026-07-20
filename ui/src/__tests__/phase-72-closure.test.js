/**
 * Phase 72 — Money unification closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { buildLegacyRedirectRoutes } from '../lib/workspaces.js'
import router from '../router/index.js'

describe('Phase 72 — money closure', () => {
  it('legacy money routes redirect into /money workspace tabs', () => {
    const paths = {
      '/operations/money': 'summary',
      '/costs': 'ledger',
      '/operations/supplies': 'supplies',
    }
    for (const [legacy, tab] of Object.entries(paths)) {
      const entry = buildLegacyRedirectRoutes().find((r) => r.path === legacy)
      expect(entry, legacy).toBeTruthy()
      const result = entry.redirect({
        path: legacy,
        query: {},
        hash: '',
        fullPath: legacy,
        matched: [],
        meta: {},
        name: undefined,
        params: {},
        redirectedFrom: undefined,
      })
      expect(result.path).toBe('/money')
      expect(result.query.tab).toBe(tab)
    }
  })

  it('/inventory redirects to natural-farming studio tabs', () => {
    const entry = buildLegacyRedirectRoutes().find((r) => r.path === '/inventory')
    expect(entry).toBeTruthy()
    const recipes = entry.redirect({
      path: '/inventory',
      query: {},
      hash: '',
      fullPath: '/inventory',
      matched: [],
      meta: {},
      name: undefined,
      params: {},
      redirectedFrom: undefined,
    })
    expect(recipes.path).toBe('/natural-farming')
    expect(recipes.query.tab).toBe('recipes')

    const stock = entry.redirect({
      path: '/inventory',
      query: { inv: 'batches' },
      hash: '',
      fullPath: '/inventory?inv=batches',
      matched: [],
      meta: {},
      name: undefined,
      params: {},
      redirectedFrom: undefined,
    })
    expect(stock.path).toBe('/natural-farming')
    expect(stock.query.tab).toBe('stock')
  })

  it('MoneyHub footer links to ledger tab not orphan /costs', () => {
    const money = readFileSync(join(process.cwd(), 'src/views/MoneyHub.vue'), 'utf8')
    expect(money).toContain("tab: 'ledger'")
    expect(money).not.toContain('to="/costs"')
  })

  it('SuppliesHub explains unit costs feed monthly spend', () => {
    const supplies = readFileSync(join(process.cwd(), 'src/views/SuppliesHub.vue'), 'utf8')
    expect(supplies).toContain('Unit costs here feed into')
    expect(supplies).toContain("tab: 'summary'")
  })

  it('registers /money workspace route', () => {
    expect(router.resolve('/money').name).toBe('money')
  })
})
