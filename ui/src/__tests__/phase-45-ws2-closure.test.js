/**
 * Phase 45 WS2 — friction backlog closure (dry-run: P0 empty).
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoDocs = join(process.cwd(), '..', 'docs')

describe('Phase 45 WS2 — friction backlog closure', () => {
  it('friction backlog documents P0 empty after dry-run', () => {
    const backlog = readFileSync(
      join(repoDocs, 'workstreams/phase-45-ws2-friction-backlog.md'),
      'utf8',
    )
    expect(backlog).toContain('P0 empty')
    expect(backlog).toMatch(/status: closed/)
  })

  it('dry-run log links friction triage with no P0 blockers', () => {
    const log = readFileSync(
      join(repoDocs, 'workstreams/sit-in-45-dry-run-log.md'),
      'utf8',
    )
    expect(log).toContain('P0')
    expect(log).toContain('empty')
    expect(log).toContain('phase-45-ws2-friction-backlog.md')
  })

  it('phase 45 plan marks WS2 complete', () => {
    const plan = readFileSync(
      join(repoDocs, 'plans/archive/phase_45_farmer_validation_whole_app_polish.plan.md'),
      'utf8',
    )
    expect(plan).toMatch(/ws2-friction-backlog[\s\S]*status: completed/)
  })
})
