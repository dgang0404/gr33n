import { describe, it, expect } from 'vitest'
import {
  activeProgramForZone,
  buildFarmFeedingCards,
  detectFeedingAttention,
  filterFeedingCardsByZone,
  countRoomsWithFeedingPlan,
} from '../lib/farmFeedingHub.js'

describe('Phase 47 WS4 — farm feeding hub', () => {
  const zones = [
    { id: 1, name: 'Veg tent', zone_type: 'indoor' },
    { id: 2, name: 'Flower', zone_type: 'indoor' },
  ]

  const programs = [
    {
      id: 10,
      name: 'Veg daily',
      target_zone_id: 1,
      is_active: true,
      schedule_id: 20,
      total_volume_liters: 0.5,
      irrigation_only: false,
      ec_target_id: 5,
      reservoir_id: 1,
    },
    {
      id: 11,
      name: 'Plain water',
      target_zone_id: 2,
      is_active: true,
      total_volume_liters: 1,
      irrigation_only: true,
    },
  ]

  const schedules = [{
    id: 20,
    name: 'Morning veg',
    cron_expression: '0 8 * * *',
    is_active: true,
  }]

  const events = [{
    id: 1,
    zone_id: 1,
    applied_at: '2026-06-04T08:00:00Z',
    volume_applied_liters: 0.5,
    ec_after_mscm: 0.8,
  }]

  const ecTargets = [{ id: 5, ec_min_mscm: 1.1, ec_max_mscm: 1.3 }]
  const reservoirs = [{ id: 1, name: 'Veg tank', status: 'ready' }]

  it('picks active program per zone', () => {
    expect(activeProgramForZone(programs, 1)?.name).toBe('Veg daily')
    expect(activeProgramForZone(programs, 99)).toBeNull()
  })

  it('builds one card per room with plan fields', () => {
    const cards = buildFarmFeedingCards({
      zones,
      programs,
      schedules,
      events,
      ecTargets,
      reservoirs,
    })
    expect(cards).toHaveLength(2)
    expect(cards[0].zone.name).toBe('Veg tent')
    expect(cards[0].plan.hasPlan).toBe(true)
    expect(cards[0].plan.statusLine).toContain('Next feed:')
    expect(cards[1].plan.irrigationOnly).toBe(true)
  })

  it('flags low EC and paused schedules', () => {
    const lowEc = detectFeedingAttention(buildFarmFeedingCards({
      zones: [zones[0]],
      programs,
      schedules,
      events,
      ecTargets,
      reservoirs,
    })[0].plan)
    expect(lowEc?.label).toBe('Last feed below target')

    const paused = detectFeedingAttention(buildFarmFeedingCards({
      zones: [zones[0]],
      programs,
      schedules: [{ ...schedules[0], is_active: false }],
      events,
      ecTargets,
      reservoirs,
    })[0].plan)
    expect(paused?.label).toBe('Feeding paused')
  })

  it('filters cards by zone_id query', () => {
    const cards = buildFarmFeedingCards({ zones, programs, schedules, events, ecTargets, reservoirs })
    expect(filterFeedingCardsByZone(cards, 2)).toHaveLength(1)
    expect(filterFeedingCardsByZone(cards, null)).toHaveLength(2)
  })

  it('counts rooms with an active or any program', () => {
    expect(countRoomsWithFeedingPlan(programs, zones)).toBe(2)
    expect(countRoomsWithFeedingPlan([], zones)).toBe(0)
  })
})
