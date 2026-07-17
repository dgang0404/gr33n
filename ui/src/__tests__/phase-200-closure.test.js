/**
 * Phase 200 — accuracy_note round-trip (persist → session reload → UI banner;
 * guardian-eval QA archive).
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { createPinia, setActivePinia } from 'pinia'
import { normalizeSessionTurn, useGuardianChatStore } from '../stores/guardianChat.js'
import { accuracyNoteMessage } from '../lib/guardianCitationLabels.js'

const repoRoot = join(process.cwd(), '..')

describe('Phase 200 — accuracy_note round-trip', () => {
  it('ships migration and handler persist + reload wiring', () => {
    const mig = readFileSync(
      join(repoRoot, 'db/migrations/20260711_phase159_accuracy_note.sql'),
      'utf8',
    )
    expect(mig).toContain('accuracy_note')
    expect(mig).toContain('conversation_turns')

    const handler = readFileSync(join(repoRoot, 'internal/handler/chat/handler.go'), 'utf8')
    expect(handler).toContain('AccuracyNote:     stringPtrOrNil(accuracyNote)')
    expect(handler).toContain('AccuracyNote:     accuracyNote')
    expect(handler).toContain('applyAnswerAccuracyNote(answer, resp.Citations)')
    expect(handler).toContain('applyAnswerAccuracyNote(answer, done.Citations)')
  })

  it('eval runner archives accuracy_note per scored turn', () => {
    const runner = readFileSync(join(repoRoot, 'internal/farmguardian/eval/runner.go'), 'utf8')
    expect(runner).toContain('AccuracyNote string')
    expect(runner).toContain('accuracyNoteFromChatResponse')
    expect(runner).toContain('AccuracyNote:  accuracyNoteFromChatResponse(parsed)')

    const summary = readFileSync(join(repoRoot, 'internal/farmguardian/eval_summary.go'), 'utf8')
    expect(summary).toContain('AccuracyNote            string  `json:"accuracy_note,omitempty"`')
  })

  it('session reload normalizes accuracy_note for the UI banner', () => {
    setActivePinia(createPinia())
    const store = useGuardianChatStore()
    store.setTranscript([
      {
        user_message: 'Why is EC high?',
        assistant_message: 'See [1].',
        accuracy_note: 'citation_number_mismatch',
      },
    ])
    expect(store.transcript[0].accuracy_note).toBe('citation_number_mismatch')
    expect(accuracyNoteMessage(store.transcript[0].accuracy_note)).toBeTruthy()

    const normalized = normalizeSessionTurn({ accuracyNote: 'dangling_list_intro' })
    expect(normalized.accuracy_note).toBe('dangling_list_intro')
  })

  it('GuardianChatPanel renders accuracy banner from reloaded turns', () => {
    const panel = readFileSync(join(repoRoot, 'ui/src/components/GuardianChatPanel.vue'), 'utf8')
    expect(panel).toContain('accuracyNoteMessage(t.accuracy_note)')
    expect(panel).toContain('data-test="chat-accuracy-banner"')
    expect(panel).toContain('setTranscript')
  })

  it('architecture doc no longer claims accuracy_note is unpersisted', () => {
    const arch = readFileSync(join(repoRoot, 'docs/farm-guardian-architecture.md'), 'utf8')
    expect(arch).not.toContain("accuracy_note` isn't persisted")
    expect(arch).toContain('Phase 159 / round-trip verified Phase 200')
  })

  it('phase 200 plan marked shipped', () => {
    const plan = readFileSync(
      join(repoRoot, 'docs/plans/phase_200_accuracy_note_round_trip.plan.md'),
      'utf8',
    )
    expect(plan).toContain('Status:** shipped')
  })
})
