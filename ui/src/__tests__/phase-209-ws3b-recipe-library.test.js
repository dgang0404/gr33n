/**
 * Phase 209 WS3b — recipe library tab wiring.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const workspace = readFileSync(
  join(process.cwd(), 'src/views/workspaces/NaturalFarmingWorkspace.vue'),
  'utf8',
)
const panel = readFileSync(
  join(process.cwd(), 'src/components/naturalfarming/RecipeLibraryPanel.vue'),
  'utf8',
)
const recipeCanon = readFileSync(join(process.cwd(), '..', 'data/recipe-canonical.yaml'), 'utf8')

describe('Phase 209 WS3b — recipe library', () => {
  it('library tab mounts RecipeLibraryPanel', () => {
    expect(workspace).toContain("activeTab === 'library'")
    expect(workspace).toContain('RecipeLibraryPanel')
  })

  it('panel has inputs, application, and programs subtabs', () => {
    expect(panel).toContain('data-test="nf-recipe-library"')
    expect(panel).toContain('nf-library-tab-${t.id}')
    expect(panel).toContain('LIBRARY_TABS')
    expect(panel).toContain('loadRecipeCanon')
    expect(panel).toContain('loadFieldGuideBody')
    expect(panel).toContain('GuideStepCards')
  })

  it('shows canon-backed dilution and make-batch deep link', () => {
    expect(panel).toContain('canonDilutionHint')
    expect(panel).toContain('nf-library-make-batch-link')
    expect(panel).toContain("tab: 'batch', process:")
  })

  it('programs tab documents bootstrap template and feed-water link', () => {
    const lib = readFileSync(join(process.cwd(), 'src/lib/naturalFarmingLibrary.js'), 'utf8')
    expect(lib).toContain('jadam_indoor_photoperiod_v1')
    expect(panel).toContain('nf-library-feed-water-link')
    expect(panel).toContain("tab: 'programs'")
  })

  it('canon on disk has 16 inputs and 14 application recipes', () => {
    expect(recipeCanon.match(/^  - seed_name:/gm)?.length).toBe(30)
    const inputsBlock = recipeCanon.slice(0, recipeCanon.indexOf('application_recipes:'))
    expect(inputsBlock.match(/^  - seed_name:/gm)?.length).toBe(16)
    expect(recipeCanon.match(/application_recipes:[\s\S]*?(?=^# Phase 211|^commercial_to_natural:)/m)?.[0].match(/^  - seed_name:/gm)?.length).toBe(14)
  })
})
