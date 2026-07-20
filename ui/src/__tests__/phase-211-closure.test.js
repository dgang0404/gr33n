/**
 * Phase 211 — Switchover packs & Commons recipe import (closure).
 * Consolidates WS1–WS4 + WS6; WS5 smoke promotion is optional and gated.
 */
import { describe, it, expect } from 'vitest'
import { existsSync, readFileSync } from 'node:fs'
import { join } from 'node:path'

const uiSrc = join(process.cwd(), 'src')
const repoRoot = join(process.cwd(), '..')
const repoDocs = join(repoRoot, 'docs')
const plan = readFileSync(
  join(repoDocs, 'plans/phase_211_natural_farming_switchover_commons.plan.md'),
  'utf8',
)
const playbooks = readFileSync(join(repoDocs, 'pattern-playbooks.md'), 'utf8')
const switchoverYaml = readFileSync(
  join(repoRoot, 'data/natural-farming-packs/switchover-packs.yaml'),
  'utf8',
)
const starterPack = readFileSync(
  join(repoRoot, 'data/natural-farming-packs/jadam_indoor_starter_recipes_v1.json'),
  'utf8',
)
const applyPackGo = readFileSync(
  join(repoRoot, 'internal/handler/naturalfarming/apply_pack.go'),
  'utf8',
)
const nfPackGo = readFileSync(join(repoRoot, 'internal/naturalfarmingpack/pack.go'), 'utf8')
const commonsPackGo = readFileSync(join(repoRoot, 'internal/commonscatalog/pack.go'), 'utf8')
const routesGo = readFileSync(join(repoRoot, 'cmd/api/routes.go'), 'utf8')
const wizard = readFileSync(
  join(uiSrc, 'components/naturalfarming/SwitchoverWizard.vue'),
  'utf8',
)
const commonsImport = readFileSync(
  join(uiSrc, 'components/naturalfarming/CommonsRecipePackImport.vue'),
  'utf8',
)
const library = readFileSync(
  join(uiSrc, 'components/naturalfarming/RecipeLibraryPanel.vue'),
  'utf8',
)
const smokeEval = readFileSync(
  join(repoRoot, 'internal/farmguardian/eval/fixtures_smoke.go'),
  'utf8',
)
const scoreGo = readFileSync(join(repoRoot, 'internal/farmguardian/eval/score.go'), 'utf8')

describe('Phase 211 — closure', () => {
  it('plan marks WS1–WS4 and WS6 shipped; WS5 optional remains pending', () => {
    expect(plan).toMatch(/Shipped \(WS1–WS4, WS6\)/)
    for (const id of [
      'ws1-recipe-packs',
      'ws2-switchover-packs',
      'ws3-livestock-templates',
      'ws4-studio-import',
      'ws6-tests-docs',
    ]) {
      expect(plan).toContain(`id: ${id}`)
      expect(plan).toMatch(new RegExp(`${id}[\\s\\S]*status: completed`))
    }
    expect(plan).toMatch(/ws5-smoke-promotion[\s\S]*status: pending/)
  })

  it('Commons natural_farming_recipe_pack kind and starter JSON exist', () => {
    const pack = JSON.parse(starterPack)
    expect(pack.kind).toBe('natural_farming_recipe_pack')
    expect(pack.input_definitions.length).toBeGreaterThanOrEqual(14)
    expect(commonsPackGo).toContain('KindNaturalFarmingRecipePack')
    expect(existsSync(join(repoRoot, 'internal/commonscatalog/apply_naturalfarming.go'))).toBe(true)
    expect(existsSync(join(repoRoot, 'internal/commonscatalog/naturalfarming_pack_test.go'))).toBe(true)
  })

  it('switchover packs YAML and apply-pack API route are wired', () => {
    expect(switchoverYaml).toContain('mericle_veg_to_jlf_v1')
    expect(switchoverYaml).toContain('mericle_flower_to_ffj_v1')
    expect(switchoverYaml).toContain('livestock_comfrey_feed_v1')
    expect(nfPackGo).toContain('ApplyPack')
    expect(applyPackGo).toContain('ApplyPack')
    expect(routesGo).toContain('POST /farms/{id}/naturalfarming/apply-pack')
  })

  it('studio Start tab exposes Commons import and switchover apply CTAs', () => {
    expect(wizard).toContain('CommonsRecipePackImport')
    expect(wizard).toContain('nf-cta-apply-switchover-pack')
    expect(commonsImport).toContain('nf-commons-import')
    expect(readFileSync(join(uiSrc, 'lib/naturalFarmingCommonsImport.js'), 'utf8')).toContain(
      'natural_farming_recipe_pack',
    )
  })

  it('livestock template gated by Animals module in recipe library', () => {
    expect(library).toContain("libraryTab === 'livestock'")
    expect(library).toContain('MODULE_SCHEMA.animals')
    expect(library).toContain('nf-library-apply-livestock-pack')
  })

  it('pattern-playbooks documents Commons import and switchover packs', () => {
    expect(playbooks).toContain('Natural farming recipe packs (Phase 211')
    expect(playbooks).toContain('jadam-indoor-starter-recipes-v1')
    expect(playbooks).toContain('mericle_veg_to_jlf_v1')
    expect(playbooks).toContain('phase-211-closure.test.js')
  })

  it('API smoke covers Commons NF import and switchover idempotency', () => {
    const smoke = readFileSync(join(repoRoot, 'cmd/api/smoke_phase211_test.go'), 'utf8')
    expect(smoke).toContain('TestPhase211CommonsNaturalFarmingRecipePackImport')
    expect(smoke).toContain('TestPhase211SwitchoverPackApplyIdempotent')
    expect(smoke).toContain('jadam-indoor-starter-recipes-v1')
    expect(smoke).toContain('mericle_veg_to_jlf_v1')
  })

  it('does not modify guardian smoke fixtures (steps 1–4 unchanged)', () => {
    expect(smokeEval).toContain('smoke-cherry-forest')
    expect(smokeEval).toContain('Grounded: false')
    expect(scoreGo).toContain(`in.Question.ID == "smoke-cherry-forest"`)
    expect(scoreGo).not.toContain('smoke-cherry-jlf')
  })
})
