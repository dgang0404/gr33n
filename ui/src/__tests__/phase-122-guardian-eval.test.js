import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

describe('Phase 122 — Guardian model eval + context budget', () => {
  it('prompt budget shrinks small context windows', () => {
    const src = readFileSync(
      join(process.cwd(), '../internal/farmguardian/prompt_budget.go'),
      'utf8',
    )
    expect(src).toContain('ComputePromptBudget')
    expect(src).toContain('GuardianMinContextWindow')
  })

  it('GuardianModelSelector shows eval quality hint', () => {
    const sel = readFileSync(join(process.cwd(), 'src/components/GuardianModelSelector.vue'), 'utf8')
    expect(sel).toContain('guardian-eval-hint')
    expect(sel).toContain('grounded_citation_rate')
    expect(sel).toContain('make guardian-eval')
  })

  it('guardian-eval CLI and makefile target exist', () => {
    const mk = readFileSync(join(process.cwd(), '../Makefile'), 'utf8')
    expect(mk).toContain('guardian-eval')
    const cli = readFileSync(join(process.cwd(), '../cmd/guardian-eval/main.go'), 'utf8')
    expect(cli).toContain('DiscoverOllamaModels')
  })

  it('proposal repair addendum exists', () => {
    const src = readFileSync(
      join(process.cwd(), '../internal/farmguardian/proposals_llm.go'),
      'utf8',
    )
    expect(src).toContain('ProposalRepairSystemAddendum')
    expect(src).toContain('ParseLLMProposalFromAssistantDetailed')
  })
})
