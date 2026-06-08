/**
 * Phase 64 WS6 / OC-64 — crop knowledge base closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'
import { formatEcTargetChip } from '../lib/growHub.js'

const repoDocs = join(process.cwd(), '..', 'docs')
const repoRoot = join(process.cwd(), '..')

describe('Phase 64 WS6 / OC-64 — crop knowledge base closure', () => {
  it('documents migration, handler, read tool, and crop field guides', () => {
    const arch = readFileSync(join(repoDocs, 'farm-guardian-architecture.md'), 'utf8')
    const plan = readFileSync(join(repoDocs, 'plans/phase_64_crop_knowledge_base.plan.md'), 'utf8')
    expect(existsSync(join(repoRoot, 'db/migrations/20260610_phase64_crop_knowledge_base.sql'))).toBe(true)
    expect(existsSync(join(repoRoot, 'internal/handler/cropprofile/handler.go'))).toBe(true)
    expect(existsSync(join(repoRoot, 'internal/farmguardian/readtools_crop.go'))).toBe(true)
    expect(arch).toContain('Phase 64')
    expect(arch).toContain('lookup_crop_targets')
    expect(plan).toContain('**Shipped.**')
    const manifest = readFileSync(join(repoDocs, 'rag/field-guide-manifest.yaml'), 'utf8')
    expect(manifest).toContain('crop-cannabis-nutrition.md')
  })

  it('start-grow wizard exposes crop profile picker', () => {
    const vue = readFileSync(join(process.cwd(), 'src/components/StartGrowWizard.vue'), 'utf8')
    expect(vue).toContain('data-test="start-grow-crop-profile"')
    expect(vue).toContain('loadCropProfiles')
  })

  it('grow strip renders EC target chip helper', () => {
    const chip = formatEcTargetChip({ ec_min: 1.4, ec_max: 1.8, ec_target: 1.6 })
    expect(chip).toContain('1.4')
    expect(chip).toContain('1.8')
    const strip = readFileSync(join(process.cwd(), 'src/components/ZoneCurrentGrowStrip.vue'), 'utf8')
    expect(strip).toContain('data-test="grow-strip-ec-target"')
  })

  it('crop profile detail route exists', () => {
    const router = readFileSync(join(process.cwd(), 'src/router/index.js'), 'utf8')
    expect(router).toContain('/crop-profiles/:id')
    expect(existsSync(join(process.cwd(), 'src/views/CropProfileDetail.vue'))).toBe(true)
  })
})
