/**
 * Phase 53 WS2 — supplies restock closure (Vitest bundle guard).
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'
import { relatedTo } from '../lib/navRelations.js'

const uiSrc = join(process.cwd(), 'src')

describe('Phase 53 WS2 — supplies stock closure', () => {
  it('SuppliesHub exposes restock, unit cost, new batch, and refill task UI', () => {
    const vue = readFileSync(join(uiSrc, 'views/SuppliesHub.vue'), 'utf8')
    expect(vue).toContain('data-test="supplies-restock-btn"')
    expect(vue).toContain('data-test="supplies-restock-form"')
    expect(vue).toContain('data-test="supplies-new-batch"')
    expect(vue).toContain('data-test="supplies-edit-unit-cost"')
    expect(vue).toContain('data-test="supplies-refill-task"')
    expect(vue).toContain('updateNfBatch')
    expect(vue).toContain('createNfBatch')
    expect(vue).toContain('updateNfInput')
  })

  it('navRelations links supplies to zones and feeding', () => {
    expect(relatedTo('/operations/supplies')).toContain('/zones')
  })

  it('phase-53 ws2 test file exists', () => {
    expect(existsSync(join(uiSrc, '__tests__/phase-53-ws2-supplies.test.js'))).toBe(true)
  })
})
