/**
 * Phase 145 — Guardian topic drift & grounding depth closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoDocs = join(process.cwd(), '..', 'docs')
const repoRoot = join(process.cwd(), '..')

describe('Phase 145 — topic drift & grounding closure', () => {
  it('plan is shipped', () => {
    const plan = readFileSync(
      join(repoDocs, 'plans/phase_145_guardian_topic_drift_and_grounding.plan.md'),
      'utf8',
    )
    expect(plan).toContain('**Shipped.**')
    expect(plan).toContain('ws6-smoke-run4-closure')
  })

  it('smoke report documents Phase 145 run #4', () => {
    const report = readFileSync(join(repoDocs, 'guardian-qa-smoke-report-20260707.md'), 'utf8')
    expect(report).toContain('Phase 145')
    expect(report).toContain('run #4')
  })

  it('architecture documents topic drift and grounding §8.9', () => {
    const arch = readFileSync(join(repoDocs, 'farm-guardian-architecture.md'), 'utf8')
    expect(arch).toContain('### 8.9 Topic drift & grounding (Phase 145)')
    expect(arch).toContain('SmokeTopicDriftNote')
    expect(arch).toContain('answer_relevance.go')
    expect(arch).toContain('answer_citation_align.go')
    expect(arch).toContain('rag_filter.go')
  })

  it('runbook documents Phase 145 drift notes', () => {
    const runbook = readFileSync(join(repoDocs, 'guardian-feedback-review-runbook.md'), 'utf8')
    expect(runbook).toContain('Phase 145 drift notes')
    expect(runbook).toContain('low_relevance')
    expect(runbook).toContain('citation_misaligned')
  })

  it('Go ships relevance, citation align, RAG filter, and drift scorer', () => {
    const rel = readFileSync(join(repoRoot, 'internal/farmguardian/answer_relevance.go'), 'utf8')
    const cite = readFileSync(join(repoRoot, 'internal/farmguardian/answer_citation_align.go'), 'utf8')
    const rag = readFileSync(join(repoRoot, 'internal/farmguardian/rag_filter.go'), 'utf8')
    const drift = readFileSync(join(repoRoot, 'internal/farmguardian/topic_drift.go'), 'utf8')
    const score = readFileSync(join(repoRoot, 'internal/farmguardian/eval/score.go'), 'utf8')
    const qa = readFileSync(join(process.cwd(), 'src/components/GuardianSettingsQARunCard.vue'), 'utf8')
    expect(rel).toContain('ScoreAnswerRelevanceFromText')
    expect(cite).toContain('CitationAlignmentNote')
    expect(rag).toContain('FilterRAGChunks')
    expect(drift).toContain('SmokeTopicDriftNote')
    expect(score).toContain('applySmokeTopicDrift')
    expect(qa).toContain('showRelevanceCol')
  })
})
