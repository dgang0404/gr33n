import { describe, it, expect, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { resolveHintPath } from '../directives/navHint.js'
import { useNavHighlightStore } from '../stores/navHighlight.js'
import { buildNavGroups, collectSidebarRoutes } from '../lib/navGroups.js'

describe('Phase 49 / 71 WS5 — nav-hint path resolution', () => {
  it('resolves a string path and maps legacy routes to feed-water workspace', () => {
    expect(resolveHintPath('/feeding')).toBe('/feed-water')
    expect(resolveHintPath('/fertigation?tab=programs')).toBe('/feed-water')
  })

  it('resolves a router object path and maps to workspace sidebar routes', () => {
    expect(resolveHintPath({ path: '/feeding', query: { zone_id: '2' } })).toBe('/feed-water')
    expect(resolveHintPath({ path: '/operations/feeding', query: { tab: 'programs' } })).toBe(
      '/feed-water',
    )
  })

  it('maps legacy operations paths to workspace sidebar routes', () => {
    expect(resolveHintPath('/operations/supplies')).toBe('/money')
    expect(resolveHintPath('/operations/feeding')).toBe('/feed-water')
    expect(resolveHintPath('/natural-farming?tab=recipes')).toBe('/natural-farming')
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

describe('Phase 78 — sidebar hint targets base tabs only', () => {
  const sidebarRoutes = new Set(collectSidebarRoutes(buildNavGroups()))

  it('resolved hint paths are top-level sidebar routes when possible', () => {
    for (const path of ['/zones', '/money', '/comfort-targets', '/feed-water', '/sensors', '/plants']) {
      const resolved = resolveHintPath(path)
      expect(resolved).toBeTruthy()
      expect(sidebarRoutes.has(resolved)).toBe(true)
    }
  })
})
