/**
 * Phase 211 WS4 — Commons recipe pack import on Recipes & apply tab.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import {
  NF_COMMONS_TAG,
  firstBatchQueryForPack,
  isNaturalFarmingCatalogEntry,
  parseNaturalFarmingPackBody,
} from '../lib/naturalFarmingCommonsImport.js'

const repoRoot = join(process.cwd(), '..')
const starterPack = readFileSync(
  join(repoRoot, 'data/natural-farming-packs/jadam_indoor_starter_recipes_v1.json'),
  'utf8',
)
const recipesApply = readFileSync(
  join(process.cwd(), 'src/components/naturalfarming/RecipesApplyPanel.vue'),
  'utf8',
)
const commonsImport = readFileSync(
  join(process.cwd(), 'src/components/naturalfarming/CommonsRecipePackImport.vue'),
  'utf8',
)

describe('Phase 211 WS4 — studio Commons import', () => {
  it('parses natural_farming_recipe_pack preview from Commons body', () => {
    const preview = parseNaturalFarmingPackBody(JSON.parse(starterPack))
    expect(preview?.inputCount).toBe(16)
    expect(preview?.recipeCount).toBe(14)
    expect(preview?.inputNames).toContain('JMS (JADAM Microbial Solution)')
  })

  it('filters catalog entries by natural_farming tag', () => {
    expect(
      isNaturalFarmingCatalogEntry({ tags: ['natural_farming', 'recipe-pack'] }),
    ).toBe(true)
    expect(isNaturalFarmingCatalogEntry({ tags: ['fertigation'] })).toBe(false)
  })

  it('suggests JMS batch deep link after starter pack import', () => {
    const preview = parseNaturalFarmingPackBody(JSON.parse(starterPack))
    expect(firstBatchQueryForPack(preview)).toEqual({ tab: 'batch', process: 'jms' })
  })

  it('Recipes tab mounts Commons import with browse, preview, and import CTA', () => {
    expect(recipesApply).toContain('CommonsRecipePackImport')
    expect(commonsImport).toContain('data-test="nf-commons-import"')
    expect(commonsImport).toContain('importCatalogEntry')
    expect(commonsImport).toContain('nf-commons-import-btn')
    expect(commonsImport).toContain('nf-commons-make-first-batch')
    expect(commonsImport).toContain(NF_COMMONS_TAG)
  })
})
