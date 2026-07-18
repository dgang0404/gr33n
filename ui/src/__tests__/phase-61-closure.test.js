/**
 * Phase 61 WS5 / OC-61 — proactive Guardian nudges closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'
import { buildNudgeReviewPayload } from '../lib/guardianNudge.js'

const repoDocs = join(process.cwd(), '..', 'docs')
const repoRoot = join(process.cwd(), '..')

describe('Phase 61 WS5 / OC-61 — proactive nudges closure', () => {
  it('documents guardian-nudge API and architecture section', () => {
    const arch = readFileSync(join(repoDocs, 'farm-guardian-architecture.md'), 'utf8')
    const plan = readFileSync(join(repoDocs, 'plans/archive/phase_61_guardian_proactive_nudges.plan.md'), 'utf8')
    expect(existsSync(join(repoRoot, 'internal/farmguardian/nudge.go'))).toBe(true)
    expect(existsSync(join(repoRoot, 'internal/handler/guardian/handler.go'))).toBe(true)
    expect(arch).toContain('guardian-nudge')
    expect(plan).toContain('**Shipped.**')
  })

  it('buildNudgeReviewPayload frames alert context_ref for critical nudges', () => {
    const payload = buildNudgeReviewPayload({
      category: 'critical_alert',
      message: 'Humidity alert — tap to review',
      action_route: '/alerts',
      nudge_id: 'alert-99',
    })
    expect(payload.message).toContain('Humidity alert')
    expect(payload.contextRef.type).toBe('alert')
    expect(payload.contextRef.id).toBe(99)
    expect(payload.contextRef.nudge_category).toBe('critical_alert')
    expect(payload.contextRef.nudge_id).toBe('alert-99')
  })

  it('UI wires nudge dot, strip, and guardianPanel store', () => {
    const edge = readFileSync(join(process.cwd(), 'src/components/GuardianEdgeTab.vue'), 'utf8')
    const top = readFileSync(join(process.cwd(), 'src/components/TopBar.vue'), 'utf8')
    const store = readFileSync(join(process.cwd(), 'src/stores/guardianPanel.js'), 'utf8')
    expect(edge).toContain('guardian-nudge-dot')
    expect(edge).toContain('showNudgeDot')
    expect(top).toContain('topbar-guardian-nudge-dot')
    expect(store).toContain('fetchNudge')
    expect(store).toContain('snoozedNudgeCategories')
  })

  it('context_ref supports nudge_category framing', () => {
    const ctx = readFileSync(join(repoRoot, 'internal/farmguardian/context_ref.go'), 'utf8')
    const nudge = readFileSync(join(repoRoot, 'internal/farmguardian/nudge.go'), 'utf8')
    expect(ctx).toContain('NudgeCategory')
    expect(ctx).toContain('NudgeID')
    expect(nudge).toContain('NudgeContextBlock')
  })

  it('registers GET /farms/{id}/guardian-nudge route', () => {
    const routes = readFileSync(join(repoRoot, 'cmd/api/routes.go'), 'utf8')
    expect(routes).toContain('GET /farms/{id}/guardian-nudge')
  })
})
