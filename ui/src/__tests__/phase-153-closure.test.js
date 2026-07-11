/**
 * Phase 153 — Guardian PR smoke gate (-fail-on-regression + opt-in CI job).
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoDocs = join(process.cwd(), '..', 'docs')
const repoRoot = join(process.cwd(), '..')

describe('Phase 153 — Guardian PR smoke gate', () => {
  it('plan documents the exit-code fix and the opt-in CI trigger', () => {
    const plan = readFileSync(
      join(repoDocs, 'plans/phase_153_guardian_pr_smoke_gate.plan.md'),
      'utf8',
    )
    expect(plan).toContain('fail-on-regression')
    expect(plan).toContain('regressionFailures')
    expect(plan).toContain('guardian-smoke')
    expect(plan).toContain('Not a required check')
  })

  it('cmd/guardian-eval exits non-zero on a fixture regression', () => {
    const main = readFileSync(join(repoRoot, 'cmd/guardian-eval/main.go'), 'utf8')
    expect(main).toContain('fail-on-regression')
    expect(main).toContain('func regressionFailures')
    expect(main).toContain('os.Exit(1)')

    const makefile = readFileSync(join(repoRoot, 'Makefile'), 'utf8')
    expect(makefile).toContain('guardian-qa-pr-check')
    expect(makefile).toContain('-fail-on-regression')
  })

  it('CI job is opt-in — label or workflow_dispatch, self-hosted+ollama runner, never a required default gate', () => {
    const ci = readFileSync(join(repoRoot, '.github/workflows/ci.yml'), 'utf8')
    expect(ci).toContain('guardian-qa-pr:')
    expect(ci).toContain('guardian-smoke')
    expect(ci).toContain('[self-hosted, ollama]')
    expect(ci).toContain('guardian-qa-pr-check')
  })
})
