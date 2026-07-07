/**
 * Phase 144 — Guardian answer quality residuals closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoDocs = join(process.cwd(), '..', 'docs')
const repoRoot = join(process.cwd(), '..')

describe('Phase 144 — answer quality residuals closure', () => {
  it('plan is shipped', () => {
    const plan = readFileSync(join(repoDocs, 'plans/phase_144_guardian_answer_quality_residuals.plan.md'), 'utf8')
    expect(plan).toContain('**Shipped.**')
  })

  it('citation sanitize covers gr33n-docs', () => {
    const cite = readFileSync(join(repoRoot, 'internal/farmguardian/answer_citation.go'), 'utf8')
    expect(cite).toContain('gr33n-docs')
  })

  it('meta correction trim ships in finalize chain', () => {
    const leak = readFileSync(join(repoRoot, 'internal/farmguardian/answer_leak.go'), 'utf8')
    const finalize = readFileSync(join(repoRoot, 'internal/handler/chat/answer_finalize.go'), 'utf8')
    const debug = readFileSync(join(process.cwd(), 'src/components/GuardianTurnDebug.vue'), 'utf8')
    expect(leak).toContain('TrimMetaCorrection')
    expect(finalize).toContain('TrimMetaCorrection')
    expect(debug).toContain('meta_correction_trimmed')
  })

  it('smoke heuristics cover run #3 residuals', () => {
    const score = readFileSync(join(repoRoot, 'internal/farmguardian/eval/score.go'), 'utf8')
    const tests = readFileSync(join(repoRoot, 'internal/farmguardian/eval/score_smoke_quality_test.go'), 'utf8')
    expect(score).toContain('smokeECPHQualityNote')
    expect(score).toContain('AnswerContainsMetaCorrection')
    expect(tests).toContain('archivedRun3MorningWalk')
    expect(tests).toContain('archivedRun3ECPHDrift')
  })
})
