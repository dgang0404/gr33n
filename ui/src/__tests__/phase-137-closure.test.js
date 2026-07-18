/**
 * Phase 137 — Guardian counsel integration (nudges, vision, offline field).
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'
import { buildNudgeReviewPayload, parseAlertIdFromNudgeId } from '../lib/guardianNudge.js'
import { buildOfflineProcedureStarters } from '../lib/guardianStarters.js'

const repoDocs = join(process.cwd(), '..', 'docs')
const repoRoot = join(process.cwd(), '..')

describe('Phase 137 — counsel integration closure', () => {
  it('plan is shipped', () => {
    const plan = readFileSync(join(repoDocs, 'plans/archive/phase_137_guardian_counsel_integration.plan.md'), 'utf8')
    expect(plan).toContain('**Shipped.**')
  })

  it('parseAlertIdFromNudgeId extracts alert id', () => {
    expect(parseAlertIdFromNudgeId('alert-42')).toBe(42)
    expect(parseAlertIdFromNudgeId('feed-missed')).toBeNull()
  })

  it('buildNudgeReviewPayload uses alert context_ref for critical nudges', () => {
    const payload = buildNudgeReviewPayload({
      category: 'critical_alert',
      nudge_id: 'alert-7',
      message: 'High humidity — tap to review',
      action_route: '/alerts',
    })
    expect(payload.contextRef.type).toBe('alert')
    expect(payload.contextRef.id).toBe(7)
  })

  it('buildOfflineProcedureStarters lists field procedures', () => {
    const starters = buildOfflineProcedureStarters()
    expect(starters.some((s) => s.message.includes('wire-pi-relay-light'))).toBe(true)
    expect(starters.some((s) => s.message.includes('diagnose-pi-offline'))).toBe(true)
  })

  it('Go awakening exposes vision_model fields', () => {
    const awakening = readFileSync(join(repoRoot, 'internal/farmguardian/awakening.go'), 'utf8')
    expect(awakening).toContain('VisionModel')
    expect(awakening).toContain('vision_model_loaded')
  })

  it('warmup API accepts include_vision', () => {
    const warmup = readFileSync(join(repoRoot, 'internal/handler/chat/warmup.go'), 'utf8')
    expect(warmup).toContain('include_vision')
  })

  it('UI wires nudge review to Farm counsel warmup and offline banner', () => {
    const nudge = readFileSync(join(process.cwd(), 'src/lib/guardianNudge.js'), 'utf8')
    const readiness = readFileSync(join(process.cwd(), 'src/stores/guardianReadiness.js'), 'utf8')
    const modes = readFileSync(join(process.cwd(), 'src/components/GuardianContextModeCards.vue'), 'utf8')
    const top = readFileSync(join(process.cwd(), 'src/components/TopBar.vue'), 'utf8')
    expect(nudge).toContain("type: 'alert'")
    expect(readiness).toContain('showOfflineFieldBanner')
    expect(readiness).toContain('fieldAssistant')
    expect(modes).toContain('guardian-mode-vision-note')
    expect(modes).toContain('guardian-mode-session-memory-note')
    expect(top).toContain('nudgeDotStirring')
    expect(existsSync(join(repoRoot, 'internal/farmguardian/ollama_contention.go'))).toBe(true)
  })
})
