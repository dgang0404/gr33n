/**
 * Phase 206 — docs/plans archive migration closure.
 */
import { describe, it, expect } from 'vitest'
import { execSync } from 'node:child_process'
import { existsSync, readFileSync, readdirSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')
const plansDir = join(repoRoot, 'docs/plans')
const archiveDir = join(plansDir, 'archive')

const HUBS_AT_TOP = [
  'product_backlog_operator_runtime.plan.md',
  'pre_development_gaps_index.plan.md',
  'phase_84_100_master_roadmap.plan.md',
  'phase_68_73_spa_workspace_roadmap.plan.md',
  'farmer_ux_roadmap_40_plus.plan.md',
  'phase_53_59_roadmap.plan.md',
  'phase_173_177_today_excellence_roadmap.plan.md',
  'phase_205_pre_existing_test_debt.plan.md',
  'phase_206_docs_plans_archive_migration.plan.md',
]

describe('Phase 206 — docs/plans archive migration', () => {
  it('ships the inventory and migration scripts', () => {
    expect(existsSync(join(repoRoot, 'scripts/docs-plans-archive-inventory.mjs'))).toBe(true)
    expect(existsSync(join(repoRoot, 'scripts/migrate-plans-to-archive.mjs'))).toBe(true)
  })

  it('keeps only hub + meta plans at docs/plans/ root', () => {
    const rootPlans = readdirSync(plansDir).filter((n) => n.endsWith('.plan.md')).sort()
    expect(rootPlans).toEqual([...HUBS_AT_TOP].sort())
  })

  it('moved shipped plans into docs/plans/archive/', () => {
    const archiveCount = readdirSync(archiveDir).filter((n) => n.endsWith('.plan.md')).length
    expect(archiveCount).toBeGreaterThanOrEqual(190)
  })

  it('inventory batch A is empty after migration', () => {
    const out = execSync('node scripts/docs-plans-archive-inventory.mjs --batch=A', {
      cwd: repoRoot,
      encoding: 'utf8',
    })
    expect(out.trim()).toBe('')
  }, 30000)

  it('archive README documents Phase 206 and hub exclusions', () => {
    const readme = readFileSync(join(archiveDir, 'README.md'), 'utf8')
    expect(readme).toContain('Phase 206')
    expect(readme).toContain('phase_68_73_spa_workspace_roadmap')
  })

  it('plan is marked shipped with acceptance criteria met', () => {
    const plan = readFileSync(join(plansDir, 'phase_206_docs_plans_archive_migration.plan.md'), 'utf8')
    expect(plan).toContain('**Status:** shipped')
    expect(plan).toContain('status: completed')
  })
})
