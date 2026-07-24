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

  it('sends batches and batch_id to Natural farming manage tab', () => {
    expect(redirectLegacyInventory({ query: { inv: 'batches' } })).toEqual({
      path: '/natural-farming',
      query: { tab: 'manage', inv: 'batches' },
    })
    expect(redirectLegacyInventory({ query: { batch_id: '9' } })).toEqual({
      path: '/natural-farming',
      query: { tab: 'manage', inv: 'batches', batch_id: '9' },
    })
  })

  it('sends definitions to Natural farming manage tab', () => {
    expect(redirectLegacyInventory({ query: { inv: 'definitions' } })).toEqual({
      path: '/natural-farming',
      query: { tab: 'manage', inv: 'definitions' },
    })
  })
})
