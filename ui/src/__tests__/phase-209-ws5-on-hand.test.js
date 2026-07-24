/**
 * Phase 209 WS5 — batch stock lives on Money → Supplies (Ready batches tab removed).
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const workspace = readFileSync(
  join(process.cwd(), 'src/views/workspaces/NaturalFarmingWorkspace.vue'),
  'utf8',
)
const supplies = readFileSync(join(process.cwd(), 'src/views/SuppliesHub.vue'), 'utf8')
const lib = readFileSync(join(process.cwd(), 'src/lib/naturalFarmingStock.js'), 'utf8')

describe('Phase 209 WS5 — batch stock on Money supplies', () => {
  it('natural farming workspace has no stock tab', () => {
    expect(workspace).not.toContain("activeTab === 'stock'")
    expect(workspace).not.toContain('OnHandPanel')
  })

  it('SuppliesHub shows batches, low-stock banner, and apply-recipe bridge', () => {
    expect(supplies).toContain('buildSupplyRows')
    expect(supplies).toContain('supplies-low-stock-banner')
    expect(supplies).toContain('recipeApplyRouteForStockRow')
    expect(supplies).toContain('supplies-apply-recipe')
    expect(supplies).toContain('OperatorConceptBanner')
  })

  it('stock lib filters ready_for_use and partially_used', () => {
    expect(lib).toContain('ready_for_use')
    expect(lib).toContain('partially_used')
    expect(lib).toContain('listLowStockBatches')
    expect(lib).toContain('buildSupplyRows')
  })
})
