/**
 * Phase 192 — create_task due-date revise must not clobber title.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')

describe('Phase 192 — due-date revise preserves task title', () => {
  it('proposals_revise rejects due-date phrases as title captures', () => {
    const revise = readFileSync(
      join(repoRoot, 'internal/farmguardian/proposals_revise.go'),
      'utf8',
    )
    expect(revise).toContain('func looksLikeDueDatePhrase')
    expect(revise).toContain('!looksLikeDueDatePhrase(title)')
    expect(revise).toContain('Due date before title')
  })

  it('Go tests cover make it due tomorrow preserving title', () => {
    const test = readFileSync(
      join(repoRoot, 'internal/farmguardian/proposals_revise_test.go'),
      'utf8',
    )
    expect(test).toContain('TestApplyRevisionDeltas_CreateTaskDueDate')
    expect(test).toContain('want preserved Refill calcium nitrate')
    expect(test).toContain('TestParseTaskTitleRevision_rejectsDueTomorrowAsTitle')
  })

  it('scenario-task-dialogue-pending still expects Refill calcium nitrate title', () => {
    const fixtures = readFileSync(
      join(repoRoot, 'internal/farmguardian/eval/fixtures_change_requests_ui.go'),
      'utf8',
    )
    expect(fixtures).toContain('WantTitle:               "Refill calcium nitrate"')
    expect(fixtures).toContain('make it due tomorrow')
  })
})
