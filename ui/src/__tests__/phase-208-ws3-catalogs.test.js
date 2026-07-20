/**
 * Phase 208 WS3 — YAML catalog closure (recipe-canonical + process-material-catalog).
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')
const recipeCanon = readFileSync(join(repoRoot, 'data/recipe-canonical.yaml'), 'utf8')
const materialCatalog = readFileSync(join(repoRoot, 'data/process-material-catalog.yaml'), 'utf8')
const masterSeed = readFileSync(join(repoRoot, 'db/seeds/master_seed.sql'), 'utf8')

const INPUT_NAMES = [
  'JMS (JADAM Microbial Solution)',
  'LAB (Lactic Acid Bacteria Serum)',
  'FPJ (Fermented Plant Juice)',
  'FFJ (Fermented Fruit Juice)',
  'BRV (Brown Rice Vinegar)',
  'OHN (Oriental Herbal Nutrient)',
  'JHS (JADAM Herbal Solution)',
  'WCA (Water-Soluble Calcium)',
  'WCS (Water-Soluble Calcium Phosphate)',
  'JWA (JADAM Wetting Agent)',
  'JS (JADAM Sulfur concentrate)',
  'JLF General (Weed and Grass)',
  'JLF Crop-Specific (Crop Residue)',
  'JLF Spring (Nettle and Comfrey)',
  'FAA (Fish Amino Acid)',
  'Compost Tea Actively Aerated',
  'Comfrey Slurry (Livestock Supplement)',
  'Sprouted Grain (Livestock Supplement)',
]

/** Phase 211 WS3 — pack/migration seeded; not in master_seed.sql crop canon block. */
const PACK_ONLY_INPUT_NAMES = new Set([
  'Comfrey Slurry (Livestock Supplement)',
  'Sprouted Grain (Livestock Supplement)',
])

const RECIPE_NAMES = [
  'JMS Soil Drench',
  'JLF General Soil Drench',
  'JLF Seedling Drench',
  'JLF and JMS Combined Drench',
  'LAB Soil Conditioner',
  'OHN Pest and Immunity Drench',
  'JMS Foliar Spray',
  'FPJ Vegetative Foliar',
  'FFJ and WCA Flowering Boost',
  'BRV and WCA Cell Strengthener',
  'JHS and JWA Natural Pesticide',
  'JS Fungicide Spray',
  'JLF Foliar Feed',
  'JWA Insecticide Spray',
  'Comfrey Slurry Flock Supplement',
  'Sprouted Grain Treat Batch',
]

describe('Phase 208 WS3 — YAML catalogs', () => {
  it('ships recipe-canonical.yaml and process-material-catalog.yaml', () => {
    expect(recipeCanon).toContain('version: 1')
    expect(materialCatalog).toContain('version: 1')
    expect(recipeCanon).toContain('application_recipes:')
    expect(recipeCanon).toContain('commercial_to_natural:')
    expect(materialCatalog).toContain('materials:')
  })

  it('lists all 18 seed inputs and 16 application recipes', () => {
    expect(recipeCanon.match(/seed_name:/g)?.length).toBe(INPUT_NAMES.length + RECIPE_NAMES.length)
    for (const name of INPUT_NAMES) {
      expect(recipeCanon).toContain(`seed_name: "${name}"`)
      if (!PACK_ONLY_INPUT_NAMES.has(name)) {
        expect(masterSeed).toContain(`'${name}'`)
      }
    }
    for (const name of RECIPE_NAMES) {
      expect(recipeCanon).toContain(`seed_name: "${name}"`)
    }
  })

  it('JMS soil drench is 1:10 not 1:500 in recipe canon', () => {
    const jmsSoil = recipeCanon.match(
      /- seed_name: "JMS Soil Drench"[\s\S]*?(?=\n  - seed_name:|\n# Phase 211)/
    )?.[0]
    expect(jmsSoil).toBeTruthy()
    expect(jmsSoil).toContain('dilution: "1:10')
    expect(jmsSoil).not.toContain('1:500')
    expect(recipeCanon).toMatch(/JLF and JMS Combined Drench[\s\S]*JMS 1:10/)
  })

  it('goldenrod material uses extension_method and 1:100 start', () => {
    expect(materialCatalog).toMatch(/id: goldenrod[\s\S]*source_tier: extension_method/)
    expect(materialCatalog).toMatch(/goldenrod[\s\S]*dilution_start: "1:100"/)
    expect(materialCatalog).toContain('natural-farming-goldenrod-jlf.md')
  })

  it('every canon input links to an on-disk field guide', () => {
    for (const line of recipeCanon.split('\n')) {
      const m = line.match(/^\s+guide: (natural-farming-.+\.md)/)
      if (m) {
        const path = join(repoRoot, 'docs/field-guides', m[1])
        expect(readFileSync(path, 'utf8').length).toBeGreaterThan(100)
      }
    }
  })

  it('switchover maps commercial EC to natural recipe names', () => {
    expect(recipeCanon).toContain('Daily EC veg feed')
    expect(recipeCanon).toContain('JLF and JMS Combined Drench')
    expect(recipeCanon).toContain('Flower boost A+B')
    expect(recipeCanon).toContain('FFJ and WCA Flowering Boost')
  })
})
