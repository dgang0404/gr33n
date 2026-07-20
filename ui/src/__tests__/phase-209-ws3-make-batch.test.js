/**
 * Phase 209 WS3 — Make a batch tab wiring.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const workspace = readFileSync(
  join(process.cwd(), 'src/views/workspaces/NaturalFarmingWorkspace.vue'),
  'utf8',
)
const panel = readFileSync(
  join(process.cwd(), 'src/components/naturalfarming/MakeBatchPanel.vue'),
  'utf8',
)
const batchFlow = readFileSync(join(process.cwd(), 'src/lib/naturalFarmingBatchFlow.js'), 'utf8')

describe('Phase 209 WS3 — make a batch', () => {
  it('batch tab mounts MakeBatchPanel', () => {
    expect(workspace).toContain("activeTab === 'batch'")
    expect(workspace).toContain('MakeBatchPanel')
  })

  it('panel implements process → variant → step cards → create flow', () => {
    expect(panel).toContain('data-test="nf-make-batch"')
    expect(panel).toContain('nf-batch-process-picker')
    expect(panel).toContain('nf-batch-variant-picker')
    expect(panel).toContain('nf-batch-step-cards')
    expect(panel).toContain('nf-batch-create-form')
    expect(panel).toContain('batchStepCards')
    expect(panel).toContain('createNfInput')
    expect(panel).toContain('createNfBatch')
  })

  it('loads field guides from commons API and canon from read API', () => {
    expect(panel).toContain('loadRecipeCanon')
    expect(batchFlow).toContain('/commons/agronomy-field-guides/')
    expect(batchFlow).toContain('buildInputPayload')
  })

  it('supports deep link ?process= from switchover wizard', () => {
    expect(panel).toContain('route.query.process')
    expect(panel).toContain('selectProcess')
  })

  it('optional prep task uses jadam_prep task type', () => {
    expect(batchFlow).toContain("task_type: 'jadam_prep'")
    expect(panel).toContain('create_prep_task')
    expect(panel).toContain('createTask')
  })
})
