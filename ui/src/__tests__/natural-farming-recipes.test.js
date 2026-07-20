/**
 * Phase 209 WS4 — farm recipe apply helpers.
 */
import { describe, it, expect } from 'vitest'
import {
  feedWaterProgramLink,
  formatGrowthStages,
  isLivestockRecipe,
  programsForZone,
  programsUsingRecipe,
} from '../lib/naturalFarmingRecipes.js'

describe('naturalFarmingRecipes', () => {
  it('builds fertigation deep link with recipe and zone', () => {
    expect(feedWaterProgramLink(42, { zoneId: 3 })).toEqual({
      path: '/fertigation',
      query: { tab: 'programs', recipe: '42', zone_id: '3' },
    })
  })

  it('formats growth stage labels', () => {
    const enums = {
      growth_stages: [{ value: 'early_veg', label: 'Early veg' }],
    }
    expect(formatGrowthStages(['early_veg'], enums)).toBe('Early veg')
    expect(formatGrowthStages([], enums)).toBe('—')
  })

  it('finds programs using a recipe in a zone', () => {
    const programs = [
      { id: 1, name: 'Veg', application_recipe_id: 10, target_zone_id: 2 },
      { id: 2, name: 'Other', application_recipe_id: 99, target_zone_id: 2 },
    ]
    const inZone = programsForZone(2, programs, [])
    expect(programsUsingRecipe(10, inZone).map((p) => p.id)).toEqual([1])
  })

  it('detects livestock recipes by application type or input category', () => {
    expect(isLivestockRecipe({ target_application_type: 'livestock_water_supplement' }, {})).toBe(true)
    expect(isLivestockRecipe({ input_definition_id: 5 }, { 5: { category: 'animal_feed' } })).toBe(true)
    expect(isLivestockRecipe({ target_application_type: 'soil_drench' }, {})).toBe(false)
  })
})
