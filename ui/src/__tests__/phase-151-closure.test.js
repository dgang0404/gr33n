/**
 * Phase 151 — Guardian alert citation enforcement (prompt, eval, detection,
 * alert-only retrieval, post-generation cite injection).
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoDocs = join(process.cwd(), '..', 'docs')
const repoRoot = join(process.cwd(), '..')

describe('Phase 151 — alert citation enforcement', () => {
  it('plan documents all five workstreams', () => {
    const plan = readFileSync(
      join(repoDocs, 'plans/archive/phase_151_guardian_alert_citation_enforcement.plan.md'),
      'utf8',
    )
    expect(plan).toContain('alertCitationDiscipline')
    expect(plan).toContain('MissingNumberedCitationsNote')
    expect(plan).toContain('FilterChunksForAlertSummary')
    expect(plan).toContain('InjectAlertCitationRefs')
    expect(plan).toContain('smoke-unread-alerts')
  })

  it('Go ships prompt override, retrieval filter, cite injection, and eval gate', () => {
    const guardian = readFileSync(join(repoRoot, 'internal/rag/synthesis/guardian.go'), 'utf8')
    expect(guardian).toContain('LIVE FARM STATE')
    expect(guardian).toContain('Do not use markdown links')

    const summary = readFileSync(join(repoRoot, 'internal/farmguardian/alert_summary.go'), 'utf8')
    expect(summary).toContain('func MatchAlertSummaryIntent')
    expect(summary).toContain('func FilterChunksForAlertSummary')

    const inject = readFileSync(join(repoRoot, 'internal/farmguardian/alert_cite_inject.go'), 'utf8')
    expect(inject).toContain('func InjectAlertCitationRefs')

    const handler = readFileSync(join(repoRoot, 'internal/handler/chat/handler.go'), 'utf8')
    expect(handler).toContain('FilterChunksForAlertSummary')

    const finalize = readFileSync(join(repoRoot, 'internal/handler/chat/answer_finalize.go'), 'utf8')
    expect(finalize).toContain('InjectAlertCitationRefs')

    const score = readFileSync(join(repoRoot, 'internal/farmguardian/eval/score.go'), 'utf8')
    expect(score).toContain('smoke-unread-alerts')
    expect(score).toContain('citationRefPresent')

    const acc = readFileSync(join(repoRoot, 'internal/farmguardian/answer_accuracy.go'), 'utf8')
    expect(acc).toContain('MissingNumberedCitationsNote')
    expect(acc).toContain('MultipleCitationsPerListItemNote')

    const normalize = readFileSync(join(repoRoot, 'internal/farmguardian/alert_cite_normalize.go'), 'utf8')
    expect(normalize).toContain('func NormalizeAlertListCitations')

    expect(guardian).toContain('alertOnlyCitationDiscipline')
    expect(guardian).toContain('exactly one [n] citation per numbered list item')
    expect(finalize).toContain('NormalizeAlertListCitations')

    const cite = readFileSync(join(repoRoot, 'internal/farmguardian/answer_citation.go'), 'utf8')
    expect(cite).toContain('gr33ncore')
  })
})
