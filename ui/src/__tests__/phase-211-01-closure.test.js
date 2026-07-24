/**
 * Phase 211.01 — Natural farming studio declutter (closure).
 * Note: Switchover guide tab removed in 211.02 WS0; wizard component retained for tests/API.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { WORKSPACES, resolveWorkspaceTab } from '../lib/workspaces.js'

const uiSrc = join(process.cwd(), 'src')
const plan = readFileSync(
  join(process.cwd(), '../docs/plans/phase_211_01_nf_studio_declutter.plan.md'),
  'utf8',
)
const workspace = readFileSync(join(uiSrc, 'views/workspaces/NaturalFarmingWorkspace.vue'), 'utf8')
const shell = readFileSync(join(uiSrc, 'components/WorkspaceShell.vue'), 'utf8')
const recipesApply = readFileSync(join(uiSrc, 'components/naturalfarming/RecipesApplyPanel.vue'), 'utf8')

describe('Phase 211.01 — NF studio declutter', () => {
  it('plan documents Option A acceptance', () => {
    expect(plan).toContain('Switchover guide')
    expect(plan).toContain('Seed farm (optional)')
    expect(plan).toContain('naturalFarmingStudio.js')
  })

  it('operational tabs only; legacy start redirects to batch', () => {
    expect(WORKSPACES.naturalfarming.tabs.map((t) => t.id)).toEqual([
      'batch', 'library', 'recipes',
    ])
    expect(resolveWorkspaceTab('naturalfarming', 'start')).toBe('batch')
    expect(workspace).not.toContain('SwitchoverWizard')
    expect(recipesApply).toContain('CommonsRecipePackImport')
  })

  it('Jump to rail on Zones and Hardware workspaces only', () => {
    expect(shell).toContain("JUMP_RAIL_WORKSPACE_IDS")
    expect(shell).toContain("'hardware'")
  })
})
