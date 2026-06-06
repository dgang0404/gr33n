import { describe, it, expect } from 'vitest'
import {
  FARM_SETUP_BLANK_ID,
  FARM_SETUP_PRIMARY_CHOICES,
  farmSetupMoreChoices,
  previewForSetupChoice,
  formatBootstrapApplyResult,
  farmSetupRoute,
} from '../lib/farmSetupWizard.js'
import { BOOTSTRAP_TEMPLATE_KEYS } from '../constants/bootstrapTemplates.js'

describe('Phase 44 WS1 — farm setup wizard helpers', () => {
  it('exposes primary template cards including blank and indoor photoperiod', () => {
    const ids = FARM_SETUP_PRIMARY_CHOICES.map((c) => c.id)
    expect(ids).toContain(FARM_SETUP_BLANK_ID)
    expect(ids).toContain(BOOTSTRAP_TEMPLATE_KEYS.JADAM_INDOOR_PHOTOPERIOD_V1)
    expect(ids).toContain(BOOTSTRAP_TEMPLATE_KEYS.GREENHOUSE_CLIMATE_V1)
  })

  it('lists additional templates not in primary row', () => {
    const more = farmSetupMoreChoices()
    expect(more.some((c) => c.id === BOOTSTRAP_TEMPLATE_KEYS.CHICKEN_COOP_V1)).toBe(true)
    expect(more.every((c) => !FARM_SETUP_PRIMARY_CHOICES.some((p) => p.id === c.id))).toBe(true)
  })

  it('preview for blank explains no auto-created rows', () => {
    const p = previewForSetupChoice(FARM_SETUP_BLANK_ID)
    expect(p.isBlank).toBe(true)
    expect(p.bullets.some((b) => /no zones/i.test(b))).toBe(true)
  })

  it('preview for indoor template includes zone bullets', () => {
    const p = previewForSetupChoice(BOOTSTRAP_TEMPLATE_KEYS.JADAM_INDOOR_PHOTOPERIOD_V1)
    expect(p.isBlank).toBe(false)
    expect(p.bullets.length).toBeGreaterThan(2)
    expect(p.bullets.some((b) => /zones/i.test(b))).toBe(true)
  })

  it('formats bootstrap API outcomes', () => {
    expect(formatBootstrapApplyResult({ applied: true }).ok).toBe(true)
    expect(formatBootstrapApplyResult({ already_applied: true }).message).toMatch(/already applied/i)
    expect(formatBootstrapApplyResult({ error: 'unknown_template' }).ok).toBe(false)
    expect(formatBootstrapApplyResult({ skipped: true }).message).toMatch(/blank/i)
  })

  it('builds setup route for farm id', () => {
    expect(farmSetupRoute(42)).toBe('/farms/42/setup')
  })
})
