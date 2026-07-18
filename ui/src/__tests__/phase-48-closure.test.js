/**
 * Phase 48 WS7 / OC-48 — dev seed hygiene closure (Vitest bundle guard).
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')
const repoDocs = join(repoRoot, 'docs')
const repoScripts = join(repoRoot, 'scripts')

describe('Phase 48 WS7 / OC-48 — dev seed closure', () => {
  it('dev-farm-profiles.md documents small_indoor and demo_showcase', () => {
    const doc = readFileSync(join(repoDocs, 'dev-farm-profiles.md'), 'utf8')
    expect(doc).toContain('small_indoor')
    expect(doc).toContain('demo_showcase')
    expect(doc).toContain('dev-reset-farm.sh')
  })

  it('migration adds farms.meta_data and sensor unique index', () => {
    const mig = readFileSync(
      join(repoRoot, 'db/migrations/20260606_phase48_dev_seed_profiles.sql'),
      'utf8',
    )
    expect(mig).toContain('meta_data')
    expect(mig).toContain('uq_sensors_farm_name_active')
    expect(mig).toContain('set_dev_seed_profile')
  })

  it('master_seed.sql uses idempotent sensor insert', () => {
    const seed = readFileSync(join(repoRoot, 'db/seeds/master_seed.sql'), 'utf8')
    expect(seed).toContain('dev_seed_profile')
    expect(seed).toContain('demo_showcase')
    expect(seed).toMatch(/INSERT INTO gr33ncore\.sensors[\s\S]*WHERE NOT EXISTS/)
  })

  it('dev-reset-farm.sh and sanity report extensions exist', () => {
    expect(existsSync(join(repoScripts, 'dev-reset-farm.sh'))).toBe(true)
    expect(existsSync(join(repoScripts, 'sql/db_sanity_report.sql'))).toBe(true)
    const sanity = readFileSync(join(repoScripts, 'sql/db_sanity_report.sql'), 'utf8')
    expect(sanity).toContain('dev_seed_profile')
    expect(sanity).toContain('Duplicate active sensor names')
  })

  it('local-operator-bootstrap documents Phase 48 reset path', () => {
    const boot = readFileSync(join(repoDocs, 'local-operator-bootstrap.md'), 'utf8')
    expect(boot).toContain('dev-reset-farm.sh')
    expect(boot).toContain('phase_48_dev_seed_and_small_farm_profiles.plan.md')
  })

  it('architecture §7.0n documents dev seed profiles shipped', () => {
    const arch = readFileSync(join(repoDocs, 'farm-guardian-architecture.md'), 'utf8')
    expect(arch).toContain('### 7.0n Dev seed profiles (Phase 48')
    expect(arch).toContain('phase-48-closure.test.js')
  })

  it('OC-48 marked completed in operational closure plan', () => {
    const closure = readFileSync(
      join(repoDocs, 'plans/archive/phase_35_37_operational_closure.plan.md'),
      'utf8',
    )
    expect(closure).toContain('oc-48-closure')
    expect(closure).toMatch(/oc-48-closure[\s\S]*status: completed/)
    expect(closure).toContain('## Phase 48 — Dev seed hygiene')
  })

  it('phase 48 plan marks all workstreams completed', () => {
    const plan = readFileSync(
      join(repoDocs, 'plans/archive/phase_48_dev_seed_and_small_farm_profiles.plan.md'),
      'utf8',
    )
    for (const id of [
      'ws1-profiles-spec',
      'ws2-seed-idempotency',
      'ws3-dev-reset-script',
      'ws4-bootstrap-alignment',
      'ws5-timescale-retention',
      'ws6-sanity-report',
      'ws7-docs-tests',
    ]) {
      expect(plan).toContain(`id: ${id}`)
      expect(plan).toMatch(new RegExp(`${id}[\\s\\S]*status: completed`))
    }
    expect(plan).toContain('**Shipped.**')
  })
})
