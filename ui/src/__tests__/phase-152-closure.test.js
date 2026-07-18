/**
 * Phase 152 WS1 — Guardian live accuracy guardrails. Phase 148/151's
 * detectors now run on every live chat turn (not just guardian-eval), plus
 * three new detectors (truncated tail, uncited timeline claim, invented
 * assumption math) and a farmer-facing UI banner.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoDocs = join(process.cwd(), '..', 'docs')
const repoRoot = join(process.cwd(), '..')

describe('Phase 152 — live accuracy guardrails', () => {
  it('plan documents WS1 + WS2 shipped and WS2b planned', () => {
    const plan = readFileSync(
      join(repoDocs, 'plans/archive/phase_152_guardian_live_accuracy_guardrails.plan.md'),
      'utf8',
    )
    expect(plan).toContain('applyAnswerAccuracyNote')
    expect(plan).toContain('TruncatedAnswerTailNote')
    expect(plan).toContain('UncitedTimelineClaimNote')
    expect(plan).toContain('InventedAssumptionMathNote')
    expect(plan).toContain('ResolveCitationRoute')
    expect(plan).toContain('WS2b')
  })

  it('Go wires accuracy detectors into the live chat path', () => {
    const finalize = readFileSync(join(repoRoot, 'internal/handler/chat/answer_finalize.go'), 'utf8')
    expect(finalize).toContain('func applyAnswerAccuracyNote')
    expect(finalize).toContain('farmguardian.AnswerAccuracyNote')

    const handler = readFileSync(join(repoRoot, 'internal/handler/chat/handler.go'), 'utf8')
    expect(handler).toContain('applyAnswerAccuracyNote(answer, resp.Citations)')
    expect(handler).toContain('applyAnswerAccuracyNote(answer, done.Citations)')
    expect(handler).toContain('AccuracyNote     string')

    const acc = readFileSync(join(repoRoot, 'internal/farmguardian/answer_accuracy.go'), 'utf8')
    expect(acc).toContain('func TruncatedAnswerTailNote')
    expect(acc).toContain('func UncitedTimelineClaimNote')
    expect(acc).toContain('func InventedAssumptionMathNote')

    const dbg = readFileSync(join(repoRoot, 'internal/farmguardian/turn_debug.go'), 'utf8')
    expect(dbg).toContain('AccuracyNote')
  })

  it('UI maps accuracy notes via guardianCitationLabels', () => {
    const labels = readFileSync(join(repoRoot, 'ui/src/lib/guardianCitationLabels.js'), 'utf8')
    expect(labels).toContain('export function accuracyNoteMessage')
  })

  it('Go resolves and wires citation deep links (WS2)', () => {
    const route = readFileSync(join(repoRoot, 'internal/farmguardian/citation_route.go'), 'utf8')
    expect(route).toContain('func ResolveCitationRoute')
    expect(route).toContain('crop_cycle')
    expect(route).toContain('fertigation_program')
    expect(route).toContain('"task"')

    const synthesis = readFileSync(join(repoRoot, 'internal/rag/synthesis/synthesis.go'), 'utf8')
    expect(synthesis).toContain('Route      string `json:"route,omitempty"`')

    const finalize = readFileSync(join(repoRoot, 'internal/handler/chat/answer_finalize.go'), 'utf8')
    expect(finalize).toContain('func attachCitationRoutes')

    const handler = readFileSync(join(repoRoot, 'internal/handler/chat/handler.go'), 'utf8')
    expect(handler).toContain('attachCitationRoutes(r.Context(), h.q, farmID, resp.Citations)')
    expect(handler).toContain('attachCitationRoutes(r.Context(), h.q, farmID, done.Citations)')
  })
})
