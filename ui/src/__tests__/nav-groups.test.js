import { describe, it, expect } from 'vitest'
import { buildNavGroups, mobileBottomNav } from '../lib/navGroups.js'

describe('Phase 68 — workspace nav groups', () => {
  const groups = buildNavGroups('/farms/1/crop-cycles/compare')

  it('uses workspace-first grow & operate labels', () => {
    const grow = groups.find((g) => g.label === 'Grow & operate')
    expect(grow.items.some((i) => i.label === 'My zones' && i.to === '/zones')).toBe(true)
    expect(grow.items.some((i) => i.label === 'Feed & water' && i.to === '/feed-water')).toBe(true)
    expect(grow.items.some((i) => i.label === 'Hardware' && i.to === '/hardware')).toBe(true)
    expect(grow.items.some((i) => i.label === 'Money' && i.to === '/money')).toBe(true)
    expect(grow.items.some((i) => i.to === '/feeding')).toBe(false)
  })

  it('keeps power-user routes in Advanced only', () => {
    const advanced = groups.find((g) => g.label === 'Advanced')
    expect(advanced.items.some((i) => i.label === 'Automations')).toBe(true)
    expect(advanced.items.some((i) => i.label === 'Setpoints (raw)')).toBe(true)
    expect(advanced.items.some((i) => i.to === '/fertigation')).toBe(false)
    expect(advanced.items.some((i) => i.to === '/sensors')).toBe(false)
  })

  it('groups Today cockpit items', () => {
    const today = groups.find((g) => g.label === 'Today')
    expect(today.items.some((i) => i.label === 'Today' && i.to === '/')).toBe(true)
    expect(today.items.some((i) => i.to === '/alerts')).toBe(false)
    expect(today.items.some((i) => i.to === '/tasks')).toBe(false)
  })

  it('puts Guardian full page under More', () => {
    const more = groups.find((g) => g.label === 'More')
    expect(more.items.some((i) => i.to === '/chat' && i.label.includes('Guardian'))).toBe(true)
  })

  it('uses farmer labels on mobile bottom nav with feed workspace', () => {
    expect(mobileBottomNav.find((i) => i.to === '/')?.label).toBe('Today')
    expect(mobileBottomNav.find((i) => i.to === '/zones')?.label).toBe('Zones')
    expect(mobileBottomNav.find((i) => i.to === '/feed-water')?.label).toBe('Feed')
  })
})
