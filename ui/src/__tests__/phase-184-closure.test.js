/**
 * Phase 184 — multi-turn PR conversation smoke closure (fixture + wiring).
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')

describe('Phase 184 — change-requests-ui scenarios', () => {
  it('fixtures define five multi-turn scenarios with task title revise', () => {
    const fixtures = readFileSync(
      join(repoRoot, 'internal/farmguardian/eval/fixtures_change_requests_ui.go'),
      'utf8',
    )
    expect(fixtures).toContain('scenario-feed-revise-confirm')
    expect(fixtures).toContain('scenario-task-dialogue-pending')
    expect(fixtures).toContain('WantTitle')
    expect(fixtures).toContain('call it Refill calcium nitrate instead')
    expect(fixtures).toContain('ChangeRequestUIScenariosQuick')
  })

  it('scenario runner supports WantTitle and session threading', () => {
    const scenario = readFileSync(
      join(repoRoot, 'internal/farmguardian/eval/scenario.go'),
      'utf8',
    )
    expect(scenario).toContain('WantTitle')
    expect(scenario).toContain('RunQuestionInSession')
    expect(scenario).toContain('resolveScenarioProposal')
  })

  it('Makefile and guardian-eval wire scenario suites', () => {
    const makefile = readFileSync(join(repoRoot, 'Makefile'), 'utf8')
    const main = readFileSync(join(repoRoot, 'cmd/guardian-eval/main.go'), 'utf8')
    expect(makefile).toContain('guardian-qa-change-requests-ui:')
    expect(makefile).toContain('guardian-qa-change-requests-ui-quick:')
    expect(main).toContain('ScenariosForSuite')
    expect(main).toContain('IsScenarioSuite')
  })

  it('ci-guardian-qa documents scenario table', () => {
    const doc = readFileSync(join(repoRoot, 'docs/ci-guardian-qa.md'), 'utf8')
    expect(doc).toContain('scenario-task-dialogue-pending')
    expect(doc).toContain('change-requests-ui-quick')
  })
})
