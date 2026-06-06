/**
 * Phase 45 WS7 / OC-45 — farmer polish docs closure (Vitest bundle guard).
 * Individual behaviors live in phase-45-ws* tests; this file guards the bundle.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'
import { GROW_PATH_ZONE_LABELS } from '../lib/farmerVocabulary.js'
import { MODULE_EMPTY_SHELLS } from '../lib/moduleEmptyShell.js'
import { guardianProposalAriaLabel } from '../lib/farmerA11y.js'

const uiSrc = join(process.cwd(), 'src')
const repoDocs = join(process.cwd(), '..', 'docs')
const repoRoot = join(process.cwd(), '..')

describe('Phase 45 WS7 / OC-45 — farmer polish closure', () => {
  it('README documents Farmer-ready v1 criteria and sit-in gate', () => {
    const readme = readFileSync(join(repoRoot, 'README.md'), 'utf8')
    expect(readme).toContain('Farmer-ready v1')
    expect(readme).toContain('sit-in gate')
    expect(readme).toContain('phase_45_farmer_validation_whole_app_polish.plan.md')
    expect(readme).toMatch(/Phase 45.*WS1\/3\/5\/6\/7 shipped/i)
  })

  it('documents operator-tour §9 as polish shipped with farmer-ready criteria', () => {
    const tour = readFileSync(join(repoDocs, 'operator-tour.md'), 'utf8')
    expect(tour).toContain('## 9. Farmer validation sit-in (Phase 45')
    expect(tour).toContain('Farmer-ready v1 criteria')
    expect(tour).toContain('farmer-sit-in-protocol.md')
    expect(tour).toContain('sit-in-45-session-log-template.md')
    expect(tour).toContain('phase-45-closure.test.js')
    expect(tour).toContain('§10a')
    expect(tour).toContain('§10b')
    expect(tour).toContain('§10c')
    expect(tour).not.toContain('## 9. Farmer validation sit-in (Phase 45 — planned)')
    expect(tour).not.toContain('## 9. Farmer validation sit-in (Phase 45 — WS1 shipped)')
  })

  it('documents architecture §7.0k as polish shipped (not WS1-only stub)', () => {
    const arch = readFileSync(join(repoDocs, 'farm-guardian-architecture.md'), 'utf8')
    expect(arch).toContain('### 7.0k Farmer sit-in & PR validation (Phase 45')
    expect(arch).toContain('**Shipped')
    expect(arch).toContain('Vocabulary v2')
    expect(arch).toContain('phase-45-closure.test.js')
    expect(arch).not.toContain('### 7.0k Farmer sit-in & PR validation (Phase 45 — WS1 shipped)')
  })

  it('OC-45 marked completed in operational closure plan', () => {
    const closure = readFileSync(
      join(repoDocs, 'plans/phase_35_37_operational_closure.plan.md'),
      'utf8',
    )
    expect(closure).toContain('oc-45-closure')
    expect(closure).toMatch(/oc-45-closure[\s\S]*status: completed/)
    expect(closure).toContain('## Phase 45 — Farmer validation & whole-app polish')
    expect(closure).toContain('OC-45 docs/tests')
  })

  it('phase 45 plan marks WS7 completed', () => {
    const plan = readFileSync(
      join(repoDocs, 'plans/phase_45_farmer_validation_whole_app_polish.plan.md'),
      'utf8',
    )
    expect(plan).toContain('id: ws7-docs-tests')
    expect(plan).toMatch(/ws7-docs-tests[\s\S]*status: completed/)
    expect(plan).toContain('OC-45')
  })

  it('polish libs expose zone labels, module shells, and Guardian a11y labels', () => {
    expect(GROW_PATH_ZONE_LABELS.navMyZones).toBe('My zones')
    expect(MODULE_EMPTY_SHELLS.animals).toBeTruthy()
    expect(MODULE_EMPTY_SHELLS.aquaponics).toBeTruthy()
    expect(guardianProposalAriaLabel('confirm', 'Ack alert')).toContain('Confirm proposed action')
    expect(guardianProposalAriaLabel('dismiss', 'Setup pack')).toContain('without changing farm data')
  })

  it('closure Vitest bundle files exist', () => {
    for (const f of [
      '__tests__/phase-45-ws1-protocol.test.js',
      '__tests__/phase-45-ws3-closure.test.js',
      '__tests__/phase-45-ws5-module-shells.test.js',
      '__tests__/phase-45-ws6-a11y.test.js',
      '__tests__/phase-45-ws4-mobile.test.js',
      '__tests__/phase-45-closure.test.js',
      '__tests__/farmer-vocabulary-grow-path.test.js',
      '__tests__/module-empty-shell.test.js',
      '__tests__/farmer-a11y.test.js',
      'lib/farmerVocabulary.js',
      'lib/moduleEmptyShell.js',
      'lib/farmerA11y.js',
      'components/ModuleEmptyShell.vue',
    ]) {
      expect(existsSync(join(uiSrc, f))).toBe(true)
    }
  })
})
