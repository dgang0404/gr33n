/**
 * Phase 147 — Smoke run #5 closure & eval isolation.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoDocs = join(process.cwd(), '..', 'docs')
const repoRoot = join(process.cwd(), '..')

describe('Phase 147 — smoke run #5 closure', () => {
  it('plan documents prompt isolation and run #5', () => {
    const plan = readFileSync(
      join(repoDocs, 'plans/archive/phase_147_guardian_smoke_run5_closure.plan.md'),
      'utf8',
    )
    expect(plan).toContain('guardian-qa-smoke-ec-ph')
    expect(plan).toContain('GUARDIAN_EVAL_PROMPT_IDS')
    expect(plan).toContain('run #5')
  })

  it('eval filter supports single smoke prompt', () => {
    const filter = readFileSync(join(repoRoot, 'internal/farmguardian/eval/filter.go'), 'utf8')
    expect(filter).toContain('FilterFixturesByIDs')
    const main = readFileSync(join(repoRoot, 'cmd/guardian-eval/main.go'), 'utf8')
    expect(main).toContain('prompt-ids')
    expect(main).toContain('GUARDIAN_EVAL_PROMPT_IDS')
  })

  it('Makefile has smoke-ec-ph target with token refresh', () => {
    const makefile = readFileSync(join(repoRoot, 'Makefile'), 'utf8')
    expect(makefile).toContain('guardian-qa-smoke-ec-ph')
    expect(makefile).toContain('-prompt-ids smoke-ec-ph')
    expect(makefile).toContain('source-local-env.sh --refresh-eval-token')
  })

  it('laptop tune recommends eval client timeout', () => {
    const tune = readFileSync(join(repoRoot, 'scripts/tune-guardian-laptop.sh'), 'utf8')
    expect(tune).toContain('GUARDIAN_EVAL_TIMEOUT_SECONDS')
  })

  it('Settings QA card shows critique column when present', () => {
    const qa = readFileSync(join(process.cwd(), 'src/components/GuardianSettingsQARunCard.vue'), 'utf8')
    expect(qa).toContain('showCritiqueCol')
    expect(qa).toContain('critique_pass')
  })

  it('smoke report documents Phase 147 run #5', () => {
    const report = readFileSync(join(repoDocs, 'guardian-qa-smoke-report-20260707.md'), 'utf8')
    expect(report).toContain('run #5')
    expect(report).toContain('guardian-qa-smoke-ec-ph')
  })

  it('architecture documents eval isolation §8.11', () => {
    const arch = readFileSync(join(repoDocs, 'farm-guardian-architecture.md'), 'utf8')
    expect(arch).toContain('### 8.11 Smoke run #5 & eval isolation (Phase 147)')
  })
})
