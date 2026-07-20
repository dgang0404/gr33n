/**
 * Phase 208 — Natural farming process & material knowledge (closure).
 * Consolidates WS0–WS5 acceptance criteria; do not modify guardian smoke fixtures here.
 */
import { describe, it, expect } from 'vitest'
import { existsSync, readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')
const guidesDir = join(repoRoot, 'docs/field-guides')
const recipeCanon = readFileSync(join(repoRoot, 'data/recipe-canonical.yaml'), 'utf8')
const materialCatalog = readFileSync(join(repoRoot, 'data/process-material-catalog.yaml'), 'utf8')
const masterSeed = readFileSync(join(repoRoot, 'db/seeds/master_seed.sql'), 'utf8')
const vocab = readFileSync(join(repoRoot, 'data/natural_farming_process_vocabulary.yaml'), 'utf8')
const auditLog = readFileSync(join(guidesDir, 'procedures/recipe-audit-log.md'), 'utf8')
const routes = readFileSync(join(repoRoot, 'cmd/api/routes.go'), 'utf8')
const manifest = readFileSync(join(repoRoot, 'docs/rag/field-guide-manifest.yaml'), 'utf8')

const GUIDES = [
  'natural-farming-jms.md',
  'natural-farming-jlf-general.md',
  'natural-farming-jlf-crop-specific.md',
  'natural-farming-jlf-spring-nettle-comfrey.md',
  'natural-farming-ffj.md',
  'natural-farming-fpj.md',
  'natural-farming-lab.md',
  'natural-farming-ohn.md',
  'natural-farming-wca-wcs.md',
  'natural-farming-jwa-js-jhs.md',
  'natural-farming-brv.md',
  'natural-farming-faa.md',
  'natural-farming-compost-tea-aact.md',
  'natural-farming-application-recipes.md',
  'natural-farming-indoor-photoperiod-program.md',
  'natural-farming-goldenrod-jlf.md',
  'natural-farming-forest-garden-understory.md',
  'natural-farming-livestock-plant-feed.md',
]

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
]

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
]

describe('Phase 208 — closure', () => {
  it('WS0 audit log and migration document JMS 1:500 → 1:10 fix', () => {
    expect(existsSync(join(repoRoot, 'db/migrations/20260720_phase208_ws0_recipe_audit.sql'))).toBe(true)
    expect(auditLog).toContain('JMS Soil Drench')
    expect(auditLog).toContain('1:500')
    expect(auditLog).toContain('**1:10**')
    expect(auditLog).toContain('FAA (Fish Amino Acid)')
    const jmsSoilSeed = masterSeed.match(
      /'JMS Soil Drench'[\s\S]*?'1:10 \(JMS:water\)'/
    )?.[0]
    expect(jmsSoilSeed).toBeTruthy()
    expect(jmsSoilSeed).not.toMatch(/1:500.*JMS:water/)
  })

  it('ships vocabulary, YAML catalogs, and all 16 seed inputs + 14 recipes', () => {
    expect(vocab).toContain('seed_inputs:')
    expect(recipeCanon).toContain('application_recipes:')
    expect(materialCatalog).toContain('materials:')
    for (const name of INPUT_NAMES) {
      expect(recipeCanon).toContain(`seed_name: "${name}"`)
      expect(masterSeed).toContain(`'${name}'`)
    }
    for (const name of RECIPE_NAMES) {
      expect(recipeCanon).toContain(`seed_name: "${name}"`)
    }
  })

  it('goldenrod is extension_method in catalog and guide', () => {
    expect(materialCatalog).toMatch(/id: goldenrod[\s\S]*source_tier: extension_method/)
    const g = readFileSync(join(guidesDir, 'natural-farming-goldenrod-jlf.md'), 'utf8')
    expect(g).toContain('source_tier: extension_method')
    expect(g).not.toContain('source_tier: cho_named')
  })

  it('JMS soil drench is 1:10 not 1:500 in canon, seed, and JMS guide', () => {
    const jmsSoil = recipeCanon.match(
      /- seed_name: "JMS Soil Drench"[\s\S]*?(?=\n  - seed_name:|\n# Phase 211)/
    )?.[0]
    expect(jmsSoil).toContain('dilution: "1:10')
    expect(jmsSoil).not.toContain('1:500')
    const jmsGuide = readFileSync(join(guidesDir, 'natural-farming-jms.md'), 'utf8')
    expect(jmsGuide).toContain('1:10')
    expect(jmsGuide).toContain('1:20')
    expect(jmsGuide).not.toMatch(/1:500.*JMS/i)
  })

  it('every natural-farming guide exists on disk with domain tag', () => {
    for (const f of GUIDES) {
      expect(existsSync(join(guidesDir, f)), f).toBe(true)
      const text = readFileSync(join(guidesDir, f), 'utf8')
      expect(text).toContain('domain: natural_farming')
    }
    expect(GUIDES.length).toBeGreaterThanOrEqual(18)
  })

  it('RAG manifest + DB migration + read API are wired', () => {
    for (const f of GUIDES) {
      expect(manifest).toContain(`  - ${f}`)
    }
    expect(existsSync(join(repoRoot, 'db/migrations/20260720_phase208_ws4_natural_farming_field_guides.sql'))).toBe(true)
    expect(routes).toContain('GET /v1/field-guides/process-catalog')
    expect(routes).toContain('GET /v1/field-guides/process-catalog/materials/{id}')
    expect(routes).toContain('GET /v1/field-guides/recipe-canon')
  })

  it('FAA input has seed row and field guide', () => {
    expect(masterSeed).toContain("'FAA (Fish Amino Acid)'")
    expect(existsSync(join(guidesDir, 'natural-farming-faa.md'))).toBe(true)
    expect(recipeCanon).toContain('seed_name: "FAA (Fish Amino Acid)"')
  })
})
