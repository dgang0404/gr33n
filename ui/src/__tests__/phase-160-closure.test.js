/**
 * Phase 160 — a11y residuals (lighting modal, form labels, mobile drawer).
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const uiSrc = join(process.cwd(), 'src')
const repoDocs = join(process.cwd(), '..', 'docs')

describe('Phase 160 — a11y residuals closure', () => {
  it('ZoneLightingEditor modal is a trapped dialog', () => {
    const editor = readFileSync(join(uiSrc, 'components/ZoneLightingEditor.vue'), 'utf8')
    expect(editor).toContain('useDialogFocusTrap')
    expect(editor).toContain('role="dialog"')
    expect(editor).toContain('aria-labelledby="zone-lighting-modal-title"')
    expect(editor).toContain('zone-lighting-modal-error')
    expect(editor).toMatch(/zone-lighting-modal-error[\s\S]*role="alert"/)
  })

  it('LightingProgramForm wires for/id on high-traffic fields', () => {
    const form = readFileSync(join(uiSrc, 'components/LightingProgramForm.vue'), 'utf8')
    expect(form).toContain('for="lighting-program-name"')
    expect(form).toContain('id="lighting-program-name"')
    expect(form).toContain('for="lighting-program-actuator"')
    expect(form).toContain(':aria-pressed="form.presetKey === p.key"')
    expect(form).toContain('lighting-photoperiod-label')
  })

  it('PhotoperiodClockEditor labels on/off hours', () => {
    const clock = readFileSync(join(uiSrc, 'components/PhotoperiodClockEditor.vue'), 'utf8')
    expect(clock).toContain('for="photoperiod-lights-on"')
    expect(clock).toContain('for="photoperiod-duration"')
    expect(clock).toContain('photoperiod-error')
    expect(clock).toContain(':aria-pressed="activePresetKey === p.key"')
  })

  it('mobile drawer uses focus trap and labelled close', () => {
    const app = readFileSync(join(uiSrc, 'App.vue'), 'utf8')
    expect(app).toContain('useDialogFocusTrap(drawerOpen')
    expect(app).toContain('aria-label="Close navigation menu"')
    expect(app).toContain('aria-label="Navigation menu"')
  })

  it('a11y audit doc records Phase 160 closures', () => {
    const audit = readFileSync(join(repoDocs, 'a11y-audit-2026-07-11.md'), 'utf8')
    expect(audit).toContain('Phase 160')
    expect(audit).toContain('D2')
    expect(audit).toContain('D3')
    expect(audit).toContain('phase-160-closure.test.js')
  })

  it('phase 160 plan marked shipped', () => {
    const plan = readFileSync(
      join(repoDocs, 'plans/phase_160_a11y_residuals.plan.md'),
      'utf8',
    )
    expect(plan).toContain('Status:** shipped')
  })
})
