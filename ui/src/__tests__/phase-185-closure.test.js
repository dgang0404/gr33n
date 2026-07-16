/**
 * Phase 185 — create_task zone revise + task dialogue scenario extension.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')

describe('Phase 185 — task zone revise', () => {
  it('proposals_revise applies zone revision cues and numeric zone_id', () => {
    const revise = readFileSync(
      join(repoRoot, 'internal/farmguardian/proposals_revise.go'),
      'utf8',
    )
    expect(revise).toContain('taskZoneRevisionCue')
    expect(revise).toContain('applyTaskZoneRevision')
    expect(revise).toContain('parseTaskZoneIDNumeric')
    expect(revise).toContain('resolveZoneIDForIntent')
  })

  it('task dialogue scenario assigns zone then revises title', () => {
    const fixtures = readFileSync(
      join(repoRoot, 'internal/farmguardian/eval/fixtures_change_requests_ui.go'),
      'utf8',
    )
    expect(fixtures).toContain('RequireTaskZone')
    expect(fixtures).toContain('Put it in Veg Room — that is the zone for this task.')
    expect(fixtures).toContain('call it Refill calcium nitrate instead')
  })

  it('scenario runner asserts RequireTaskZone on final proposal', () => {
    const scenario = readFileSync(
      join(repoRoot, 'internal/farmguardian/eval/scenario.go'),
      'utf8',
    )
    expect(scenario).toContain('RequireTaskZone')
    expect(scenario).toContain('proposal args missing zone_id')
  })
})
