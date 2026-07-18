/**
 * Phase 141 — Guardian feedback review closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoDocs = join(process.cwd(), '..', 'docs')
const repoRoot = join(process.cwd(), '..')

describe('Phase 141 — feedback review closure', () => {
  it('plan is shipped', () => {
    const plan = readFileSync(join(repoDocs, 'plans/archive/phase_141_guardian_feedback_review.plan.md'), 'utf8')
    expect(plan).toContain('**Shipped.**')
  })

  it('runbook and virtual-pi path doc exist', () => {
    const runbook = readFileSync(join(repoDocs, 'guardian-feedback-review-runbook.md'), 'utf8')
    const vpi = readFileSync(join(repoDocs, 'virtual-pi-field-validation-path.md'), 'utf8')
    expect(runbook).toContain('make guardian-qa-smoke')
    expect(vpi).toContain('/virtual-pi')
  })

  it('Settings wires feedback review card', () => {
    const settings = readFileSync(join(process.cwd(), 'src/views/Settings.vue'), 'utf8')
    const card = readFileSync(join(process.cwd(), 'src/components/GuardianSettingsFeedbackReviewCard.vue'), 'utf8')
    expect(settings).toContain('GuardianSettingsFeedbackReviewCard')
    expect(card).toContain('data-test="settings-guardian-feedback"')
    expect(card).toContain('/v1/chat/feedback/export')
  })

  it('QA archive includes feedback_review_prompt', () => {
    const summary = readFileSync(join(repoRoot, 'internal/farmguardian/eval_summary.go'), 'utf8')
    expect(summary).toContain('FeedbackReviewPrompt')
    expect(summary).toContain('feedback_review_prompt')
  })
})
