import { describe, it, expect } from 'vitest'
import {
  buildZoneGuardianContextRef,
  buildZoneGuardianPrompt,
} from '../lib/guardianContextPrompts.js'
import { buildZoneStarters, buildFeedingHubStarters } from '../lib/guardianStarters.js'

describe('Phase 40 WS7b — contextual zone Guardian prompts', () => {
  const zone = { id: 3, name: 'Flower Room' }

  it('prefers alert explanation over generic status', () => {
    const msg = buildZoneGuardianPrompt({
      zone,
      unreadAlerts: [{ id: 99, subject_rendered: 'Humidity high — Flower Room' }],
    })
    expect(msg).toContain('alert #99')
    expect(msg).not.toMatch(/current status/i)
  })

  it('asks about queue when commands are pending', () => {
    const msg = buildZoneGuardianPrompt({
      zone,
      unreadAlerts: [],
      queueDepth: 3,
    })
    expect(msg).toContain('queued')
    expect(msg).toContain('Flower Room')
  })

  it('includes zone tab on contextRef when not overview', () => {
    const ref = buildZoneGuardianContextRef({
      zone,
      activeTab: 'water',
      unreadAlerts: [{ id: 1 }],
    })
    expect(ref).toEqual({
      type: 'zone',
      id: 3,
      name: 'Flower Room',
      tab: 'water',
      alert_id: 1,
    })
  })

  it('builds starter chips from snapshot (max 5)', () => {
    const starters = buildZoneStarters('zone_overview', {
      zone,
      unreadAlerts: [{ id: 1, subject_rendered: 'Humidity high' }],
      nextSchedule: { schedule: { name: 'Water Early Flower Daily' } },
      queueDepth: 1,
      missingComfortTargets: 2,
    })
    expect(starters.length).toBeGreaterThan(0)
    expect(starters.length).toBeLessThanOrEqual(5)
    expect(starters[0].message).toContain('alert #1')
  })

  it('builds Phase 47 water tab feeding starters', () => {
    const starters = buildZoneStarters('zone_water', {
      zone,
      activeProgramName: 'Flower FFJ',
    })
    expect(starters.map((s) => s.id)).toEqual(['next-feed', 'run-feed-safe', 'water-only'])
    expect(starters[0].message).toContain('next feed')
    expect(starters[0].contextRef.tab).toBe('water')
  })

  it('builds feeding hub starters for a focused room', () => {
    const starters = buildFeedingHubStarters({
      zones: [{ id: 3, name: 'Flower Room' }],
      zoneContextId: 3,
    })
    expect(starters[0].id).toBe('next-feed')
    expect(starters[0].message).toContain('Flower Room')
  })
})
