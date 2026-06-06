import { describe, it, expect } from 'vitest'
import {
  formatGrowthStage,
  reservoirFillPct,
  reservoirStatusLabel,
  buildProgramAdminCard,
  buildReservoirAdminCard,
  buildEcTargetAdminCard,
  filterReservoirsForZone,
  filterEcTargetsForZone,
} from '../lib/feedingAdminHub.js'

describe('Phase 43 WS3 — feeding admin hub helpers', () => {
  const zones = [{ id: 1, name: 'Flower Room' }, { id: 2, name: 'Veg' }]
  const schedules = [{ id: 10, name: 'Morning feed', cron_expression: '0 6 * * *', is_active: true }]

  it('formats growth stage labels', () => {
    expect(formatGrowthStage('early_veg')).toBe('early veg')
  })

  it('computes reservoir fill percent and status', () => {
    expect(reservoirFillPct({ capacity_liters: 100, current_volume_liters: 25 })).toBe(25)
    expect(reservoirStatusLabel({ status: 'ready' }).label).toBe('Ready')
    expect(reservoirStatusLabel({ status: 'needs_top_up' }).label).toBe('Needs top-up')
  })

  it('builds program cards with room name and water-only badge', () => {
    const card = buildProgramAdminCard({
      id: 5,
      name: 'Bloom feed',
      target_zone_id: 1,
      schedule_id: 10,
      is_active: true,
      irrigation_only: true,
      total_volume_liters: 12,
    }, zones, schedules)
    expect(card.zoneName).toBe('Flower Room')
    expect(card.irrigationOnly).toBe(true)
    expect(card.isActive).toBe(true)
    expect(card.nextRunLabel).toBeTruthy()
  })

  it('builds reservoir cards with volume bar data', () => {
    const card = buildReservoirAdminCard({
      id: 3,
      name: 'Tank A',
      zone_id: 1,
      capacity_liters: 200,
      current_volume_liters: 40,
      status: 'needs_top_up',
    }, zones)
    expect(card.fillPct).toBe(20)
    expect(card.statusLabel).toBe('Needs top-up')
    expect(card.zoneName).toBe('Flower Room')
  })

  it('builds EC target cards with stage and range', () => {
    const card = buildEcTargetAdminCard({
      id: 7,
      growth_stage: 'mid_flower',
      zone_id: null,
      ec_min_mscm: 1.8,
      ec_max_mscm: 2.2,
      ph_min: 5.8,
      ph_max: 6.2,
    }, zones)
    expect(card.stageLabel).toBe('mid flower')
    expect(card.zoneName).toBe('All zones')
    expect(card.ecRange).toBe('1.8–2.2 mS/cm')
    expect(card.phRange).toContain('pH')
  })

  it('filters reservoirs and EC targets by zone', () => {
    const reservoirs = [
      { id: 1, zone_id: 1 },
      { id: 2, zone_id: 2 },
      { id: 3, zone_id: null },
    ]
    expect(filterReservoirsForZone(reservoirs, 1).map((r) => r.id)).toEqual([1, 3])

    const targets = [
      { id: 1, zone_id: 2 },
      { id: 2, zone_id: null },
    ]
    expect(filterEcTargetsForZone(targets, 1).map((t) => t.id)).toEqual([2])
  })
})
