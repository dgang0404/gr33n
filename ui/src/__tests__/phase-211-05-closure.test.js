import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { dirname, join } from 'node:path'
import {
  RECIPE_OUTCOME_MIN_SAMPLE,
  attributeRecipeFromHits,
  attributionHitsFromOpsEvents,
  formatCycleRecipeTrackRecord,
  formatRecipeTrackRecord,
} from '../lib/recipeTrackRecord.js'

const root = join(dirname(fileURLToPath(import.meta.url)), '..')

describe('Phase 211.05 recipe outcome insights closure', () => {
  it('farm store loads recipe-outcomes endpoint', () => {
    const src = readFileSync(join(root, 'stores/farm.js'), 'utf8')
    expect(src).toContain('loadRecipeOutcomes')
    expect(src).toContain('/crop-analytics/recipe-outcomes')
  })

  it('formatRecipeTrackRecord requires min sample and gates cost', () => {
    const outcome = {
      cycle_count: RECIPE_OUTCOME_MIN_SAMPLE,
      avg_yield_grams: 182,
      avg_cost_per_gram: 0.21,
      cost_currency: 'USD',
    }
    expect(formatRecipeTrackRecord(outcome)).toContain('2 harvested cycles')
    expect(formatRecipeTrackRecord(outcome)).toContain('avg 182g')
    expect(formatRecipeTrackRecord(outcome, { showCosts: false })).not.toContain('USD/g')
    expect(formatRecipeTrackRecord({ cycle_count: 1, avg_yield_grams: 100 })).toBe('')
  })

  it('attributes dominant recipe from ops timeline events', () => {
    const hits = attributionHitsFromOpsEvents([
      { kind: 'mix', details: { application_recipe_id: 5, application_recipe_revision_id: 2 } },
      { kind: 'program_run', details: { application_recipe_id: 5, application_recipe_revision_id: 2 } },
      { kind: 'apply', details: {} },
    ])
    const attr = attributeRecipeFromHits(hits)
    expect(attr.mixed).toBe(false)
    expect(attr.application_recipe_id).toBe(5)
    expect(attr.application_recipe_revision_id).toBe(2)
  })

  it('formatCycleRecipeTrackRecord states historical average disclaimer', () => {
    const line = formatCycleRecipeTrackRecord(
      { cycle_count: 4, avg_yield_grams: 182, avg_cost_per_gram: 0.21, cost_currency: 'USD' },
      { recipeLabel: 'JMS Foliar', revisionId: 3 },
    )
    expect(line).toContain('JMS Foliar')
    expect(line).toContain('rev #3')
    expect(line).toContain('4 harvested cycles')
    expect(line).toContain('not a forecast')
  })

  it('RecipesApplyPanel renders track record chip', () => {
    const panel = readFileSync(join(root, 'components/naturalfarming/RecipesApplyPanel.vue'), 'utf8')
    expect(panel).toContain('RecipeTrackRecordChip')
    expect(panel).toContain('loadRecipeOutcomes')
    expect(panel).toContain('FARM_SCOPES.moneyRead')
  })

  it('CropCycleSummary embeds cycle recipe track record', () => {
    const summary = readFileSync(join(root, 'views/CropCycleSummary.vue'), 'utf8')
    expect(summary).toContain('CycleRecipeTrackRecord')
  })

  it('architecture doc cross-links recipe outcome insights', () => {
    const doc = readFileSync(join(root, '../../docs/farm-guardian-architecture.md'), 'utf8')
    expect(doc).toContain('summarize_recipe_outcomes')
    expect(doc).toContain('7.0ah Recipe outcome insights')
  })
})
