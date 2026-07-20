/**
 * Phase 209 WS3 — batch flow logic tests.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import {
  processTypesFromCanon,
  variantsForProcess,
  buildInputPayload,
  canonDilutionHint,
} from '../lib/naturalFarmingBatchFlow.js'
import { extractGuideSections, sectionBodyByPrefix } from '../lib/naturalFarmingGuideSections.js'

const jmsGuide = readFileSync(
  join(process.cwd(), '..', 'docs/field-guides/natural-farming-jms.md'),
  'utf8',
)

const canonFixture = {
  inputs: [
    {
      seed_name: 'JMS (JADAM Microbial Solution)',
      process_type: 'jms',
      schema_category: 'microbial_inoculant',
      reference_source: 'JADAM Organic Farming, Youngsang Cho, 2016',
      guide: 'natural-farming-jms.md',
    },
    {
      seed_name: 'JLF General (Weed and Grass)',
      process_type: 'jlf',
      schema_category: 'other_ferment',
      dilution_start: '1:100',
      dilution_strong: '1:20',
      guide: 'natural-farming-jlf-general.md',
    },
    {
      seed_name: 'JLF Spring (Nettle and Comfrey)',
      process_type: 'jlf',
      schema_category: 'other_ferment',
      guide: 'natural-farming-jlf-spring-nettle-comfrey.md',
    },
  ],
}

describe('naturalFarmingBatchFlow', () => {
  it('lists unique process types from canon inputs', () => {
    const types = processTypesFromCanon(canonFixture)
    expect(types.map((t) => t.id).sort()).toEqual(['jlf', 'jms'])
  })

  it('returns JLF variants for jlf process', () => {
    expect(variantsForProcess('jlf', canonFixture)).toHaveLength(2)
  })

  it('buildInputPayload pulls text from guide sections not hardcoded vue', () => {
    const payload = buildInputPayload(canonFixture.inputs[0], jmsGuide)
    expect(payload.name).toBe('JMS (JADAM Microbial Solution)')
    expect(payload.category).toBe('microbial_inoculant')
    expect(payload.typical_ingredients).toMatch(/leaf mold/i)
    const sections = extractGuideSections(jmsGuide)
    expect(payload.preparation_summary).toBe(
      sectionBodyByPrefix(sections, 'Step-by-step preparation').slice(0, 500),
    )
  })

  it('canonDilutionHint uses YAML dilution fields', () => {
    expect(canonDilutionHint(canonFixture.inputs[1])).toContain('1:100')
    expect(canonDilutionHint(canonFixture.inputs[1])).toContain('1:20')
  })
})
