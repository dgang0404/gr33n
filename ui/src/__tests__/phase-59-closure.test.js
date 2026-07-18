/**
 * Phase 59 WS4 / OC-59 — enterprise tier boundary closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'
import { findBannedCopyInFarmerVue } from '../lib/enterpriseTierCopyAudit.js'

const repoDocs = join(process.cwd(), '..', 'docs')
const repoRoot = join(process.cwd(), '..')

describe('Phase 59 WS4 / OC-59 — enterprise tier boundary', () => {
  it('publishes enterprise-tier-boundary.md and marks plan shipped', () => {
    const boundary = readFileSync(join(repoDocs, 'enterprise-tier-boundary.md'), 'utf8')
    const plan = readFileSync(join(repoDocs, 'plans/archive/phase_59_enterprise_tier_boundary.plan.md'), 'utf8')
    expect(boundary).toContain('Out of scope (enterprise tier')
    expect(boundary).toContain('METRC')
    expect(boundary).toContain('Purchase orders')
    expect(plan).toContain('**Shipped.**')
  })

  it('README and gaps index link to boundary doc', () => {
    const readme = readFileSync(join(repoRoot, 'README.md'), 'utf8')
    const gaps = readFileSync(join(repoDocs, 'plans/pre_development_gaps_index.plan.md'), 'utf8')
    const roadmap = readFileSync(join(repoDocs, 'plans/phase_53_59_roadmap.plan.md'), 'utf8')
    expect(readme).toContain('enterprise-tier-boundary.md')
    expect(readme).toContain('Product tiers')
    expect(gaps).toContain('enterprise-tier-boundary.md')
    expect(roadmap).toContain('enterprise-tier-boundary.md')
  })

  it('farmer Vue surfaces have no banned ERP jargon', () => {
    const hits = findBannedCopyInFarmerVue(join(process.cwd(), 'src'))
    expect(hits, hits.map((h) => `${h.file}: ${h.term}`).join('\n')).toEqual([])
  })

  it('phase 53–58 plans defer enterprise capabilities explicitly', () => {
    const plans = [
      'phase_53_grow_stock_money_closure.plan.md',
      'phase_54_zone_connection_nav.plan.md',
      'phase_55_guardian_ops_grow_money.plan.md',
      'phase_56_grow_schema_harvest_analytics.plan.md',
      'phase_57_pi_device_api_keys.plan.md',
      'phase_58_task_consumptions_runtime.plan.md',
    ]
    for (const name of plans) {
      const text = readFileSync(join(repoDocs, 'plans/archive', name), 'utf8')
      const defers =
        /phase_59|enterprise.tier|enterprise tier|METRC|purchase order/i.test(text) ||
        name.includes('phase_58')
      expect(defers, `${name} should reference enterprise boundary or be runtime-only`).toBe(true)
    }
  })

  it('boundary doc exists on main path', () => {
    expect(existsSync(join(repoDocs, 'enterprise-tier-boundary.md'))).toBe(true)
  })
})
