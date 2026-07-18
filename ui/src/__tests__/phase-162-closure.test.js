/**
 * Phase 162 — Guardian confirm→DB smoke closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')

describe('Phase 162 — confirm→DB smoke closure', () => {
  it('ConfirmProposal client ships in eval', () => {
    const confirm = readFileSync(join(repoRoot, 'internal/farmguardian/eval/confirm.go'), 'utf8')
    expect(confirm).toContain('ConfirmProposal')
    expect(confirm).toContain('/v1/chat/confirm')
  })

  it('side-effect verification ships per fixture', () => {
    const verify = readFileSync(join(repoRoot, 'internal/farmguardian/eval/confirm_verify.go'), 'utf8')
    expect(verify).toContain('write-ack')
    expect(verify).toContain('write-feed')
    expect(verify).toContain('write-schedule')
    expect(verify).toContain('write-task')
    expect(verify).toContain('ListFarmAlerts')
  })

  it('guardian-eval exposes -confirm-proposals flag', () => {
    const main = readFileSync(join(repoRoot, 'cmd/guardian-eval/main.go'), 'utf8')
    expect(main).toContain('confirm-proposals')
    expect(main).toContain('ConfirmAndVerifyPassedProposals')
  })

  it('Makefile ships guardian-qa-change-requests-confirm', () => {
    const makefile = readFileSync(join(repoRoot, 'Makefile'), 'utf8')
    expect(makefile).toContain('guardian-qa-change-requests-confirm')
    expect(makefile).toContain('-confirm-proposals')
  })

  it('phase 162 plan marked shipped', () => {
    const plan = readFileSync(
      join(repoRoot, 'docs/plans/archive/phase_162_guardian_confirm_db_smoke.plan.md'),
      'utf8',
    )
    expect(plan).toContain('Status:** shipped')
  })
})
