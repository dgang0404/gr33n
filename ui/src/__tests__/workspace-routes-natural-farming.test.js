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

  it('sends batches and batch_id to Money supplies', () => {
    expect(redirectLegacyInventory({ query: { inv: 'batches' } })).toEqual({
      path: '/money',
      query: { tab: 'supplies' },
    })
    expect(redirectLegacyInventory({ query: { batch_id: '9' } })).toEqual({
      path: '/money',
      query: { tab: 'supplies', batch_id: '9' },
    })
  })

  it('sends definitions to Money supplies', () => {
    expect(redirectLegacyInventory({ query: { inv: 'definitions' } })).toEqual({
      path: '/money',
      query: { tab: 'supplies' },
    })
  })
})
