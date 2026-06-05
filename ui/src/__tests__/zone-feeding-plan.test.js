import { describe, it, expect } from 'vitest'
import {
  buildZoneFeedingPlan,
  buildFeedingPlanStatusLine,
  mapReservoirStatus,
  resolveEcRange,
  formatLastEventSummary,
} from '../lib/zoneFeedingPlan.js'

describe('Phase 47 WS1 — zone feeding plan', () => {
  it('resolves EC range from linked target', () => {
    const range = resolveEcRange(
      { ec_target_id: 5 },
      [{ id: 5, ec_min_mscm: 1.1, ec_max_mscm: 1.3 }],
    )
    expect(range.label).toBe('1.1–1.3 mS/cm')
  })

  it('maps reservoir status for operators', () => {
    expect(mapReservoirStatus({ name: 'Veg tank', status: 'ready' }).label).toContain('ready')
    expect(mapReservoirStatus({ name: 'Main', status: 'needs_top_up' }).tone).toBe('warn')
  })

  it('formats last feed summary in plain language', () => {
    expect(formatLastEventSummary({
      volume_applied_liters: 0.3,
      ec_after_mscm: 1.2,
    })).toBe('Fed 0.3L · EC 1.2 · OK')
  })

  it('builds status line with next run, volume, and EC', () => {
    const plan = buildZoneFeedingPlan({
      zoneId: 3,
      activeProgram: {
        id: 10,
        name: 'Flower FFJ',
        schedule_id: 20,
        total_volume_liters: 0.3,
        irrigation_only: false,
        ec_target_id: 5,
        reservoir_id: 1,
      },
      schedules: [{
        id: 20,
        name: 'Water Early Flower Daily',
        cron_expression: '0 8 * * *',
        is_active: true,
      }],
      events: [{
        id: 2,
        zone_id: 3,
        applied_at: '2026-06-04T08:00:00Z',
        volume_applied_liters: 0.9,
        ec_after_mscm: 2.1,
      }],
      ecTargets: [{ id: 5, ec_min_mscm: 1.1, ec_max_mscm: 1.3 }],
      reservoirs: [{ id: 1, name: 'Flower tank', status: 'ready', delivery_actuator_id: 99 }],
      waterStatus: { queue_depth: 0, mix_required: true },
    })

    expect(plan.hasPlan).toBe(true)
    expect(plan.statusLine).toContain('Next feed:')
    expect(plan.statusLine).toContain('0.3L')
    expect(plan.statusLine).toContain('EC 1.1–1.3')
    expect(plan.reservoirLabel).toContain('Flower tank')
    expect(plan.irrigationOnly).toBe(false)
    expect(buildFeedingPlanStatusLine(plan)).toBe(plan.statusLine)
  })

  it('hides mix requirement for irrigation-only programs', () => {
    const plan = buildZoneFeedingPlan({
      zoneId: 3,
      activeProgram: {
        id: 11,
        name: 'Outdoor plain water',
        irrigation_only: true,
        total_volume_liters: 1,
      },
      schedules: [],
      events: [],
      waterStatus: { mix_required: false },
    })
    expect(plan.irrigationOnly).toBe(true)
    expect(plan.mixRequired).toBe(false)
  })
})
