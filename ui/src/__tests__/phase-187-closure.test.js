/**
 * Phase 187 — relative due_date revise matchers.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')

describe('Phase 187 — relative due_date revise', () => {
  it('proposals_revise parses relative due dates', () => {
    const revise = readFileSync(
      join(repoRoot, 'internal/farmguardian/proposals_revise.go'),
      'utf8',
    )
    expect(revise).toContain('parseTaskRelativeDueDateAt')
    expect(revise).toContain('taskDueDateRevisionCue')
    expect(revise).toContain('reviseDueInDaysPattern')
  })

  it('task dialogue scenario uses make it due tomorrow', () => {
    const fixtures = readFileSync(
      join(repoRoot, 'internal/farmguardian/eval/fixtures_change_requests_ui.go'),
      'utf8',
    )
    expect(fixtures).toContain('make it due tomorrow')
    expect(fixtures).toContain('WantDueDateOffsetDays')
  })

  it('scenario runner supports WantDueDateOffsetDays', () => {
    const scenario = readFileSync(
      join(repoRoot, 'internal/farmguardian/eval/scenario.go'),
      'utf8',
    )
    expect(scenario).toContain('WantDueDateOffsetDays')
    expect(scenario).toContain('AddDate(0, 0, sc.WantDueDateOffsetDays)')
  })
})
