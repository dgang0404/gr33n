/**
 * Phase 209 WS2 — switchover wizard logic tests.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import {
  COMMERCIAL_PATTERNS,
  FIRST_BATCH_SEED_NAMES,
  resolveSwitchoverMapping,
  firstBatchSuggestions,
  fieldGuideDocPath,
} from '../lib/naturalFarmingSwitchover.js'

const repoRoot = join(process.cwd(), '..')
const recipeCanonYaml = readFileSync(join(repoRoot, 'data/recipe-canonical.yaml'), 'utf8')

/** Minimal mirror of recipe-canonical.yaml for unit tests (no yaml parser in ui). */
const canonFixture = {
  commercial_to_natural: [
    {
      commercial: 'Daily EC veg feed 1.6–1.8 mS/cm',
      natural_equivalent: [
        { recipe: 'JLF and JMS Combined Drench', frequency: 'Weekly peak season' },
        { recipe: 'JMS Foliar Spray', frequency: 'Every 1–2 weeks' },
      ],
    },
    {
      commercial: 'Flower boost A+B',
      natural_equivalent: [
        { recipe: 'FFJ and WCA Flowering Boost', frequency: 'Weekly from first buds' },
      ],
    },
    {
      commercial: 'Seedling/cloner light feed',
      natural_equivalent: [
        { recipe: 'JLF Seedling Drench', frequency: 'Weekly through first 2 weeks post-transplant' },
        { recipe: 'JMS Soil Drench', frequency: 'Every 2 weeks' },
      ],
    },
  ],
  application_recipes: [
    {
      seed_name: 'JLF and JMS Combined Drench',
      dilution: 'JLF 1:20 + JMS 1:10 in same water',
      guide: 'natural-farming-application-recipes.md',
    },
    {
      seed_name: 'JMS Foliar Spray',
      dilution: '1:20 (JMS:water) + JWA',
      guide: 'natural-farming-application-recipes.md',
    },
    {
      seed_name: 'FFJ and WCA Flowering Boost',
      dilution: 'FFJ 1:500 + WCA 1:1000 + JWA',
      guide: 'natural-farming-application-recipes.md',
    },
  ],
  inputs: [
    { seed_name: 'JMS (JADAM Microbial Solution)', process_type: 'jms', tradition: 'jadam' },
    {
      seed_name: 'JLF General (Weed and Grass)',
      process_type: 'jlf',
      tradition: 'jadam',
      dilution_start: '1:100',
    },
  ],
}

describe('naturalFarmingSwitchover', () => {
  it('maps single-part EC to JLF+JMS combined and JMS foliar from YAML', () => {
    const mapping = resolveSwitchoverMapping('indoor', 'single_part_ec', canonFixture)
    expect(mapping.commercialLabel).toContain('Daily EC veg feed')
    const names = mapping.naturalEquivalent.map((r) => r.recipe)
    expect(names).toContain('JLF and JMS Combined Drench')
    expect(names).toContain('JMS Foliar Spray')
    const combined = mapping.naturalEquivalent.find((r) => r.recipe === 'JLF and JMS Combined Drench')
    expect(combined?.dilution).toMatch(/JMS 1:10/)
    expect(combined?.dilution).not.toMatch(/1:500/)
  })

  it('maps A+B flower to FFJ and WCA boost', () => {
    const mapping = resolveSwitchoverMapping('greenhouse', 'ab_two_part', canonFixture)
    expect(mapping.naturalEquivalent.some((r) => r.recipe === 'FFJ and WCA Flowering Boost')).toBe(true)
  })

  it('suggests JMS and JLF General as first batches from canon inputs', () => {
    const batches = firstBatchSuggestions(canonFixture)
    expect(batches.map((b) => b.seed_name)).toEqual(FIRST_BATCH_SEED_NAMES)
    expect(batches[0].process_type).toBe('jms')
    expect(batches[1].process_type).toBe('jlf')
  })

  it('every commercial pattern key exists in recipe-canonical.yaml on disk', () => {
    for (const p of COMMERCIAL_PATTERNS) {
      expect(recipeCanonYaml).toContain(p.commercialKey)
    }
  })

  it('field guide learn paths use field-guides/ prefix', () => {
    expect(fieldGuideDocPath('natural-farming-jms.md')).toBe('field-guides/natural-farming-jms.md')
  })
})
