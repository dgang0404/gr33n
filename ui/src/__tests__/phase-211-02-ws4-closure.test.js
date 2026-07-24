/**
 * Phase 211.02 WS4 — recipe revision history + restore UI.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const panel = readFileSync(join(process.cwd(), 'src/components/naturalfarming/RecipesApplyPanel.vue'), 'utf8')
const store = readFileSync(join(process.cwd(), 'src/stores/farm.js'), 'utf8')

describe('Phase 211.02 WS4 — recipe revision history UI', () => {
  it('RecipesApplyPanel exposes history panel and restore', () => {
    expect(panel).toContain('data-test="nf-recipe-history"')
    expect(panel).toContain('openRecipeHistory')
    expect(panel).toContain('restoreRevision')
    expect(panel).toContain('loadRecipeRevisions')
  })

  it('farm store calls revisions API', () => {
    expect(store).toContain('loadRecipeRevisions')
    expect(store).toContain('/naturalfarming/recipes/${recipeId}/revisions')
    expect(store).toContain('restoreRecipeRevision')
    expect(store).toContain('/restore')
  })
})
