/**
 * Phase 205 WS1 — regression safety net closure.
 * WS2-6 (fixing the 24 pre-existing failures) land incrementally; this
 * guards the baseline-diff mechanism itself, not the debt it tracks.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')

describe('Phase 205 WS1 — UI test baseline safety net', () => {
  it('ships the baseline file and checker script', () => {
    expect(existsSync(join(process.cwd(), 'test-baseline-known-failures.json'))).toBe(true)
    expect(existsSync(join(repoRoot, 'scripts/check-ui-test-baseline.mjs'))).toBe(true)
  })

  it('baseline file is well-formed and only tracks failures, not passes', () => {
    const baseline = JSON.parse(readFileSync(join(process.cwd(), 'test-baseline-known-failures.json'), 'utf8'))
    expect(baseline.failures).toBeTypeOf('object')
    const total = Object.values(baseline.failures).flat().length
    expect(total).toBeGreaterThan(0)
    for (const [file, tests] of Object.entries(baseline.failures)) {
      expect(existsSync(join(process.cwd(), file))).toBe(true)
      expect(Array.isArray(tests)).toBe(true)
      expect(tests.length).toBeGreaterThan(0)
    }
  })

  it('make target and CONTRIBUTING.md wire the check in', () => {
    const makefile = readFileSync(join(repoRoot, 'Makefile'), 'utf8')
    expect(makefile).toContain('check-ui-test-baseline:')
    const contributing = readFileSync(join(repoRoot, 'CONTRIBUTING.md'), 'utf8')
    expect(contributing).toContain('check-ui-test-baseline')
    expect(contributing).toContain('phase_205_pre_existing_test_debt.plan.md')
  })

  it('plan documents the root-cause triage', () => {
    const plan = readFileSync(join(repoRoot, 'docs/plans/phase_205_pre_existing_test_debt.plan.md'), 'utf8')
    expect(plan).toContain('attachProposals')
    expect(plan).toContain('resetUnauthorizedGate')
  })
})
