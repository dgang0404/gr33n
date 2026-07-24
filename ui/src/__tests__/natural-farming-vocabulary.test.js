import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { NF_VOCAB, NF_WORKSPACE_TAB_LABELS } from '../lib/naturalFarmingVocabulary.js'
import { WORKSPACES } from '../lib/workspaces.js'
import { LIBRARY_TABS } from '../lib/naturalFarmingLibrary.js'
import { operatorConcept } from '../lib/operatorConcepts.js'

describe('naturalFarmingVocabulary', () => {
  it('uses distinct operator terms aligned to DB tables', () => {
    expect(NF_VOCAB.inputs).toBe('Inputs')
    expect(NF_VOCAB.batches).toBe('Batches')
    expect(NF_VOCAB.applyRecipes).toBe('Apply recipes')
    expect(NF_VOCAB.fieldGuide).toBe('Field guide')
    expect(operatorConcept('input_definition')?.label).toBe('Input')
    expect(operatorConcept('application_recipe')?.label).toBe('Apply recipe')
    expect(operatorConcept('nf_field_guide')?.label).toBe('Field guide')
  })

  it('workspace tabs use vocabulary labels (three operational tabs)', () => {
    const labels = WORKSPACES.naturalfarming.tabs.map((t) => t.label)
    expect(labels).toEqual([
      NF_WORKSPACE_TAB_LABELS.batch,
      NF_WORKSPACE_TAB_LABELS.library,
      NF_WORKSPACE_TAB_LABELS.recipes,
    ])
    expect(labels.join(' ')).not.toMatch(/Recipe library|Ready batches|On hand|Recipes & apply/)
  })

  it('field guide sub-tabs name inputs and apply recipes', () => {
    expect(LIBRARY_TABS.map((t) => t.label)).toEqual(['Inputs', 'Apply recipes', 'Programs'])
    expect(LIBRARY_TABS.find((t) => t.id === 'inputs')?.conceptId).toBe('input_definition')
  })

  it('HelpTip teleports to body so overflow chrome cannot clip tooltips', () => {
    const src = readFileSync(join(process.cwd(), 'src/components/HelpTip.vue'), 'utf8')
    expect(src).toContain('<Teleport to="body">')
    expect(src).toContain('fixed z-[9999]')
  })
})
