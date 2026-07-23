/**
 * Phase 209 WS4 — Recipes & apply tab wiring.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const workspace = readFileSync(
  join(process.cwd(), 'src/views/workspaces/NaturalFarmingWorkspace.vue'),
  'utf8',
)
const panel = readFileSync(
  join(process.cwd(), 'src/components/naturalfarming/RecipesApplyPanel.vue'),
  'utf8',
)
const lib = readFileSync(join(process.cwd(), 'src/lib/naturalFarmingRecipes.js'), 'utf8')

describe('Phase 209 WS4 — recipes & apply', () => {
  it('recipes tab mounts RecipesApplyPanel', () => {
    expect(workspace).toContain("activeTab === 'recipes'")
    expect(workspace).toContain('RecipesApplyPanel')
  })

  it('panel rehosts recipe CRUD and component editor', () => {
    expect(panel).toContain('data-test="nf-recipes-apply"')
    expect(panel).toContain('createRecipe')
    expect(panel).toContain('updateRecipe')
    expect(panel).toContain('loadRecipeComponents')
    expect(panel).toContain('addRecipeComponent')
    expect(panel).toContain('target_growth_stages')
    expect(panel).toContain('dilution_ratio')
    expect(panel).toContain('target_application_type')
  })

  it('apply panel links zone to fertigation programs', () => {
    expect(panel).toContain('data-test="nf-recipe-apply-panel"')
    expect(panel).toContain('nf-apply-zone')
    expect(panel).toContain('feedWaterProgramLink')
    expect(panel).toContain('nf-apply-feed-water-link')
    expect(panel).toContain('Open Feed &amp; water → Programs')
  })

  it('livestock recipes can link to Animals when module enabled', () => {
    expect(panel).toContain('isLivestockRecipe')
    expect(panel).toContain('MODULE_SCHEMA.animals')
    expect(panel).toContain('nf-apply-animals-link')
    expect(lib).toContain('livestock_water_supplement')
    expect(lib).toContain('animal_feed')
  })

  it('supports deep link ?recipe= on natural-farming tab', () => {
    expect(panel).toContain('route.query.recipe')
    expect(lib).toContain('feedWaterFertigationRoute')
    expect(lib).toContain('recipe: recipeId')
  })
})
