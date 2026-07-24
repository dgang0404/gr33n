/**
 * Phase 211.02 WS0 — Remove switchover tab and jump rail (closure).
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { WORKSPACES, resolveWorkspaceTab } from '../lib/workspaces.js'

const plan = readFileSync(
  join(process.cwd(), '../docs/plans/phase_211_02_recipe_formula_history.plan.md'),
  'utf8',
)
const workspace = readFileSync(
  join(process.cwd(), 'src/views/workspaces/NaturalFarmingWorkspace.vue'),
  'utf8',
)

describe('Phase 211.02 WS0 — NF workspace trim', () => {
  it('plan documents recipe revision and crop ops reporting', () => {
    expect(plan).toContain('application_recipe_revisions')
    expect(plan).toContain('ops-timeline')
    expect(plan).toContain('formula_snapshot')
  })

  it('four operational tabs; no switchover wizard mount', () => {
    expect(WORKSPACES.naturalfarming.tabs).toHaveLength(4)
    expect(workspace).toContain('MakeBatchPanel')
    expect(workspace).not.toContain('SwitchoverWizard')
    expect(resolveWorkspaceTab('naturalfarming', undefined)).toBe('batch')
    expect(resolveWorkspaceTab('naturalfarming', 'start')).toBe('batch')
  })
})
