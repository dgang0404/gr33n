/**
 * Phase 46 WS5 — proposal observability log closure guard.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')
const repoDocs = join(repoRoot, 'docs')
const internal = join(repoRoot, 'internal/farmguardian')

describe('Phase 46 WS5 — observability closure', () => {
  it('proposals_observability.go defines matcher, suggested, and rejected logs', () => {
    const src = readFileSync(join(internal, 'proposals_observability.go'), 'utf8')
    expect(src).toContain('LogMatcherProposalHit')
    expect(src).toContain('LogLLMProposalSuggested')
    expect(src).toContain('LogLLMProposalRejected')
    expect(src).toContain('guardian_matcher_proposal_hit')
    expect(src).toContain('guardian_llm_proposal_suggested')
    expect(src).toContain('guardian_llm_proposal_rejected')
  })

  it('matcher and LLM paths emit observability logs on insert', () => {
    const proposals = readFileSync(join(internal, 'proposals.go'), 'utf8')
    const llm = readFileSync(join(internal, 'proposals_llm.go'), 'utf8')
    expect(proposals).toContain('LogMatcherProposalHit')
    expect(llm).toContain('LogLLMProposalSuggested')
    expect(llm).toContain('LogLLMProposalRejected')
  })

  it('Go tests cover observability log shape', () => {
    expect(existsSync(join(internal, 'proposals_observability_test.go'))).toBe(true)
  })

  it('audit playbook documents proposal observability logs', () => {
    const audit = readFileSync(join(repoDocs, 'audit-events-operator-playbook.md'), 'utf8')
    expect(audit).toContain('guardian_matcher_proposal_hit')
    expect(audit).toContain('guardian_llm_proposal_suggested')
    expect(audit).toContain('guardian_llm_proposal_rejected')
  })

  it('phase 46 plan marks WS5 completed', () => {
    const plan = readFileSync(
      join(repoDocs, 'plans/phase_46_guardian_llm_tool_proposals.plan.md'),
      'utf8',
    )
    expect(plan).toContain('id: ws5-observability')
    expect(plan).toMatch(/ws5-observability[\s\S]*status: completed/)
    expect(plan).toContain('proposals_observability.go')
  })
})
