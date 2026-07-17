/**
 * Phase 196 — revision timeline helpers.
 */
import { describe, it, expect } from 'vitest'
import {
  buildRevisionTimeline,
  inferRevisionCue,
  revisionTimelineLabel,
  snippetUserMessage,
} from '../lib/guardianRevisionTimeline.js'

const taskDialogueTurns = [
  { user_message: 'Create a task to refill calcium nitrate when stock is low.' },
  { user_message: 'Put it in Veg Room — that is the zone for this task.' },
  { user_message: 'call it Refill calcium nitrate instead' },
  { user_message: 'make it due tomorrow' },
]

describe('guardianRevisionTimeline (Phase 196)', () => {
  it('builds four user turns from scenario-task-dialogue-pending fixture', () => {
    const entries = buildRevisionTimeline(taskDialogueTurns, { tool: 'create_task' })
    expect(entries).toHaveLength(4)
    expect(entries[0].userMessage).toContain('refill calcium nitrate')
    expect(entries[1].cue).toBe('zone assigned')
    expect(entries[2].cue).toBe('title updated')
    expect(entries[3].cue).toBe('due_date set')
  })

  it('infers volume cue for feed revise turns', () => {
    expect(inferRevisionCue('Please revise — use 0.3 L instead of 0.5.', 'patch_fertigation_program'))
      .toBe('volume updated')
  })

  it('truncates long user snippets', () => {
    const long = 'a'.repeat(120)
    expect(snippetUserMessage(long).length).toBeLessThanOrEqual(96)
    expect(snippetUserMessage(long).endsWith('…')).toBe(true)
  })

  it('labels header with turn count when known', () => {
    expect(revisionTimelineLabel(4, 4)).toBe('Revision history (4 turns)')
  })
})
