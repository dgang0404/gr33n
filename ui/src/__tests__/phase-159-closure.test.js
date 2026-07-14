/**
 * Phase 159 — Guardian citation completeness + accuracy_note persistence.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'

const uiSrc = join(process.cwd(), 'src')
const repoRoot = join(process.cwd(), '..')

describe('Phase 159 — citation completeness closure', () => {
  it('ships migration for accuracy_note on conversation_turns', () => {
    const mig = readFileSync(
      join(repoRoot, 'db/migrations/20260711_phase159_accuracy_note.sql'),
      'utf8',
    )
    expect(mig).toContain('accuracy_note')
    expect(mig).toContain('conversation_turns')
  })

  it('ResolveCitationRoute handles WS2b source types', () => {
    const route = readFileSync(join(repoRoot, 'internal/farmguardian/citation_route.go'), 'utf8')
    for (const needle of [
      'case "schedule"',
      'case "alert_notification"',
      'case "field_guide", "platform_doc"',
      'GetRagChunkMetadataByFarmSource',
      'GetFertigationProgramZoneBySchedule',
      'GetLightingProgramZoneBySchedule',
      'zoneFromScheduleNameHint',
    ]) {
      expect(route).toContain(needle)
    }
  })

  it('chat handler persists and reloads accuracy_note', () => {
    const handler = readFileSync(join(repoRoot, 'internal/handler/chat/handler.go'), 'utf8')
    expect(handler).toContain('AccuracyNote:     stringPtrOrNil(accuracyNote)')
    expect(handler).toContain('AccuracyNote:     accuracyNote')
    expect(handler).toContain('attachCitationRoutes(r.Context(), h.q, farmID, cites)')
  })

  it('doc citation landing pages use readable doc view', () => {
    expect(readFileSync(join(uiSrc, 'views/FarmKnowledge.vue'), 'utf8')).toContain('CitationDocView')
    expect(readFileSync(join(uiSrc, 'views/OperatorGuide.vue'), 'utf8')).toContain('CitationDocView')
    expect(readFileSync(join(uiSrc, 'components/CitationDocView.vue'), 'utf8')).toContain('data-test="citation-doc-view"')
  })

  it('current-state reflects 154-158 shipped', () => {
    const cs = readFileSync(join(repoRoot, 'docs/current-state.md'), 'utf8')
    expect(cs).toContain('154–158')
    expect(cs).not.toContain('158 accessibility | Dedicated a11y pass not started')
  })

  it('phase 159 plan marked shipped', () => {
    const plan = readFileSync(
      join(repoRoot, 'docs/plans/phase_159_guardian_citation_completeness.plan.md'),
      'utf8',
    )
    expect(plan).toContain('Status:** shipped')
  })
})
