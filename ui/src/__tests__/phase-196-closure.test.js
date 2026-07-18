/**
 * Phase 196 — Pending proposal revision timeline on card.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')

describe('Phase 196 — revision timeline closure', () => {
  it('GuardianActionProposal embeds collapsible revision timeline', () => {
    const card = readFileSync(join(repoRoot, 'ui/src/components/GuardianActionProposal.vue'), 'utf8')
    expect(card).toContain('GuardianProposalRevisionTimeline')
    expect(card).toContain('showRevisionTimeline')
    expect(card).toContain('local.revision || 0) > 1')
  })

  it('timeline component lazy-loads session and exposes data-test hooks', () => {
    const timeline = readFileSync(
      join(repoRoot, 'ui/src/components/GuardianProposalRevisionTimeline.vue'),
      'utf8',
    )
    expect(timeline).toContain('data-test="guardian-proposal-revision-timeline"')
    expect(timeline).toContain('/v1/chat/sessions/${props.sessionId}')
    expect(timeline).toContain('buildRevisionTimeline')
  })

  it('scenario-task-dialogue-pending fixture has four revise turns', () => {
    const fixtures = readFileSync(
      join(repoRoot, 'internal/farmguardian/eval/fixtures_change_requests_ui.go'),
      'utf8',
    )
    expect(fixtures).toContain('scenario-task-dialogue-pending')
    expect(fixtures).toContain('MinRevision:      4')
    expect(fixtures).toContain('make it due tomorrow')
  })
})
