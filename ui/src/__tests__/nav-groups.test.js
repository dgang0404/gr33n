import { describe, it, expect } from 'vitest'
import { buildNavGroups, mobileBottomNav } from '../lib/navGroups.js'

describe('Phase 68 — workspace nav groups', () => {
  const groups = buildNavGroups()

  it('uses workspace-first grow & operate labels', () => {
    const grow = groups.find((g) => g.label === 'Grow & operate')
    expect(grow.items.some((i) => i.label === 'My zones' && i.to === '/zones')).toBe(true)
    expect(grow.items.some((i) => i.label === 'Feed & water' && i.to === '/feed-water')).toBe(true)
    expect(grow.items.some((i) => i.label === 'Hardware' && i.to === '/hardware')).toBe(true)
    expect(grow.items.some((i) => i.label === 'Money' && i.to === '/money')).toBe(true)
    expect(grow.items.some((i) => i.to === '/feeding')).toBe(false)
  })

  it('comfort & automation is a single workspace entry (Phase 75)', () => {
    const grow = groups.find((g) => g.label === 'Grow & operate')
    expect(grow.items.some((i) => i.to === '/comfort-targets' && i.label === 'Comfort & automation')).toBe(true)
    expect(groups.some((g) => g.label === 'Advanced')).toBe(false)
    expect(grow.items.some((i) => i.to === '/schedules')).toBe(false)
  })

  it('groups Today cockpit items', () => {
    const today = groups.find((g) => g.label === 'Today')
    expect(today.items.some((i) => i.label === 'Today' && i.to === '/')).toBe(true)
    expect(today.items.some((i) => i.to === '/alerts')).toBe(false)
    expect(today.items.some((i) => i.to === '/tasks')).toBe(false)
  })

  it('More holds Help workspace instead of scattered reference pages', () => {
    const more = groups.find((g) => g.label === 'More')
    expect(more.items.some((i) => i.label === 'Help' && i.to === '/operator-guide')).toBe(true)
    expect(more.items.some((i) => i.to === '/chat')).toBe(false)
    expect(more.items.some((i) => i.to === '/farm-knowledge')).toBe(false)
    expect(more.items.some((i) => i.to === '/catalog')).toBe(false)
    expect(more.items.some((i) => i.to.includes('crop-cycles/compare'))).toBe(false)
  })

  it('uses farmer labels on mobile bottom nav with feed and targets workspaces', () => {
    expect(mobileBottomNav.find((i) => i.to === '/')?.label).toBe('Today')
    expect(mobileBottomNav.find((i) => i.to === '/zones')?.label).toBe('Zones')
    expect(mobileBottomNav.find((i) => i.to === '/feed-water')?.label).toBe('Feed')
    expect(mobileBottomNav.find((i) => i.to === '/comfort-targets')?.label).toBe('Targets')
    expect(mobileBottomNav.find((i) => i.to === '/alerts')).toBeUndefined()
  })
})
