import { describe, it, expect } from 'vitest'
import { buildNavGroups, mobileBottomNav } from '../lib/navGroups.js'

describe('Phase 68 / 78 — workspace nav groups', () => {
  const groups = buildNavGroups()

  it('uses zone-first grow & operate labels', () => {
    const grow = groups.find((g) => g.label === 'Grow & operate')
    expect(grow.items.some((i) => i.label === 'My zones' && i.to === '/zones')).toBe(true)
    expect(grow.items.some((i) => i.label === 'Natural farming' && i.to === '/natural-farming')).toBe(true)
    expect(grow.items.some((i) => i.to === '/feed-water')).toBe(false)
    expect(grow.items.some((i) => i.to === '/hardware')).toBe(false)
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

  it('More includes Farm Guardian full page and Help workspace', () => {
    const more = groups.find((g) => g.label === 'More')
    expect(more.items.some((i) => i.label === 'Farm Guardian' && i.to === '/chat')).toBe(true)
    expect(more.items.some((i) => i.label === 'Help' && i.to === '/operator-guide')).toBe(true)
    expect(more.items.some((i) => i.to === '/farm-knowledge')).toBe(false)
    expect(more.items.some((i) => i.to === '/catalog')).toBe(false)
    expect(more.items.some((i) => i.to.includes('crop-cycles/compare'))).toBe(false)
  })

  it('uses farmer labels on mobile bottom nav with zones, targets, and money', () => {
    expect(mobileBottomNav.find((i) => i.to === '/')?.label).toBe('Today')
    expect(mobileBottomNav.find((i) => i.to === '/zones')?.label).toBe('Zones')
    expect(mobileBottomNav.find((i) => i.to === '/money')?.label).toBe('Money')
    expect(mobileBottomNav.find((i) => i.to === '/comfort-targets')?.label).toBe('Targets')
    expect(mobileBottomNav.find((i) => i.to === '/alerts')).toBeUndefined()
  })
})
