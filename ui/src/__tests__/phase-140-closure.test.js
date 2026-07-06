/**
 * Phase 140 — Guardian QA Settings summary closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoDocs = join(process.cwd(), '..', 'docs')
const repoRoot = join(process.cwd(), '..')

describe('Phase 140 — QA Settings summary closure', () => {
  it('plan is shipped', () => {
    const plan = readFileSync(join(repoDocs, 'plans/phase_140_guardian_qa_settings.plan.md'), 'utf8')
    expect(plan).toContain('**Shipped.**')
  })

  it('131 WS7 marked completed', () => {
    const plan = readFileSync(join(repoDocs, 'plans/phase_131_guardian_qa_harness.plan.md'), 'utf8')
    expect(plan).toContain('ws7-ui-optional')
    expect(plan).toMatch(/ws7-ui-optional[\s\S]*status: completed/)
  })

  it('Settings wires QA run card', () => {
    const settings = readFileSync(join(process.cwd(), 'src/views/Settings.vue'), 'utf8')
    const card = readFileSync(join(process.cwd(), 'src/components/GuardianSettingsQARunCard.vue'), 'utf8')
    expect(settings).toContain('GuardianSettingsQARunCard')
    expect(card).toContain('data-test="settings-guardian-qa"')
    expect(card).toContain('/v1/guardian/qa/latest')
  })

  it('Go exposes latest QA loader and route', () => {
    const routes = readFileSync(join(repoRoot, 'cmd/api/routes.go'), 'utf8')
    const summary = readFileSync(join(repoRoot, 'internal/farmguardian/eval_summary.go'), 'utf8')
    const handler = readFileSync(join(repoRoot, 'internal/handler/chat/qa_latest.go'), 'utf8')
    expect(routes).toContain('/v1/guardian/qa/latest')
    expect(summary).toContain('LoadLatestQARun')
    expect(handler).toContain('GetLatestQARun')
  })
})
