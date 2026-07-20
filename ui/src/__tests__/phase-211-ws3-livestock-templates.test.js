/**
 * Phase 211 WS3 — livestock feed templates.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { LIVESTOCK_FEED_TEMPLATES } from '../lib/naturalFarmingLibrary.js'
import { SWITCHOVER_PACK_KEYS } from '../lib/naturalFarmingSwitchover.js'

const repoRoot = join(process.cwd(), '..')
const packJson = readFileSync(
  join(repoRoot, 'data/natural-farming-packs/livestock_comfrey_feed_v1.json'),
  'utf8',
)
const switchoverYaml = readFileSync(
  join(repoRoot, 'data/natural-farming-packs/switchover-packs.yaml'),
  'utf8',
)
const recipeCanon = readFileSync(join(repoRoot, 'data/recipe-canonical.yaml'), 'utf8')
const libraryPanel = readFileSync(
  join(process.cwd(), 'src/components/naturalfarming/RecipeLibraryPanel.vue'),
  'utf8',
)
const switchoverWizard = readFileSync(
  join(process.cwd(), 'src/components/naturalfarming/SwitchoverWizard.vue'),
  'utf8',
)

describe('Phase 211 WS3 — livestock feed templates', () => {
  it('livestock pack JSON defines animal_feed inputs and supplement recipes', () => {
    const pack = JSON.parse(packJson)
    expect(pack.pack_key).toBe('livestock_comfrey_feed_v1')
    expect(pack.input_definitions).toHaveLength(2)
    expect(pack.input_definitions.every((i) => i.category === 'animal_feed')).toBe(true)
    expect(pack.application_recipes.some((r) => r.target_application_type === 'livestock_water_supplement')).toBe(true)
  })

  it('switchover yaml registers livestock_comfrey_feed_v1 with apply_full', () => {
    expect(switchoverYaml).toContain('livestock_comfrey_feed_v1')
    expect(switchoverYaml).toContain('apply_full: true')
    expect(switchoverYaml).toContain('gr33nanimals')
  })

  it('recipe canon includes livestock supplement inputs', () => {
    expect(recipeCanon).toContain('Comfrey Slurry (Livestock Supplement)')
    expect(recipeCanon).toContain('Sprouted Grain (Livestock Supplement)')
    expect(recipeCanon).toContain('schema_category: animal_feed')
  })

  it('library exports livestock template gated by Animals module in UI', () => {
    expect(LIVESTOCK_FEED_TEMPLATES[0].packKey).toBe(SWITCHOVER_PACK_KEYS.LIVESTOCK_COMFREY_FEED_V1)
    expect(libraryPanel).toContain('`nf-library-tab-${t.id}`')
    expect(libraryPanel).toContain("libraryTab === 'livestock'")
    expect(libraryPanel).toContain('MODULE_SCHEMA.animals')
    expect(libraryPanel).toContain('nf-library-apply-livestock-pack')
  })

  it('switchover wizard maps livestock context to feed pack key', () => {
    expect(switchoverWizard).toContain('switchoverPackKey')
    expect(readFileSync(join(process.cwd(), 'src/lib/naturalFarmingSwitchover.js'), 'utf8')).toContain(
      'LIVESTOCK_COMFREY_FEED_V1',
    )
  })
})
