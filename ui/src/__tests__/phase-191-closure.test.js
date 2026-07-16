/**
 * Phase 191 — Guardian revise: question-phrased description additions.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')

describe('Phase 191 — Guardian revise question-phrased clarification', () => {
  it('adds reviseDescriptionAppendPattern + parseTaskDescriptionAppendRevision', () => {
    const revise = readFileSync(join(repoRoot, 'internal/farmguardian/proposals_revise.go'), 'utf8')
    expect(revise).toContain('reviseDescriptionAppendPattern')
    expect(revise).toContain('func parseTaskDescriptionAppendRevision')
    expect(revise).toContain('parseTaskDescriptionAppendRevision(question, priorArgs)')
  })

  it('has regression tests built from the live question-phrased correction turn', () => {
    const test = readFileSync(join(repoRoot, 'internal/farmguardian/proposals_revise_test.go'), 'utf8')
    expect(test).toContain('TestApplyRevisionDeltas_CreateTaskDescriptionAppend_questionPhrased')
    expect(test).toContain('TestApplyRevisionDeltas_CreateTaskDescriptionAppend_appendsToExisting')
    expect(test).toContain('TestApplyRevisionDeltas_CreateTaskDescriptionAppend_explicitReplaceStillWins')
  })
})
