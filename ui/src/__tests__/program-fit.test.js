import { describe, it, expect } from 'vitest'
import {
  programFitResult,
  programFitBadge,
  sortProgramsByFit,
  parseProgramMeta,
} from '../lib/programFit.js'

describe('Phase 102 — program fit helpers', () => {
  const vegProgram = {
    name: 'Veg Daily JLF Program',
    metadata: {
      recommended_crop_keys: ['cannabis', 'tomato'],
      recommended_stages: ['early_veg', 'late_veg'],
      ec_band_mscm: { min: 1.4, max: 2.2 },
    },
  }

  it('flags flower stage on veg-tagged program', () => {
    const fit = programFitResult(vegProgram, { cropKey: 'cannabis', stage: 'early_flower' })
    expect(fit.ok).toBe(false)
    expect(fit.warnings.some((w) => /early_flower|recommended_stages/i.test(w))).toBe(true)
  })

  it('accepts matching veg grow', () => {
    expect(programFitBadge(vegProgram, { cropKey: 'cannabis', stage: 'late_veg' })).toBe('fit')
  })

  it('sorts matching programs first', () => {
    const flowerProgram = {
      name: 'Flower FFJ',
      metadata: { recommended_stages: ['early_flower'] },
    }
    const sorted = sortProgramsByFit([vegProgram, flowerProgram], {
      cropKey: 'cannabis',
      stage: 'early_flower',
    })
    expect(sorted[0].name).toBe('Flower FFJ')
  })

  it('parses EC band from metadata', () => {
    expect(parseProgramMeta(vegProgram.metadata).ec_band_mscm.max).toBe(2.2)
  })
})
