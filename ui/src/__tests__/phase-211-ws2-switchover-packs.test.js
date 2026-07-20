/**
 * Phase 211 WS2 — switchover pack apply wiring.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import {
  SWITCHOVER_PACK_KEYS,
  switchoverPackKeyForPattern,
} from '../lib/naturalFarmingSwitchover.js'

const repoRoot = join(process.cwd(), '..')
const switchoverYaml = readFileSync(
  join(repoRoot, 'data/natural-farming-packs/switchover-packs.yaml'),
  'utf8',
)
const wizard = readFileSync(
  join(process.cwd(), 'src/components/naturalfarming/SwitchoverWizard.vue'),
  'utf8',
)
const farmContext = readFileSync(join(process.cwd(), 'src/stores/farmContext.js'), 'utf8')

describe('Phase 211 WS2 — switchover packs', () => {
  it('switchover-packs.yaml defines Mericle veg and flower keys', () => {
    expect(switchoverYaml).toContain('mericle_veg_to_jlf_v1')
    expect(switchoverYaml).toContain('mericle_flower_to_ffj_v1')
    expect(switchoverYaml).toContain('Daily EC veg feed 1.6–1.8 mS/cm')
    expect(switchoverYaml).toContain('Flower boost A+B')
  })

  it('maps commercial patterns to pack keys', () => {
    expect(switchoverPackKeyForPattern('single_part_ec')).toBe(
      SWITCHOVER_PACK_KEYS.MERICLE_VEG_TO_JLF_V1,
    )
    expect(switchoverPackKeyForPattern('ab_two_part')).toBe(
      SWITCHOVER_PACK_KEYS.MERICLE_FLOWER_TO_FFJ_V1,
    )
    expect(switchoverPackKeyForPattern('dry_salts')).toBeNull()
  })

  it('wizard exposes apply switchover pack CTA and store API', () => {
    expect(wizard).toContain('nf-cta-apply-switchover-pack')
    expect(wizard).toContain('applySwitchoverPack')
    expect(wizard).toContain('switchoverPackKey')
    expect(farmContext).toContain('applyNaturalFarmingPack')
    expect(farmContext).toContain('/naturalfarming/apply-pack')
  })
})
