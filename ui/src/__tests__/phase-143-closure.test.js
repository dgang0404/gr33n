/**
 * Phase 143 — Guardian answer quality closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoDocs = join(process.cwd(), '..', 'docs')
const repoRoot = join(process.cwd(), '..')

describe('Phase 143 — answer quality closure', () => {
  it('plan is shipped', () => {
    const plan = readFileSync(join(repoDocs, 'plans/archive/phase_143_guardian_answer_quality.plan.md'), 'utf8')
    expect(plan).toContain('**Shipped.**')
  })

  it('smoke report links Phase 143 and runbook checklist', () => {
    const report = readFileSync(join(repoDocs, 'guardian-qa-smoke-report-20260707.md'), 'utf8')
    const runbook = readFileSync(join(repoDocs, 'guardian-feedback-review-runbook.md'), 'utf8')
    expect(report).toContain('Phase 143')
    expect(report).toContain('run #3')
    expect(runbook).toContain('## Smoke quality checklist (Phase 143)')
  })

  it('architecture documents answer hygiene finalize path', () => {
    const arch = readFileSync(join(repoDocs, 'farm-guardian-architecture.md'), 'utf8')
    expect(arch).toContain('### 8.8 Answer hygiene (Phase 143)')
    expect(arch).toContain('TrimInstructionLeak')
    expect(arch).toContain('phase_143_guardian_answer_quality.plan.md')
  })

  it('Go ships leak guard, citation sanitize, and smoke quality heuristics', () => {
    const leak = readFileSync(join(repoRoot, 'internal/farmguardian/answer_leak.go'), 'utf8')
    const cite = readFileSync(join(repoRoot, 'internal/farmguardian/answer_citation.go'), 'utf8')
    const score = readFileSync(join(repoRoot, 'internal/farmguardian/eval/score.go'), 'utf8')
    expect(leak).toContain('TrimInstructionLeak')
    expect(cite).toContain('SanitizeCitationURLs')
    expect(score).toContain('SmokeTopicDriftNote')
  })
})
