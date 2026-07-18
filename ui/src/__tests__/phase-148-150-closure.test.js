/**
 * Phases 148–150 — Guardian citation-claim accuracy, alert citation ordering,
 * and dev-jargon answer hygiene, prompted by smoke run #6 findings.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoDocs = join(process.cwd(), '..', 'docs')
const repoRoot = join(process.cwd(), '..')

describe('Phase 148 — citation-claim accuracy', () => {
  it('plan documents the run #6 failure modes', () => {
    const plan = readFileSync(
      join(repoDocs, 'plans/archive/phase_148_guardian_citation_claim_accuracy.plan.md'),
      'utf8',
    )
    expect(plan).toContain('CitationClaimMismatchNote')
    expect(plan).toContain('DuplicateListItemNote')
    expect(plan).toContain('GarbledTokenNote')
    expect(plan).toContain('ECPHUnitConfusionNote')
  })

  it('Go ships all four accuracy detectors wired into the drift scorer', () => {
    const acc = readFileSync(join(repoRoot, 'internal/farmguardian/answer_accuracy.go'), 'utf8')
    expect(acc).toContain('func CitationClaimMismatchNote')
    expect(acc).toContain('func DuplicateListItemNote')
    expect(acc).toContain('func GarbledTokenNote')
    expect(acc).toContain('func ECPHUnitConfusionNote')
    expect(acc).toContain('func AnswerAccuracyNote')
    const drift = readFileSync(join(repoRoot, 'internal/farmguardian/topic_drift.go'), 'utf8')
    expect(drift).toContain('AnswerAccuracyNote')
  })
})

describe('Phase 149 — alert citation ordering', () => {
  it('plan documents severity-first ordering', () => {
    const plan = readFileSync(
      join(repoDocs, 'plans/archive/phase_149_guardian_alert_citation_ordering.plan.md'),
      'utf8',
    )
    expect(plan).toContain('PrioritizeAlertChunks')
    expect(plan).toContain('alertCitationDiscipline')
  })

  it('Go ships alert reordering wired into retrieval and RAG instructions', () => {
    const order = readFileSync(join(repoRoot, 'internal/farmguardian/alert_chunk_order.go'), 'utf8')
    expect(order).toContain('func PrioritizeAlertChunks')
    const handler = readFileSync(join(repoRoot, 'internal/handler/chat/handler.go'), 'utf8')
    expect(handler).toContain('PrioritizeAlertChunks')
    const guardian = readFileSync(join(repoRoot, 'internal/rag/synthesis/guardian.go'), 'utf8')
    expect(guardian).toContain('alertCitationDiscipline')
    expect(guardian).toContain('most severe to least severe')
  })
})

describe('Phase 150 — dev jargon answer hygiene', () => {
  it('plan documents the redaction approach', () => {
    const plan = readFileSync(
      join(repoDocs, 'plans/archive/phase_150_guardian_dev_jargon_hygiene.plan.md'),
      'utf8',
    )
    expect(plan).toContain('RedactDevAPIJargon')
    expect(plan).toContain('PATCH /alerts')
  })

  it('Go ships redaction wired into the finalize pipeline and drift scorer', () => {
    const leak = readFileSync(join(repoRoot, 'internal/farmguardian/answer_leak.go'), 'utf8')
    expect(leak).toContain('func RedactDevAPIJargon')
    expect(leak).toContain('func AnswerContainsDevAPIJargon')
    const finalize = readFileSync(join(repoRoot, 'internal/handler/chat/answer_finalize.go'), 'utf8')
    expect(finalize).toContain('RedactDevAPIJargon')
    const drift = readFileSync(join(repoRoot, 'internal/farmguardian/topic_drift.go'), 'utf8')
    expect(drift).toContain('AnswerContainsDevAPIJargon')
  })
})
