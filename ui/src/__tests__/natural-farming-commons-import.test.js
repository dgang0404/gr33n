import { describe, it, expect } from 'vitest'
import {
  firstBatchQueryForPack,
  formatCommonsImportMessage,
  parseNaturalFarmingPackBody,
} from '../lib/naturalFarmingCommonsImport.js'

describe('naturalFarmingCommonsImport', () => {
  it('returns null for non-NF pack kinds', () => {
    expect(parseNaturalFarmingPackBody({ kind: 'fertigation_recipe_pack' })).toBeNull()
  })

  it('defaults first batch query to batch tab when no known process', () => {
    expect(
      firstBatchQueryForPack({ inputNames: ['Mystery Input'] }),
    ).toEqual({ tab: 'batch' })
  })

  it('explains noop re-import with skip counts', () => {
    expect(
      formatCommonsImportMessage({
        apply: {
          status: 'noop',
          message: 'All pack inputs and recipes already exist on this farm.',
          inputs_skipped: 16,
          recipes_skipped: 14,
        },
      }),
    ).toContain('already on farm')
  })
})
