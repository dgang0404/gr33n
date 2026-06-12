/**
 * Phase 72 — Money workspace tabs.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { WORKSPACES, resolveWorkspaceTab } from '../lib/workspaces.js'
import router from '../router/index.js'

describe('Phase 72 — money tabs', () => {
  it('declares summary, ledger, supplies, inventory, and grows tabs', () => {
    const tabs = WORKSPACES.money.tabs.map((t) => t.id)
    expect(tabs).toEqual(['summary', 'ledger', 'supplies', 'inventory', 'grows'])
  })

  it('defaults to summary (This month)', () => {
    expect(resolveWorkspaceTab('money', undefined)).toBe('summary')
    expect(resolveWorkspaceTab('money', 'bogus')).toBe('summary')
  })

  it('MoneyWorkspace hosts MoneyHub, Costs, SuppliesHub, and Inventory', () => {
    const src = readFileSync(join(process.cwd(), 'src/views/workspaces/MoneyWorkspace.vue'), 'utf8')
    expect(src).toContain("activeTab === 'summary'")
    expect(src).toContain("activeTab === 'ledger'")
    expect(src).toContain("activeTab === 'supplies'")
    expect(src).toContain("activeTab === 'inventory'")
    expect(src).toContain('MoneyHub')
    expect(src).toContain('Costs')
    expect(src).toContain('SuppliesHub')
    expect(src).toContain('Inventory')
  })

  it('deep-links ?tab=ledger on /money', () => {
    const resolved = router.resolve({ path: '/money', query: { tab: 'ledger' } })
    expect(resolved.name).toBe('money')
    expect(resolved.query.tab).toBe('ledger')
  })
})
