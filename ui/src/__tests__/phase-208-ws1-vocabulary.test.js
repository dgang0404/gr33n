/**
 * Phase 208 WS1 — natural farming process vocabulary closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')
const vocabPath = join(repoRoot, 'data/natural_farming_process_vocabulary.yaml')
const masterSeed = readFileSync(join(repoRoot, 'db/seeds/master_seed.sql'), 'utf8')
const vocab = readFileSync(vocabPath, 'utf8')

const SEED_INPUTS = [
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

describe('Phase 208 WS1 — process vocabulary', () => {
  it('ships vocabulary YAML at data/natural_farming_process_vocabulary.yaml', () => {
    expect(vocab).toContain('version: 1')
    expect(vocab).toContain('traditions:')
    expect(vocab).toContain('source_tiers:')
    expect(vocab).toContain('process_types:')
    expect(vocab).toContain('material_roles:')
    expect(vocab).toContain('schema_category_map:')
    expect(vocab).toContain('seed_inputs:')
  })

  it('defines all four tradition honesty labels', () => {
    for (const key of ['jadam', 'knf', 'extension', 'other']) {
      expect(vocab).toContain(`${key}:`)
    }
  })

  it('maps KNF inputs separately from JADAM core', () => {
    expect(vocab).toMatch(/process_type: fpj[\s\S]*tradition: knf/)
    expect(vocab).toMatch(/process_type: jms[\s\S]*tradition: jadam/)
    expect(vocab).toMatch(/seed_name: "FPJ[\s\S]*tradition: knf/)
    expect(vocab).toMatch(/seed_name: "JMS[\s\S]*tradition: jadam/)
    expect(vocab).toMatch(/seed_name: "Compost Tea[\s\S]*tradition: other/)
  })

  it('covers every post-WS0 master_seed input definition', () => {
    for (const name of SEED_INPUTS) {
      expect(vocab).toContain(`seed_name: "${name}"`)
      expect(masterSeed).toContain(`'${name}'`)
    }
    expect(vocab.match(/seed_name:/g)?.length).toBe(SEED_INPUTS.length)
  })

  it('maps compost tea to Ingham third_party tier, not Cho', () => {
    expect(vocab).toMatch(/Compost Tea Actively Aerated[\s\S]*source_tier: third_party/)
    expect(vocab).toMatch(/compost_tea_aact:[\s\S]*tradition: other/)
  })

  it('schema_category_map keys match declared process_types', () => {
    const processTypeBlock = vocab.slice(vocab.indexOf('process_types:'), vocab.indexOf('material_roles:'))
    const mapBlock = vocab.slice(vocab.indexOf('schema_category_map:'), vocab.indexOf('seed_inputs:'))
    for (const key of ['jms', 'jlf', 'fpj', 'ffj', 'lab', 'ohn', 'faa', 'compost_tea_aact']) {
      expect(processTypeBlock).toContain(`${key}:`)
      expect(mapBlock).toContain(`${key}:`)
    }
  })
})
