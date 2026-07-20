/**
 * Phase 210 — Guardian natural farming integration (closure).
 * Guards read/write tools, LLM allowlist, regression fixture, and smoke safety.
 */
import { describe, it, expect } from 'vitest'
import { existsSync, readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')
const repoDocs = join(repoRoot, 'docs')
const fg = join(repoRoot, 'internal/farmguardian')
const evalDir = join(fg, 'eval')
const toolsDir = join(fg, 'tools')

const READ_TOOLS = [
  'lookup_process_catalog',
  'suggest_process_from_material',
  'summarize_natural_farming_inventory',
]

const DRAFT_TOOLS = [
  'draft_input_definition',
  'draft_application_recipe',
  'draft_input_batch',
]

describe('Phase 210 — Guardian natural farming closure', () => {
  it('plan marks WS1–WS6 shipped', () => {
    const plan = readFileSync(
      join(repoDocs, 'plans/phase_210_natural_farming_guardian_integration.plan.md'),
      'utf8',
    )
    expect(plan).toContain('**Status:** Shipped (WS1–WS6)')
    for (const id of [
      'ws1-read-tools',
      'ws2-read-router',
      'ws3-write-tools',
      'ws4-llm-allowlist',
      'ws5-regression-fixture',
      'ws6-tests-docs',
    ]) {
      expect(plan).toContain(`id: ${id}`)
      expect(plan).toMatch(new RegExp(`${id}[\\s\\S]*status: completed`))
    }
  })

  it('read tools registered and routed through PlanReadTools', () => {
    const readtools = readFileSync(join(fg, 'readtools.go'), 'utf8')
    const router = readFileSync(join(fg, 'readtools_router.go'), 'utf8')
    const nf = readFileSync(join(fg, 'readtools_naturalfarming.go'), 'utf8')
    for (const id of READ_TOOLS) {
      expect(readtools).toContain(`"${id}"`)
      expect(router).toContain(`"${id}"`)
    }
    expect(nf).toContain('SuggestProcessFromMaterial')
    expect(nf).toContain('goldenrod')
    expect(existsSync(join(fg, 'readtools_naturalfarming_test.go'))).toBe(true)
  })

  it('draft write tools registered with Confirm executors', () => {
    const registry = readFileSync(join(toolsDir, 'registry.go'), 'utf8')
    const draft = readFileSync(join(toolsDir, 'naturalfarming_draft.go'), 'utf8')
    for (const id of DRAFT_TOOLS) {
      expect(registry).toContain(`"${id}"`)
    }
    expect(draft).toContain('resolveDraftInputFromCatalog')
    expect(existsSync(join(toolsDir, 'naturalfarming_draft_test.go'))).toBe(true)
  })

  it('LLM allowlist and schema cover natural farming draft tools', () => {
    const llm = readFileSync(join(fg, 'proposals_llm.go'), 'utf8')
    const validate = readFileSync(join(fg, 'proposals_llm_validate.go'), 'utf8')
    for (const id of DRAFT_TOOLS) {
      expect(llm).toContain(`"${id}"`)
      expect(validate).toContain(`case "${id}"`)
    }
    expect(validate).toContain('llmRejectFarmIDArg')
    expect(validate).toContain('unknown material_id')
    expect(llm).toMatch(/\bdraft\b/)
  })

  it('platform context exposes natural farming read tools via ReadToolIDs', () => {
    const readtools = readFileSync(join(fg, 'readtools.go'), 'utf8')
    const ctx = readFileSync(join(fg, 'platform_context.go'), 'utf8')
    expect(ctx).toContain('ReadToolIDs()')
    for (const id of READ_TOOLS) {
      expect(readtools).toContain(`"${id}"`)
    }
  })

  it('regression fixture added; smoke cherry forest untouched', () => {
    const smoke = readFileSync(join(evalDir, 'fixtures_smoke.go'), 'utf8')
    const regression = readFileSync(join(evalDir, 'fixtures_regression.go'), 'utf8')
    const score = readFileSync(join(evalDir, 'score.go'), 'utf8')
    expect(smoke).toContain('smoke-cherry-forest')
    expect(smoke).toContain('Grounded: false')
    expect(regression).toContain('regression-cherry-goldenrod-jlf')
    expect(regression).toContain('suggest_process_from_material')
    expect(score).toContain('scoreRegressionCherryGoldenrodJLF')
    expect(score).toContain(`in.Question.ID == "smoke-cherry-forest"`)
    expect(existsSync(join(evalDir, 'fixtures_regression_test.go'))).toBe(true)
  })

  it('architecture documents Phase 210 natural farming tools', () => {
    const arch = readFileSync(join(repoDocs, 'farm-guardian-architecture.md'), 'utf8')
    expect(arch).toContain('### 7.0ai Natural farming (Phase 210 — shipped)')
    for (const id of [...READ_TOOLS, ...DRAFT_TOOLS]) {
      expect(arch).toContain(`\`${id}\``)
    }
    expect(arch).toContain('phase-210-closure.test.js')
    expect(arch).toContain('regression-cherry-goldenrod-jlf')
    expect(arch).toContain('smoke-cherry-forest')
  })
})
