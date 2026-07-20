/**
 * Phase 209 WS6 — legacy inventory redirect helper.
 */
import { describe, it, expect } from 'vitest'
import { redirectLegacyInventory } from '../lib/workspaceRoutes.js'

describe('redirectLegacyInventory', () => {
  it('defaults to recipes tab', () => {
    expect(redirectLegacyInventory({ query: {} })).toEqual({
      path: '/natural-farming',
      query: { tab: 'recipes' },
    })
  })

  it('sends batches and batch_id to stock tab', () => {
    expect(redirectLegacyInventory({ query: { inv: 'batches' } }).query.tab).toBe('stock')
    expect(redirectLegacyInventory({ query: { batch_id: '9' } }).query).toMatchObject({
      tab: 'stock',
      batch_id: '9',
    })
  })

  it('keeps input definitions on Money advanced tab', () => {
    expect(redirectLegacyInventory({ query: { inv: 'definitions' } })).toEqual({
      path: '/money',
      query: { tab: 'inventory', inv: 'definitions' },
    })
  })
})
