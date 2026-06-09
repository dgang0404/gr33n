/**
 * Phase 56 WS5 / OC-56 — grow schema + harvest analytics closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'
import {
  buildHarvestFlowStarters,
  buildZoneGrowStripStarters,
} from '../lib/guardianStarters.js'
import { buildStartGrowPayload, buildPostHarvestCompareRoute } from '../lib/growHub.js'

const repoDocs = join(process.cwd(), '..', 'docs')

describe('Phase 56 WS5 / OC-56 — grow schema closure', () => {
  it('documents migration, architecture, and plan shipped status', () => {
    const arch = readFileSync(join(repoDocs, 'farm-guardian-architecture.md'), 'utf8')
    const plan = readFileSync(join(repoDocs, 'plans/phase_56_grow_schema_harvest_analytics.plan.md'), 'utf8')
    const tour = readFileSync(join(repoDocs, 'operator-tour.md'), 'utf8')
    expect(existsSync(join(repoDocs, '../db/migrations/20260608_phase56_grow_schema_harvest.sql'))).toBe(true)
    expect(arch).toContain('### 7.0t Grow schema + harvest analytics (Phase 56 — shipped)')
    expect(arch).toContain('plant_id')
    expect(arch).toContain('crop_cycle_stage_events')
    expect(plan).toMatch(/ws5-docs-tests[\s\S]*status: completed/)
    expect(plan).toContain('**Shipped.**')
    expect(tour).toContain('### 6k. Grow schema + harvest analytics (Phase 56 — shipped)')
  })

  it('OC-56 row is closed in operational closure doc', () => {
    const oc = readFileSync(
      join(repoDocs, 'plans/phase_35_37_operational_closure.plan.md'),
      'utf8',
    )
    expect(oc).toContain('## Phase 56 — Grow schema + harvest analytics')
    expect(oc).toMatch(/oc-56-closure[\s\S]*status: completed/)
    expect(oc).toContain('phase-56-closure.test.js')
  })

  it('start grow payload carries plant_id', () => {
    const payload = buildStartGrowPayload({
      zoneId: 2,
      strain: 'OG Kush',
      name: 'Veg run',
      plantId: 7,
    })
    expect(payload.plant_id).toBe(7)
  })

  it('compare route pre-selects current + prior harvested cycles', () => {
    const cycles = [
      { id: 10, zone_id: 1, is_active: false, harvested_at: '2026-01-01' },
      { id: 12, zone_id: 1, is_active: false, harvested_at: '2026-06-01' },
      { id: 15, zone_id: 2, is_active: true },
    ]
    const route = buildPostHarvestCompareRoute(1, cycles, 12, 1)
    expect(route.query.ids).toBe('12,10')
  })

  it('post-harvest Guardian starter includes compare_ids when prior exists', () => {
    const starters = buildHarvestFlowStarters({
      zone: { id: 1, name: 'Flower' },
      activeCycle: { id: 12, name: 'This run' },
      priorHarvestedCycle: { id: 10, name: 'Last run' },
      farmId: 1,
      allCycles: [
        { id: 10, zone_id: 1, is_active: false },
        { id: 12, zone_id: 1, is_active: false },
      ],
      surface: 'post_harvest',
    })
    const compare = starters.find((s) => s.id === 'how-did-we-do')
    expect(compare?.contextRef?.compare_ids).toBe('12,10')
  })

  it('zone grow strip compare starter carries compare_ids', () => {
    const starters = buildZoneGrowStripStarters({
      zone: { id: 1, name: 'Veg' },
      activeCycle: { id: 9, name: 'Run A', current_stage: 'early_veg' },
      farmId: 1,
      priorHarvestedCycle: { id: 8, name: 'Last' },
      allCycles: [
        { id: 8, zone_id: 1, is_active: false },
        { id: 9, zone_id: 1, is_active: true },
      ],
    })
    const compare = starters.find((s) => s.id === 'compare-last-cycle')
    expect(compare?.contextRef?.compare_ids).toBe('9,8')
  })

  it('UI surfaces plant_id, compare ids, and money grow filter', () => {
    const wizard = readFileSync(join(process.cwd(), 'src/components/StartGrowWizard.vue'), 'utf8')
    const plants = readFileSync(join(process.cwd(), 'src/views/Plants.vue'), 'utf8')
    const summary = readFileSync(join(process.cwd(), 'src/views/CropCycleSummary.vue'), 'utf8')
    const money = readFileSync(join(process.cwd(), 'src/views/MoneyHub.vue'), 'utf8')
    expect(wizard).toContain('plantId,')
    expect(plants).toContain('plant-active-cycles')
    expect(summary).toContain('buildPostHarvestCompareRoute')
    expect(summary).toContain('summary-harvest-economics')
    expect(money).toContain('money-grow-filter')
  })
})
