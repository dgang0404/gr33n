/**
 * Phase 209 WS3b — recipe library helpers.
 */
import { describe, it, expect } from 'vitest'
import {
  LIBRARY_PROGRAMS,
  LIBRARY_TABS,
  readyBatchesForComponents,
  traditionBadge,
} from '../lib/naturalFarmingLibrary.js'

describe('naturalFarmingLibrary', () => {
  it('defines inputs, application, and programs tabs', () => {
    expect(LIBRARY_TABS.map((t) => t.id)).toEqual(['inputs', 'application', 'programs'])
    expect(LIBRARY_PROGRAMS[0].bootstrapTemplate).toBe('jadam_indoor_photoperiod_v1')
  })

  it('labels KNF separately from JADAM', () => {
    expect(traditionBadge('knf')?.text).toBe('KNF')
    expect(traditionBadge('jadam')?.text).toBe('JADAM')
  })

  it('finds ready farm batches matching recipe components', () => {
    const rows = readyBatchesForComponents(
      ['JMS (JADAM Microbial Solution)'],
      [{ id: 1, name: 'JMS (JADAM Microbial Solution)' }],
      [{ id: 10, input_definition_id: 1, status: 'ready_for_use', batch_identifier: 'JMS-1' }],
    )
    expect(rows).toHaveLength(1)
    expect(rows[0].batch.batch_identifier).toBe('JMS-1')
  })
})
