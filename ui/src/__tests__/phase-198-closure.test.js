/**
 * Phase 198 — re-run scenario-task-dialogue-pending after Phase 192.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')

describe('Phase 198 — task dialogue eval re-run closure', () => {
  it('Makefile exposes guardian-qa-change-requests-ui-task subset target', () => {
    const makefile = readFileSync(join(repoRoot, 'Makefile'), 'utf8')
    expect(makefile).toContain('guardian-qa-change-requests-ui-task:')
    expect(makefile).toContain('-prompt-ids scenario-task-dialogue-pending')
    expect(makefile).toContain('-fail-on-regression')
  })

  it('ci-guardian-qa documents 2026-07-16 failure and Phase 192 fix', () => {
    const doc = readFileSync(join(repoRoot, 'docs/ci-guardian-qa.md'), 'utf8')
    expect(doc).toContain('Phase 198')
    expect(doc).toContain('proposal title "due tomorrow" want "Refill calcium nitrate"')
    expect(doc).toContain('guardian-qa-change-requests-ui-task')
    expect(doc).toContain('scenario-task-dialogue-pending')
  })

  it('fixture still expects Refill calcium nitrate title and rev 4', () => {
    const fixtures = readFileSync(
      join(repoRoot, 'internal/farmguardian/eval/fixtures_change_requests_ui.go'),
      'utf8',
    )
    expect(fixtures).toContain('scenario-task-dialogue-pending')
    expect(fixtures).toContain('WantTitle:               "Refill calcium nitrate"')
    expect(fixtures).toContain('MinRevision:      4')
    expect(fixtures).toContain('make it due tomorrow')
  })

  it('Phase 192 revise guard rejects due tomorrow as title', () => {
    const revise = readFileSync(
      join(repoRoot, 'internal/farmguardian/proposals_revise.go'),
      'utf8',
    )
    expect(revise).toContain('looksLikeDueDatePhrase')
  })
})
