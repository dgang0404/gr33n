/**
 * Phase 44 WS6 / OC-44 — getting started + edge wizard docs closure (Vitest bundle guard).
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'
import {
  farmSetupRoute,
  farmBootstrapApplyPath,
  FARM_SETUP_PRIMARY_CHOICES,
} from '../lib/farmSetupWizard.js'
import { zoneSetupRoute } from '../lib/zoneSetupWizard.js'
import { deviceSetupRoute, PI_FIELD_CHECKLIST } from '../lib/deviceSetupWizard.js'
import {
  computeFirstRunChecklist,
  shouldShowFirstRunChecklist,
} from '../lib/firstRunChecklist.js'
import { buildSetupStarters } from '../lib/guardianStarters.js'
import { routeContextRefFromRoute } from '../lib/guardianRouteRef.js'

const uiSrc = join(process.cwd(), 'src')
const repoDocs = join(process.cwd(), '..', 'docs')
const repoRoot = join(process.cwd(), '..')

describe('Phase 44 WS6 / OC-44 — setup wizard closure', () => {
  it('documents operator-tour §8 and §6g as shipped', () => {
    const tour = readFileSync(join(repoDocs, 'operator-tour.md'), 'utf8')
    expect(tour).toContain('## 8. Getting started & edge install (Phase 44')
    expect(tour).toContain('**Shipped.**')
    expect(tour).toContain('### 6g. Guardian during setup (Phase 44')
    expect(tour).toContain('/farms/')
    expect(tour).toContain('Guardian setup chips')
    expect(tour).not.toContain('## 8. Getting started & edge install (Phase 44 — planned)')
    expect(tour).not.toContain('### 6g. Guardian during setup (Phase 44 — planned)')
  })

  it('documents architecture §7.0j as shipped (not a stub)', () => {
    const arch = readFileSync(join(repoDocs, 'farm-guardian-architecture.md'), 'utf8')
    expect(arch).toContain('### 7.0j Getting started & edge wizards (Phase 44)')
    expect(arch).toContain('**Shipped')
    expect(arch).toContain('setup_mode.go')
    expect(arch).not.toContain('### 7.0j Getting started & edge wizards (Phase 44 — planned)')
    expect(arch).not.toContain('Phase 44 — planned)')
  })

  it('wizard routes and bootstrap apply path match UI helpers', () => {
    expect(farmSetupRoute(12)).toBe('/farms/12/setup')
    expect(zoneSetupRoute(12)).toBe('/farms/12/zones/new')
    expect(deviceSetupRoute(12, 3)).toBe('/farms/12/devices/new?zone_id=3')
    expect(farmBootstrapApplyPath(12)).toBe('/farms/12/bootstrap-template')
    const router = readFileSync(join(uiSrc, 'router/index.js'), 'utf8')
    expect(router).toContain("name: 'farm-setup'")
    expect(router).toContain("name: 'zone-setup'")
    expect(router).toContain("name: 'device-setup'")
  })

  it('first-run checklist links to wizards and comfort targets', () => {
    const items = computeFirstRunChecklist({
      farmId: 4,
      zones: [],
      devices: [],
      setpoints: [],
      schedules: [],
    })
    expect(items[0].to).toBe('/farms/4/zones/new')
    expect(items[1].to).toBe('/farms/4/devices/new')
    expect(items[2].to).toBe('/comfort-targets')
    expect(items[3].to).toBe('/comfort-targets?tab=schedules')
    expect(shouldShowFirstRunChecklist(4, items)).toBe(true)
  })

  it('Guardian setup starters and route refs exist for wizard paths', () => {
    const starters = buildSetupStarters({
      surface: 'first_run_dashboard',
      farmId: 8,
      zoneCount: 0,
    })
    expect(starters.length).toBeLessThanOrEqual(3)
    const ref = routeContextRefFromRoute({
      path: '/farms/8/devices/new',
      meta: {},
    })
    expect(ref.name).toBe('Connect edge device')
  })

  it('device wizard embeds Pi field checklist items', () => {
    expect(PI_FIELD_CHECKLIST.length).toBeGreaterThanOrEqual(5)
    expect(PI_FIELD_CHECKLIST.some((i) => /pi_client/i.test(i.label))).toBe(true)
  })

  it('farm setup exposes template cards for wizard choose step', () => {
    expect(FARM_SETUP_PRIMARY_CHOICES.length).toBeGreaterThanOrEqual(3)
  })

  it('closure Vitest and Go smoke files exist', () => {
    for (const f of [
      '__tests__/farm-setup-wizard.test.js',
      '__tests__/zone-setup-wizard.test.js',
      '__tests__/device-setup-wizard.test.js',
      '__tests__/first-run-checklist.test.js',
      '__tests__/guardian-setup-starters.test.js',
      '__tests__/phase-44-wizard-navigation.test.js',
      'views/FarmSetupWizard.vue',
      'views/ZoneSetupWizard.vue',
      'views/DeviceSetupWizard.vue',
      'lib/firstRunChecklist.js',
    ]) {
      expect(existsSync(join(uiSrc, f))).toBe(true)
    }
    expect(existsSync(join(repoRoot, 'internal/farmguardian/setup_mode.go'))).toBe(true)
    expect(existsSync(join(uiSrc, '__tests__/phase-44-guardian-closure.test.js'))).toBe(true)
    const smoke = readFileSync(join(repoRoot, 'cmd/api/smoke_farms_test.go'), 'utf8')
    expect(smoke).toContain('TestPhase44WizardBootstrapApply')
  })
})
