import { describe, it, expect, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { resolveHintPath } from '../directives/navHint.js'
import { useNavHighlightStore } from '../stores/navHighlight.js'

describe('Phase 49 WS3 — nav-hint path resolution', () => {
  it('resolves a string path and maps legacy routes to workspaces', () => {
    expect(resolveHintPath('/feeding')).toBe('/feed-water')
    expect(resolveHintPath('/fertigation?tab=programs')).toBe('/feed-water')
  })

  it('resolves a router object path and maps to workspace sidebar routes', () => {
    expect(resolveHintPath({ path: '/feeding', query: { zone_id: '2' } })).toBe('/feed-water')
    expect(resolveHintPath({ path: '/operations/feeding', query: { tab: 'programs' } })).toBe(
      '/feed-water',
    )
  })

  it('returns null for empty or malformed values', () => {
    expect(resolveHintPath(null)).toBe(null)
    expect(resolveHintPath(undefined)).toBe(null)
    expect(resolveHintPath({})).toBe(null)
    expect(resolveHintPath(42)).toBe(null)
  })
})

describe('Phase 49 WS3 — navHighlight store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('sets and clears the highlighted route', () => {
    const store = useNavHighlightStore()
    expect(store.route).toBe(null)
    store.set('/fertigation')
    expect(store.route).toBe('/fertigation')
    store.clear()
    expect(store.route).toBe(null)
  })

  it('normalizes empty to null', () => {
    const store = useNavHighlightStore()
    store.set('')
    expect(store.route).toBe(null)
  })
})
