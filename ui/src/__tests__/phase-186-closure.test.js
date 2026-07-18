/**
 * Phase 186 — create_task due_date revise + tool wiring closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')

describe('Phase 186 — task due_date revise', () => {
  it('proposals_revise parses due date corrections', () => {
    const revise = readFileSync(
      join(repoRoot, 'internal/farmguardian/proposals_revise.go'),
      'utf8',
    )
    expect(revise).toContain('parseTaskDueDateRevision')
    expect(revise).toContain('reviseDueDatePattern')
    expect(revise).toContain('due_date')
  })

  it('create_task tool passes due_date to CreateTask', () => {
    const tasks = readFileSync(
      join(repoRoot, 'internal/farmguardian/tools/tasks.go'),
      'utf8',
    )
    const args = readFileSync(
      join(repoRoot, 'internal/farmguardian/tools/args.go'),
      'utf8',
    )
    expect(args).toContain('optionalDateFromArgs')
    expect(tasks).toContain('DueDate:')
    expect(tasks).toContain('optionalDateFromArgs(args, "due_date")')
  })

  it('task dialogue scenario adds due-date turn and due-date assertions', () => {
    const fixtures = readFileSync(
      join(repoRoot, 'internal/farmguardian/eval/fixtures_change_requests_ui.go'),
      'utf8',
    )
    const scenario = readFileSync(
      join(repoRoot, 'internal/farmguardian/eval/scenario.go'),
      'utf8',
    )
    expect(fixtures).toContain('WantDueDateOffsetDays')
    expect(fixtures).toContain('MinRevision:      4')
    expect(scenario).toContain('WantDueDate')
    expect(scenario).toContain('WantDueDateOffsetDays')
    expect(scenario).toContain('proposal args missing due_date')
  })
})
