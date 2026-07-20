/**
 * Phase 209 WS5 — On hand tab wiring.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const workspace = readFileSync(
  join(process.cwd(), 'src/views/workspaces/NaturalFarmingWorkspace.vue'),
  'utf8',
)
const panel = readFileSync(
  join(process.cwd(), 'src/components/naturalfarming/OnHandPanel.vue'),
  'utf8',
)
const lib = readFileSync(join(process.cwd(), 'src/lib/naturalFarmingStock.js'), 'utf8')

describe('Phase 209 WS5 — on hand', () => {
  it('stock tab mounts OnHandPanel', () => {
    expect(workspace).toContain("activeTab === 'stock'")
    expect(workspace).toContain('OnHandPanel')
  })

  it('panel shows ready batches and low-stock banner', () => {
    expect(panel).toContain('data-test="nf-on-hand"')
    expect(panel).toContain('stockRows')
    expect(panel).toContain('lowStockFromReady')
    expect(panel).toContain('nf-stock-low-stock-banner')
    expect(panel).toContain('loadNfBatches')
  })

  it('bridges to Money for restock and unit costs', () => {
    expect(panel).toContain('moneyTabRoute')
    expect(panel).toContain('nf-stock-money-supplies')
    expect(panel).toContain('Restock / edit costs → Money')
    expect(panel).toContain("moneyTabRoute('supplies')")
  })

  it('links make batch and apply recipe tabs', () => {
    expect(panel).toContain('nf-stock-make-batch')
    expect(panel).toContain("tab: 'batch'")
    expect(panel).toContain("tab: 'recipes'")
  })

  it('stock lib filters ready_for_use and partially_used', () => {
    expect(lib).toContain('ready_for_use')
    expect(lib).toContain('partially_used')
    expect(lib).toContain('listLowStockBatches')
    expect(lib).toContain('buildSupplyRows')
  })
})
