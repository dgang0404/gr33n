import { describe, it, expect } from 'vitest'
import { buildNavGroups, mobileBottomNav } from '../lib/navGroups.js'

describe('Phase 40 WS7 — farmer nav groups', () => {
  const groups = buildNavGroups('/farms/1/crop-cycles/compare')

  it('uses grow-first labels', () => {
    const grow = groups.find((g) => g.label === 'Grow')
    expect(grow.items.some((i) => i.label === 'My rooms' && i.to === '/zones')).toBe(true)
    expect(grow.items.some((i) => i.label === 'Feed & water' && i.to === '/feeding')).toBe(true)
    expect(grow.items.some((i) => i.label === 'Supplies')).toBe(true)
  })

  it('groups Today cockpit items and moves alerts out of Monitor-only', () => {
    const today = groups.find((g) => g.label === 'Today')
    expect(today.items.some((i) => i.label === 'Today' && i.to === '/')).toBe(true)
    expect(today.items.some((i) => i.to === '/alerts')).toBe(true)

    const monitor = groups.find((g) => g.label === 'Monitor')
    expect(monitor.items.some((i) => i.to === '/alerts')).toBe(false)
  })

  it('puts Guardian full page under System', () => {
    const system = groups.find((g) => g.label === 'System')
    expect(system.items.some((i) => i.to === '/chat' && i.label.includes('Guardian'))).toBe(true)
  })

  it('renames Advanced power-user entries', () => {
    const advanced = groups.find((g) => g.label === 'Advanced')
    expect(advanced.items.some((i) => i.label === 'Automations')).toBe(true)
    expect(advanced.items.some((i) => i.label === 'Comfort bands')).toBe(true)
    expect(advanced.items.some((i) => i.label === 'Feeding (technical)' && i.to === '/fertigation')).toBe(true)
  })

  it('uses farmer labels on mobile bottom nav', () => {
    expect(mobileBottomNav.find((i) => i.to === '/')?.label).toBe('Today')
    expect(mobileBottomNav.find((i) => i.to === '/zones')?.label).toBe('Rooms')
  })
})
