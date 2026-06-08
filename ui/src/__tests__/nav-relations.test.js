import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { NAV_RELATIONS, relatedTo } from '../lib/navRelations.js'
import { buildNavGroups, collectSidebarRoutes } from '../lib/navGroups.js'

describe('Phase 49 WS2 — nav relations', () => {
  const navRoutes = new Set(collectSidebarRoutes(buildNavGroups('/farms/1/crop-cycles/compare')))

  it('returns related routes for grow-path siblings', () => {
    expect(relatedTo('/zones')).toEqual(['/feeding', '/comfort-targets', '/plants'])
    expect(relatedTo('/plants')).toEqual(['/zones', '/comfort-targets'])
    expect(relatedTo('/feeding')).toEqual(['/zones', '/comfort-targets', '/plants'])
    expect(relatedTo('/comfort-targets')).toContain('/zones')
    expect(relatedTo('/comfort-targets')).toContain('/plants')
  })

  it('links tasks, alerts, and fertigation across the grow story', () => {
    expect(relatedTo('/tasks')).toContain('/zones')
    expect(relatedTo('/alerts')).toContain('/zones')
    expect(relatedTo('/fertigation')).toContain('/feeding')
    expect(relatedTo('/fertigation')).toContain('/operations/feeding')
  })

  it('returns empty for unknown routes', () => {
    expect(relatedTo('/chat')).toEqual([])
    expect(relatedTo(null)).toEqual([])
  })

  it('only points at routes that exist in the sidebar', () => {
    for (const [from, targets] of Object.entries(NAV_RELATIONS)) {
      expect(navRoutes.has(from), `missing nav route ${from}`).toBe(true)
      for (const to of targets) {
        expect(navRoutes.has(to), `${from} → ${to} not in sidebar`).toBe(true)
      }
    }
  })
})
