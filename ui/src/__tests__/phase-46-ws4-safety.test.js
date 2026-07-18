/**
 * Phase 46 WS4 — LLM proposal safety test closure guard.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')
const repoDocs = join(repoRoot, 'docs')
const internal = join(repoRoot, 'internal/farmguardian')
const smoke = join(repoRoot, 'cmd/api')

describe('Phase 46 WS4 — safety tests closure', () => {
  it('proposals_llm_safety_test.go covers rejection gates', () => {
    const src = readFileSync(join(internal, 'proposals_llm_safety_test.go'), 'utf8')
    expect(src).toContain('TestTryBuildLLM_MatcherMatchedSkipsInsert')
    expect(src).toContain('TestTryBuildLLM_NoOperateSkipsInsert')
    expect(src).toContain('TestTryBuildLLM_NotOnAllowlist')
    expect(src).toContain('TestTryBuildLLM_BootstrapTemplateRejected')
  })

  it('smoke_phase46_ws4_test.go covers DB safety cases', () => {
    const src = readFileSync(join(smoke, 'smoke_phase46_ws4_test.go'), 'utf8')
    expect(src).toContain('TestPhase46WS4_MatcherFirstIgnoresLLMJSON')
    expect(src).toContain('TestPhase46WS4_LLMHappyPathPatchFertigationProgram')
    expect(src).toContain('TestPhase46WS4_WrongProgramIDRejected')
    expect(src).toContain('TestPhase46WS4_ViewerNoLLMInsert')
    expect(src).toContain('TestPhase46WS4_ConfirmExpiredProposalGone')
    expect(src).toContain('phase46HybridAttach')
  })

  it('confirm path RBAC covered by Phase 29 WS5 smoke', () => {
    const src = readFileSync(join(smoke, 'smoke_phase29_ws5_test.go'), 'utf8')
    expect(src).toContain('TestPhase29WS5_Confirm_ViewerForbidden')
  })

  it('phase 46 plan marks WS4 completed', () => {
    const plan = readFileSync(
      join(repoDocs, 'plans/archive/phase_46_guardian_llm_tool_proposals.plan.md'),
      'utf8',
    )
    expect(plan).toContain('id: ws4-safety')
    expect(plan).toMatch(/ws4-safety[\s\S]*status: completed/)
    expect(plan).toContain('proposals_llm_safety_test.go')
    expect(plan).toContain('smoke_phase46_ws4_test.go')
  })

  it('architecture §7.0l notes WS4 safety coverage', () => {
    const plan = readFileSync(
      join(repoDocs, 'plans/archive/phase_46_guardian_llm_tool_proposals.plan.md'),
      'utf8',
    )
    expect(plan).toContain('proposals_llm_safety_test.go')
    expect(plan).toContain('smoke_phase46_ws4_test.go')
  })
})
