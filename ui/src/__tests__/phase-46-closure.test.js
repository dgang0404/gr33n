/**
 * Phase 46 WS6 / OC-46 — Guardian LLM tool proposals docs closure (Vitest bundle guard).
 * Individual behaviors live in phase-46-ws* tests; this file guards the bundle.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')
const repoDocs = join(repoRoot, 'docs')
const uiSrc = join(process.cwd(), 'src')

describe('Phase 46 WS6 / OC-46 — LLM proposals closure', () => {
  it('README documents Phase 46 shipped and hybrid C flag', () => {
    const readme = readFileSync(join(repoRoot, 'README.md'), 'utf8')
    expect(readme).toContain('Phase 46')
    expect(readme).toContain('GUARDIAN_LLM_PROPOSALS')
    expect(readme).toContain('phase_46_guardian_llm_tool_proposals.plan.md')
    expect(readme).toContain('phase-46-closure.test.js')
    expect(readme).toMatch(/Phase 46.*shipped|hybrid C/i)
  })

  it('guardian-change-requests-guide documents LLM card path §3.3', () => {
    const guide = readFileSync(join(repoDocs, 'guardian-change-requests-guide.md'), 'utf8')
    expect(guide).toContain('### 3.3 When the LLM opens a card (Phase 46 — shipped)')
    expect(guide).toContain('GUARDIAN_LLM_PROPOSALS')
    expect(guide).toContain('meta.llm_sourced')
    expect(guide).toContain('Hybrid C')
  })

  it('operator-tour §6h documents shipped LLM proposal expectations', () => {
    const tour = readFileSync(join(repoDocs, 'operator-tour.md'), 'utf8')
    expect(tour).toContain('### 6h. When Guardian opens a card from your words (Phase 46 — shipped)')
    expect(tour).toContain('GUARDIAN_LLM_PROPOSALS=true')
    expect(tour).toContain('phase-46-closure.test.js')
    expect(tour).not.toContain('### 6h. When Guardian opens a card from your words (Phase 46 — planned)')
  })

  it('architecture §7.0l documents Phase 46 shipped with OC-46', () => {
    const arch = readFileSync(join(repoDocs, 'farm-guardian-architecture.md'), 'utf8')
    expect(arch).toContain('### 7.0l LLM tool proposals (Phase 46 — shipped)')
    expect(arch).toContain('phase-46-closure.test.js')
    expect(arch).toContain('OC-46')
    expect(arch).not.toContain('handler WS3 pending')
  })

  it('guardian PR UX plan marks §8 implemented', () => {
    const prUx = readFileSync(
      join(repoDocs, 'plans/archive/guardian_pr_ux_through_farmer_phases.plan.md'),
      'utf8',
    )
    expect(prUx).toContain('## §8 — LLM structured tool proposals (Phase 46) ✅')
    expect(prUx).toContain('phase-46-closure.test.js')
  })

  it('OC-46 marked completed in operational closure plan', () => {
    const closure = readFileSync(
      join(repoDocs, 'plans/archive/phase_35_37_operational_closure.plan.md'),
      'utf8',
    )
    expect(closure).toContain('oc-46-closure')
    expect(closure).toMatch(/oc-46-closure[\s\S]*status: completed/)
    expect(closure).toContain('## Phase 46 — Guardian LLM tool proposals')
    expect(closure).toContain('OC-46 docs/tests')
  })

  it('phase 46 plan marks all workstreams completed', () => {
    const plan = readFileSync(
      join(repoDocs, 'plans/archive/phase_46_guardian_llm_tool_proposals.plan.md'),
      'utf8',
    )
    for (const id of [
      'ws1-policy',
      'ws2-schema',
      'ws3-handler',
      'ws4-safety',
      'ws5-observability',
      'ws6-docs',
    ]) {
      expect(plan).toContain(`id: ${id}`)
      expect(plan).toMatch(new RegExp(`${id}[\\s\\S]*status: completed`))
    }
    expect(plan).toContain('**Shipped.** WS1–WS6 complete')
  })

  it('closure Vitest bundle files exist', () => {
    for (const f of [
      '__tests__/phase-46-ws1-policy.test.js',
      '__tests__/phase-46-ws2-schema.test.js',
      '__tests__/phase-46-ws3-handler.test.js',
      '__tests__/phase-46-ws4-safety.test.js',
      '__tests__/phase-46-ws5-observability.test.js',
      '__tests__/phase-46-closure.test.js',
    ]) {
      expect(existsSync(join(uiSrc, f))).toBe(true)
    }
  })
})
