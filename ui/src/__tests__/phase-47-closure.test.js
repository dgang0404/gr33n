/**
 * Phase 47 WS7 / OC-47 — documents closure acceptance in Vitest.
 * Individual behaviors are tested in the files listed below; this file guards the bundle.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'
import { buildZoneFeedingPlan } from '../lib/zoneFeedingPlan.js'
import { buildFarmFeedingCards } from '../lib/farmFeedingHub.js'

const uiSrc = join(process.cwd(), 'src')
const repoDocs = join(process.cwd(), '..', 'docs')

describe('Phase 47 WS7 / OC-47 — feeding closure', () => {
  it('documents operator-tour §7b and architecture §7.0m as shipped', () => {
    const tour = readFileSync(join(repoDocs, 'operator-tour.md'), 'utf8')
    const arch = readFileSync(join(repoDocs, 'farm-guardian-architecture.md'), 'utf8')
    expect(tour).toContain('## 7b. Feeding & water for this zone (Phase 47)')
    expect(tour).toContain('**Shipped.**')
    expect(arch).toContain('### 7.0m Feeding & water plain language (Phase 47)')
    expect(arch).not.toContain('7.0m Feeding & water plain language (Phase 47 — planned)')
  })

  it('feeding plan card acceptance — plan fields for grow story', () => {
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
      schedules: [{ id: 20, name: 'Daily', cron_expression: '0 8 * * *', is_active: true }],
      events: [],
      ecTargets: [{ id: 5, ec_min_mscm: 1.1, ec_max_mscm: 1.3 }],
      reservoirs: [{ id: 1, name: 'Flower tank', status: 'ready' }],
      waterStatus: { queue_depth: 0, mix_required: true },
    })
    expect(plan.hasPlan).toBe(true)
    expect(plan.statusLine).toContain('Next feed:')
    expect(plan.irrigationOnly).toBe(false)
    expect(plan.mixRequired).toBe(true)
  })

  it('irrigation-only hides mix requirement on the plan model', () => {
    const plan = buildZoneFeedingPlan({
      zoneId: 3,
      activeProgram: { id: 11, name: 'Plain water', irrigation_only: true, total_volume_liters: 1 },
      schedules: [],
      events: [],
      ecTargets: [],
      reservoirs: [],
      waterStatus: { mix_required: false },
    })
    expect(plan.irrigationOnly).toBe(true)
    expect(plan.mixRequired).toBe(false)
  })

  it('farm feeding hub builds one card per zone', () => {
    const cards = buildFarmFeedingCards({
      zones: [{ id: 1, name: 'Veg' }, { id: 2, name: 'Flower' }],
      programs: [{ id: 10, target_zone_id: 1, is_active: true, name: 'Veg daily' }],
      schedules: [],
      events: [],
      ecTargets: [],
      reservoirs: [],
    })
    expect(cards).toHaveLength(2)
  })

  it('closure Vitest files exist', () => {
    for (const f of [
      '__tests__/zone-feeding-water.test.js',
      '__tests__/zone-feeding-plan.test.js',
      '__tests__/farm-feeding-hub.test.js',
      '__tests__/farmer-vocabulary-grow-path.test.js',
      '__tests__/guardian-context-prompts.test.js',
    ]) {
      expect(existsSync(join(uiSrc, f))).toBe(true)
    }
  })
})
