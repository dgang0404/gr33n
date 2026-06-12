/**
 * Phase 99 — CI guards: UI fallback enums stay aligned with backend/OpenAPI.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join, dirname } from 'node:path'
import { fileURLToPath } from 'node:url'
import { FALLBACK_GROWTH_STAGE_VALUES } from '../lib/domainEnums.fallback.js'
import { GROWTH_STAGES } from '../lib/growHub.js'

const __dirname = dirname(fileURLToPath(import.meta.url))
const uiSrc = join(__dirname, '..')

const CANONICAL_GROWTH_STAGES = [
  'clone', 'seedling', 'early_veg', 'late_veg', 'transition',
  'early_flower', 'mid_flower', 'late_flower', 'flush', 'harvest', 'dry_cure',
]

describe('Phase 99 — UI domain parity', () => {
  it('fallback growth stages match canonical 11 (SetpointRow bug guard)', () => {
    expect(FALLBACK_GROWTH_STAGE_VALUES).toHaveLength(11)
    expect(FALLBACK_GROWTH_STAGE_VALUES).toEqual(CANONICAL_GROWTH_STAGES)
    expect(GROWTH_STAGES).toEqual(FALLBACK_GROWTH_STAGE_VALUES)
  })

  it('SetpointRow default uses full fallback growth stage list', () => {
    const src = readFileSync(join(uiSrc, 'components/SetpointRow.vue'), 'utf8')
    expect(src).toContain('FALLBACK_GROWTH_STAGE_VALUES')
    expect(src).toContain('[...FALLBACK_GROWTH_STAGE_VALUES]')
  })

  it('cropLibraryPicker has no dead CATEGORY_ORDER (API picker owns order)', () => {
    const src = readFileSync(join(uiSrc, 'lib/cropLibraryPicker.js'), 'utf8')
    expect(src).not.toMatch(/export const CATEGORY_ORDER/)
  })
})
