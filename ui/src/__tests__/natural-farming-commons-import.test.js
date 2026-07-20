import { describe, it, expect } from 'vitest'
import {
  firstBatchQueryForPack,
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
})
