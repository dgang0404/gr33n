/**
 * Phase 44 WS8 — Guardian PR slice closure (starters, setup mode, anti-patterns).
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { buildSetupStarters } from '../lib/guardianStarters.js'
import { isFirstRunIncomplete } from '../lib/firstRunChecklist.js'

const repoDocs = join(process.cwd(), '..', 'docs')

const SETUP_SURFACES = [
  'first_run_dashboard',
  'farm_setup_wizard',
  'zone_wizard',
  'device_wizard',
  'empty_zone_grow',
  'setup_mode_chat',
]

function allSetupStarters(params = {}) {
  return SETUP_SURFACES.flatMap((surface) => buildSetupStarters({ surface, farmId: 1, ...params }))
}

describe('Phase 44 WS8 — Guardian PR slice closure', () => {
  it('empty_zone_grow surface offers start-grow chip for room without cycle', () => {
    const starters = buildSetupStarters({
      surface: 'empty_zone_grow',
      farmId: 3,
      zoneCount: 1,
      zones: [{ id: 12, name: 'Flower Room' }],
      zoneName: 'Flower Room',
      activeCycles: [],
    })
    expect(starters.length).toBeLessThanOrEqual(3)
    const grow = starters.find((s) => s.id === 'start-grow')
    expect(grow).toBeTruthy()
    expect(grow.label).toContain('Flower Room')
    expect(grow.message).toContain('philodendron')
    expect(grow.contextRef).toEqual({ type: 'zone', id: 12, name: 'Flower Room' })
    expect(grow.setupMode).toBe(true)
  })

  it('does not promote bootstrap template via any setup starter chip', () => {
    const starters = allSetupStarters({
      zoneCount: 0,
      zones: [],
      deviceWizardStep: true,
    })
    for (const s of starters) {
      const blob = `${s.label} ${s.message}`.toLowerCase()
      expect(blob).not.toMatch(/apply bootstrap|bootstrap template/)
      expect(s.id).not.toBe('apply-bootstrap')
    }
  })

  it('grow-setup starter message matches Phase 32 matcher-friendly wording', () => {
    const starters = buildSetupStarters({
      surface: 'empty_zone_grow',
      farmId: 1,
      zoneCount: 1,
      zones: [{ id: 5, name: 'Veg Room' }],
      zoneName: 'Veg Room',
      activeCycles: [],
    })
    const grow = starters.find((s) => s.id === 'start-grow')
    expect(grow.message).toBe(
      'Add my philodendron to Veg Room with a light fertigation program',
    )
  })

  it('first-run incomplete helper drives optional setup-mode extend', () => {
    expect(isFirstRunIncomplete([
      { done: true },
      { done: false },
    ])).toBe(true)
    expect(isFirstRunIncomplete([{ done: true }, { done: true }])).toBe(false)
  })

  it('guardian PR spec definition of done is marked complete', () => {
    const spec = readFileSync(join(repoDocs, 'plans/phase_44_guardian_pr_spec.md'), 'utf8')
    expect(spec).toContain('status: completed')
    expect(spec).toContain('- [x] Starters on first-run checklist, wizards, empty zone')
    expect(spec).toContain('- [x] Setup-mode hint when `zone_count == 0` or `?setup=1`')
    expect(spec).toContain('- [x] Grow-setup starter triggers existing setup-pack matcher')
    expect(spec).toContain('- [x] Bootstrap **not** promoted via starter chips')
  })
})
