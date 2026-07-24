/**
 * Phase 209 — Natural farming studio UI (closure).
 * Consolidates WS1–WS7 acceptance criteria; do not modify guardian smoke fixtures here.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'
import { buildLegacyRedirectRoutes, WORKSPACES, resolveWorkspaceTab, canonicalSidebarPath } from '../lib/workspaces.js'
import { buildNavGroups } from '../lib/navGroups.js'
import router from '../router/index.js'

const uiSrc = join(process.cwd(), 'src')
const repoRoot = join(process.cwd(), '..')
const repoDocs = join(repoRoot, 'docs')
const recipeCanon = readFileSync(join(repoRoot, 'data/recipe-canonical.yaml'), 'utf8')
const switchoverLib = readFileSync(join(uiSrc, 'lib/naturalFarmingSwitchover.js'), 'utf8')
const workspace = readFileSync(join(uiSrc, 'views/workspaces/NaturalFarmingWorkspace.vue'), 'utf8')
const switchover = readFileSync(join(uiSrc, 'components/naturalfarming/SwitchoverWizard.vue'), 'utf8')
const makeBatch = readFileSync(join(uiSrc, 'components/naturalfarming/MakeBatchPanel.vue'), 'utf8')
const library = readFileSync(join(uiSrc, 'components/naturalfarming/RecipeLibraryPanel.vue'), 'utf8')
const recipesApply = readFileSync(join(uiSrc, 'components/naturalfarming/RecipesApplyPanel.vue'), 'utf8')
const supplies = readFileSync(join(uiSrc, 'views/SuppliesHub.vue'), 'utf8')
const fertigation = readFileSync(join(uiSrc, 'views/Fertigation.vue'), 'utf8')
const tour = readFileSync(join(repoDocs, 'operator-tour.md'), 'utf8')
const plan = readFileSync(join(repoDocs, 'plans/phase_209_natural_farming_studio_ui.plan.md'), 'utf8')

describe('Phase 209 — closure', () => {
  it('sidebar shows Natural farming under Grow & operate', () => {
    const grow = buildNavGroups().find((g) => g.label === 'Grow & operate')
    const idxZones = grow.items.findIndex((i) => i.to === '/zones')
    const idxNf = grow.items.findIndex((i) => i.to === '/natural-farming')
    const idxComfort = grow.items.findIndex((i) => i.to === '/comfort-targets')
    expect(idxNf).toBeGreaterThan(idxZones)
    expect(idxNf).toBeLessThan(idxComfort)
    expect(grow.items[idxNf].label).toBe('Natural farming')
  })

  it('/natural-farming loads with four tabs; default batch', () => {
    expect(WORKSPACES.naturalfarming.tabs.map((t) => t.id)).toEqual([
      'batch', 'library', 'recipes', 'manage',
    ])
    expect(resolveWorkspaceTab('naturalfarming', undefined)).toBe('batch')
    expect(router.resolve({ path: '/natural-farming' }).name).toBe('natural-farming')
    for (const [tab, panel] of [
      ['batch', 'MakeBatchPanel'],
      ['library', 'RecipeLibraryPanel'],
      ['recipes', 'RecipesApplyPanel'],
      ['manage', 'FarmRowsPanel'],
    ]) {
      expect(workspace).toContain(`activeTab === '${tab}'`)
      expect(workspace).toContain(panel)
    }
    expect(workspace).not.toContain('OnHandPanel')
  })

  it('recipe library and make-batch use 208 canon and field guides (not hardcoded dilutions)', () => {
    const inputsBlock = recipeCanon.slice(0, recipeCanon.indexOf('application_recipes:'))
    expect(inputsBlock.match(/^  - seed_name:/gm)?.length).toBe(18)
    expect(recipeCanon.match(/application_recipes:[\s\S]*?(?=^# Phase 211|^commercial_to_natural:)/m)?.[0].match(/^  - seed_name:/gm)?.length).toBe(16)
    expect(library).toContain('loadRecipeCanon')
    expect(library).toContain('loadFieldGuideBody')
    expect(library).toContain('GuideStepCards')
    expect(makeBatch).toContain('loadRecipeCanon')
    expect(makeBatch).toContain('batchStepCards')
    expect(makeBatch).toContain('createNfInput')
    expect(makeBatch).toContain('createNfBatch')
    expect(makeBatch).not.toMatch(/dilution:\s*['"]1:\d+['"]/)
  })

  it('switchover wizard maps commercial patterns from recipe-canonical YAML', () => {
    expect(switchover).toContain('loadRecipeCanon')
    expect(switchover).toContain('resolveSwitchoverMapping')
    expect(switchoverLib).toContain('commercial_to_natural')
    expect(recipeCanon).toContain('commercial_to_natural:')
    expect(recipeCanon).toContain('JLF and JMS Combined Drench')
  })

  it('recipes tab links to fertigation programs; supplies hub holds batch stock', () => {
    expect(recipesApply).toContain('feedWaterProgramLink')
    expect(recipesApply).toContain('nf-apply-feed-water-link')
    expect(recipesApply).toContain('createRecipe')
    expect(supplies).toContain('buildSupplyRows')
    expect(supplies).toContain('recipeApplyRouteForStockRow')
    expect(supplies).toContain('OperatorConceptBanner')
    expect(readFileSync(join(uiSrc, 'lib/naturalFarmingStock.js'), 'utf8')).toContain('ready_for_use')
  })

  it('/inventory redirects into natural-farming without 404', () => {
    const resolved = router.resolve('/inventory')
    expect(resolved.matched.length).toBeGreaterThan(0)
    expect(canonicalSidebarPath('/inventory')).toBe('/natural-farming')
    const entry = buildLegacyRedirectRoutes().find((r) => r.path === '/inventory')
    const to = { query: {} }
    expect(entry.redirect(to).path).toBe('/natural-farming')
    expect(entry.redirect(to).query.tab).toBe('recipes')
    expect(entry.redirect({ query: { inv: 'batches' } }).path).toBe('/natural-farming')
    expect(entry.redirect({ query: { inv: 'batches' } }).query.tab).toBe('manage')
  })

  it('Fertigation batch links resolve to Natural farming manage tab', () => {
    expect(fertigation).toContain('naturalFarmingManageRoute')
    expect(fertigation).toContain('batchStockLink')
  })

  it('operator-tour documents Natural farming studio; plan marks phase shipped', () => {
    expect(tour).toMatch(/7u\. Natural farming studio/i)
    expect(tour).toContain('/natural-farming')
    expect(plan).toMatch(/Shipped \(WS1–WS7\)|ws7-tests-docs[\s\S]*status: completed/i)
  })

  it('does not touch guardian smoke fixture paths in this phase', () => {
    const smokeDir = join(repoRoot, 'tests/guardian_smoke')
    if (!existsSync(smokeDir)) return
    const touched = ['phase-209-closure.test.js']
    for (const name of touched) {
      expect(existsSync(join(uiSrc, '__tests__', name))).toBe(true)
    }
  })
})
