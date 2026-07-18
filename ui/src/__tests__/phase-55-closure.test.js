/**
 * Phase 55 WS5 / OC-55 — Guardian ops read tools closure (Vitest bundle guard).
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'
import {
  buildSuppliesHubStarters,
  buildMoneyHubStarters,
  buildZoneGrowStripStarters,
  buildHarvestFlowStarters,
  buildDashboardOpsStarters,
} from '../lib/guardianStarters.js'

const uiSrc = join(process.cwd(), 'src')
const repoDocs = join(process.cwd(), '..', 'docs')

describe('Phase 55 WS5 / OC-55 — Guardian ops closure', () => {
  it('documents architecture, spec, and plan shipped status', () => {
    const arch = readFileSync(join(repoDocs, 'farm-guardian-architecture.md'), 'utf8')
    const spec = readFileSync(join(repoDocs, 'plans/archive/phase_55_guardian_pr_spec.md'), 'utf8')
    const plan = readFileSync(join(repoDocs, 'plans/archive/phase_55_guardian_ops_grow_money.plan.md'), 'utf8')
    expect(arch).toContain('### 7.0s Guardian ops read depth (Phase 55 — shipped)')
    expect(arch).toContain('summarize_cycle_cost')
    expect(arch).toContain('summarize_farm_spending')
    expect(arch).toContain('restock_priority')
    expect(spec).toContain('summarize_active_grows')
    expect(plan).toMatch(/ws5-docs-tests[\s\S]*status: completed/)
    expect(plan).toContain('**Shipped.**')
  })

  it('OC-55 row is closed in operational closure doc', () => {
    const oc = readFileSync(
      join(repoDocs, 'plans/archive/phase_35_37_operational_closure.plan.md'),
      'utf8',
    )
    expect(oc).toContain('## Phase 55 — Guardian ops read depth')
    expect(oc).toMatch(/oc-55-closure[\s\S]*status: completed/)
    expect(oc).toContain('phase-55-closure.test.js')
  })

  it('Go read tools module exists with Phase 55 tool ids', () => {
    const go = readFileSync(join(repoDocs, '../internal/farmguardian/readtools_ops.go'), 'utf8')
    expect(go).toContain('summarize_cycle_cost')
    expect(go).toContain('summarize_farm_spending')
    expect(go).toContain('restock_priority')
    expect(go).toContain('summarize_active_grows')
    expect(existsSync(join(repoDocs, '../internal/farmguardian/readtools_ops_test.go'))).toBe(true)
  })

  it('hub starters reference Phase 55 read tools', () => {
    const supplies = buildSuppliesHubStarters({
      lowStockRows: [{ id: 1, inputName: 'OHN' }],
      lowStockAlerts: [],
      recipes: [],
      zones: [],
    })
    expect(supplies.some((s) => s.id === 'restock-first' && s.message.includes('restock_priority'))).toBe(true)

    const money = buildMoneyHubStarters()
    expect(money.some((s) => s.id === 'tag-receipt-help')).toBe(true)
    expect(money.some((s) => s.message.includes('summarize_farm_spending'))).toBe(true)

    const grow = buildZoneGrowStripStarters({
      zone: { id: 1, name: 'Veg Room' },
      activeCycle: { id: 9, name: 'Run A', current_stage: 'early_veg' },
      farmId: 1,
    })
    expect(grow.some((s) => s.id === 'vpd-on-target')).toBe(true)
    expect(grow[0].contextRef.crop_cycle_id).toBe(9)

    const harvest = buildHarvestFlowStarters({
      zone: { id: 1, name: 'Flower Room' },
      activeCycle: { id: 10, name: 'Harvested run' },
      priorHarvestedCycle: { id: 8, name: 'Last run' },
      surface: 'post_harvest',
    })
    expect(harvest.some((s) => s.id === 'how-did-we-do')).toBe(true)
    expect(harvest.some((s) => s.id === 'cost-per-gram')).toBe(true)

    const dash = buildDashboardOpsStarters({ lowStockCount: 2, lowStockAlerts: [] })
    expect(dash.some((s) => s.id === 'open-supplies')).toBe(true)
  })

  it('PostHarvestScreen wires Guardian starters', () => {
    const vue = readFileSync(join(uiSrc, 'components/PostHarvestScreen.vue'), 'utf8')
    expect(vue).toContain('GuardianStarterChips')
    expect(vue).toContain('buildHarvestFlowStarters')
    expect(vue).toContain("surface: 'post_harvest'")
  })

  it('context_ref carries crop_cycle_id for zone grow starters', () => {
    const go = readFileSync(join(repoDocs, '../internal/farmguardian/context_ref.go'), 'utf8')
    expect(go).toContain('CropCycleID')
    expect(go).toContain('summarize_cycle_cost')
    expect(go).toContain('summarize_farm_spending')
  })
})
