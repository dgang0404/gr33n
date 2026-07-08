/**
 * Phase 146 — Guardian quality loop, judge & ops hardening closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoDocs = join(process.cwd(), '..', 'docs')
const repoRoot = join(process.cwd(), '..')

describe('Phase 146 — quality loop closure', () => {
  it('plan documents optional critique and ops hardening', () => {
    const plan = readFileSync(
      join(repoDocs, 'plans/phase_146_guardian_quality_loop_and_judge.plan.md'),
      'utf8',
    )
    expect(plan).toContain('GUARDIAN_ANSWER_CRITIQUE')
    expect(plan).toContain('guardian-feedback-to-fixture')
  })

  it('synthesis prompt forbids source dumps', () => {
    const syn = readFileSync(join(repoRoot, 'internal/rag/synthesis/synthesis.go'), 'utf8')
    expect(syn).toContain('Do NOT append a "Sources:" list')
    expect(syn).toContain('never invent citation lines')
  })

  it('answer critique module exists', () => {
    const crit = readFileSync(join(repoRoot, 'internal/farmguardian/answer_critique.go'), 'utf8')
    expect(crit).toContain('CritiqueAnswer')
    expect(crit).toContain('GUARDIAN_ANSWER_CRITIQUE')
  })

  it('eval ops use warmup timeout env and Makefile token refresh', () => {
    const env = readFileSync(join(repoRoot, 'internal/farmguardian/eval/env.go'), 'utf8')
    expect(env).toContain('WarmupTimeoutFromEnv')
    expect(env).toContain('ClientTimeoutFromEnv')
    const makefile = readFileSync(join(repoRoot, 'Makefile'), 'utf8')
    expect(makefile).toContain('source-local-env.sh --refresh-eval-token')
  })

  it('ci-guardian-qa documents embed default and optional critique', () => {
    const ci = readFileSync(join(repoDocs, 'ci-guardian-qa.md'), 'utf8')
    expect(ci).toContain('Phase 146')
    expect(ci).toContain('GUARDIAN_ANSWER_CRITIQUE')
  })

  it('runbook documents feedback fixture promotion', () => {
    const runbook = readFileSync(join(repoDocs, 'guardian-feedback-review-runbook.md'), 'utf8')
    expect(runbook).toContain('Promote feedback to regression')
    expect(runbook).toContain('guardian-feedback-to-fixture')
  })
})
