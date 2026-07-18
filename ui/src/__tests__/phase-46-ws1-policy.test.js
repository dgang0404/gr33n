/**
 * Phase 46 WS1 — hybrid LLM proposal policy closure (file + plan guard).
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')
const repoDocs = join(repoRoot, 'docs')
const internal = join(repoRoot, 'internal/farmguardian')

describe('Phase 46 WS1 — hybrid policy closure', () => {
  it('proposals_llm.go exists with policy + allowlist + write intent', () => {
    const src = readFileSync(join(internal, 'proposals_llm.go'), 'utf8')
    expect(src).toContain('LoadLLMProposalPolicyFromEnv')
    expect(src).toContain('ShouldAttemptLLMProposal')
    expect(src).toContain('HasWriteIntent')
    expect(src).toContain('IsLLMToolAllowed')
    expect(src).toContain('ParseLLMProposalFromAssistant')
    expect(src).toContain('patch_fertigation_program')
    expect(src).not.toContain('apply_grow_setup_pack": true')
  })

  it('Go tests cover LLM proposal policy', () => {
    expect(existsSync(join(internal, 'proposals_llm_test.go'))).toBe(true)
  })

  it('phase 46 plan marks WS1 completed', () => {
    const plan = readFileSync(
      join(repoDocs, 'plans/archive/phase_46_guardian_llm_tool_proposals.plan.md'),
      'utf8',
    )
    expect(plan).toContain('id: ws1-policy')
    expect(plan).toMatch(/ws1-policy[\s\S]*status: completed/)
    expect(plan).toContain('proposals_llm.go')
  })

  it('.env.example documents GUARDIAN_LLM_PROPOSALS flag', () => {
    const env = readFileSync(join(repoRoot, '.env.example'), 'utf8')
    expect(env).toContain('GUARDIAN_LLM_PROPOSALS')
  })

  it('architecture §7.0l references hybrid policy', () => {
    const arch = readFileSync(join(repoDocs, 'farm-guardian-architecture.md'), 'utf8')
    expect(arch).toContain('### 7.0l LLM tool proposals (Phase 46')
    expect(arch).toContain('proposals_llm.go')
  })
})
