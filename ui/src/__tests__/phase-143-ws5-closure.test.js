/**
 * Phase 143 — WS5 smoke quality checklist closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoDocs = join(process.cwd(), '..', 'docs')
const repoRoot = join(process.cwd(), '..')

describe('Phase 143 — WS5 feedback quality checklist', () => {
  it('runbook documents smoke quality checklist', () => {
    const runbook = readFileSync(join(repoDocs, 'guardian-feedback-review-runbook.md'), 'utf8')
    expect(runbook).toContain('## Smoke quality checklist (Phase 143)')
    expect(runbook).toContain('smoke-morning-walk')
    expect(runbook).toContain('smoke-ec-ph')
    expect(runbook).toContain('gr33n.com')
  })

  it('QA archive prompt references smoke quality checklist', () => {
    const summary = readFileSync(join(repoRoot, 'internal/farmguardian/eval_summary.go'), 'utf8')
    expect(summary).toContain('QAFeedbackReviewPrompt')
    expect(summary).toContain('Smoke quality checklist')
    expect(summary).toContain('guardian qa: archive saved')
  })

  it('Settings QA card nudges operators to quality checklist', () => {
    const card = readFileSync(join(process.cwd(), 'src/components/GuardianSettingsQARunCard.vue'), 'utf8')
    expect(card).toContain('settings-guardian-qa-quality-nudge')
    expect(card).toContain('Guardian feedback')
  })
})
