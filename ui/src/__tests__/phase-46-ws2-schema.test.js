/**
 * Phase 46 WS2 — per-tool schema + farm ID binding closure guard.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')
const repoDocs = join(repoRoot, 'docs')
const internal = join(repoRoot, 'internal/farmguardian')

describe('Phase 46 WS2 — schema validation closure', () => {
  it('proposals_llm_validate.go implements schema + bind + reject log', () => {
    const src = readFileSync(join(internal, 'proposals_llm_validate.go'), 'utf8')
    expect(src).toContain('validateLLMProposalSchema')
    expect(src).toContain('bindLLMProposalFarmIDs')
    expect(src).toContain('LogLLMProposalRejected')
    expect(src).toContain('guardian_llm_proposal_rejected')
    expect(src).toContain('GetFertigationProgramByID')
    expect(src).toContain('LLM patch_rule only allows is_active false v1')
  })

  it('TryBuildLLMProposalsFromAssistant logs rejections', () => {
    const src = readFileSync(join(internal, 'proposals_llm.go'), 'utf8')
    expect(src).toContain('LogLLMProposalRejected')
    expect(src).toContain('ValidateLLMProposalDraft(ctx, q, farmID')
  })

  it('Go tests cover schema validation', () => {
    expect(existsSync(join(internal, 'proposals_llm_validate_test.go'))).toBe(true)
  })

  it('phase 46 plan marks WS2 completed', () => {
    const plan = readFileSync(
      join(repoDocs, 'plans/phase_46_guardian_llm_tool_proposals.plan.md'),
      'utf8',
    )
    expect(plan).toContain('id: ws2-schema')
    expect(plan).toMatch(/ws2-schema[\s\S]*status: completed/)
    expect(plan).toContain('proposals_llm_validate.go')
  })

  it('architecture §7.0l notes WS2 validation', () => {
    const arch = readFileSync(join(repoDocs, 'farm-guardian-architecture.md'), 'utf8')
    expect(arch).toContain('proposals_llm_validate.go')
  })
})
