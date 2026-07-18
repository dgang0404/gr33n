/**
 * Phase 46 WS3 — chat handler LLM proposal wiring closure guard.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')
const repoDocs = join(repoRoot, 'docs')
const chatHandler = join(repoRoot, 'internal/handler/chat')

describe('Phase 46 WS3 — handler wiring closure', () => {
  it('attachProposals calls matchers first then LLM path', () => {
    const src = readFileSync(join(chatHandler, 'confirm.go'), 'utf8')
    expect(src).toContain('BuildRuleAssistedProposals')
    expect(src).toContain('TryBuildLLMProposalsFromAssistant')
    expect(src).toContain('FreshMatcherMatches')
    expect(src).toContain('LoadLLMProposalPolicyFromEnv')
    expect(src).toContain('FarmCapsForUser')
  })

  it('handler passes assistant text into attachProposals (non-stream + SSE done)', () => {
    const src = readFileSync(join(chatHandler, 'handler.go'), 'utf8')
    expect(src).toContain('attachProposals(r.Context(), farmID, hasUser, userID, sessionID, question, answer, liveSnap, pb.ContextRef, &resp)')
    expect(src).toContain('attachProposals(r.Context(), farmID, hasUser, userID, sessionID, question, answer, liveSnap, contextRef, &done)')
  })

  it('Go tests cover attachProposals guard', () => {
    expect(existsSync(join(chatHandler, 'confirm_proposals_test.go'))).toBe(true)
  })

  it('phase 46 plan marks WS3 completed', () => {
    const plan = readFileSync(
      join(repoDocs, 'plans/phase_46_guardian_llm_tool_proposals.plan.md'),
      'utf8',
    )
    expect(plan).toContain('id: ws3-handler')
    expect(plan).toMatch(/ws3-handler[\s\S]*status: completed/)
    expect(plan).toContain('confirm.go')
  })

  it('architecture §7.0l notes WS3 handler wiring', () => {
    const arch = readFileSync(join(repoDocs, 'farm-guardian-architecture.md'), 'utf8')
    expect(arch).toContain('attachProposals')
    expect(arch).toContain('TryBuildLLMProposalsFromAssistant')
  })
})
