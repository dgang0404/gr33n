/**
 * Phase 50 WS6 / OC-50 — hardware wiring visibility closure bundle.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')
const repoDocs = join(repoRoot, 'docs')

describe('Phase 50 WS6 / OC-50 — hardware wiring closure', () => {
  it('hardware wiring lib exists', () => {
    const lib = readFileSync(join(process.cwd(), 'src/lib/hardwareWiring.js'), 'utf8')
    expect(lib).toContain('formatWiringLabel')
    expect(lib).toContain('resolveWiring')
    expect(lib).toContain('findWiringConflict')
  })

  it('Sensors list shows wiring badge column', () => {
    const view = readFileSync(join(process.cwd(), 'src/views/Sensors.vue'), 'utf8')
    expect(view).toContain('HardwareWiringBadge')
    expect(view).toContain('Wiring')
  })

  it('Controls cards show wiring badge', () => {
    const view = readFileSync(join(process.cwd(), 'src/views/Actuators.vue'), 'utf8')
    const badge = readFileSync(join(process.cwd(), 'src/components/HardwareWiringBadge.vue'), 'utf8')
    expect(view).toContain('HardwareWiringBadge')
    expect(badge).toContain('Not wired yet')
    expect(badge).toContain("v-nav-hint=\"'/pi-setup'\"")
  })

  it('Sensor detail includes wiring panel with PATCH and conflict preview', () => {
    const panel = readFileSync(join(process.cwd(), 'src/components/HardwareWiringPanel.vue'), 'utf8')
    expect(panel).toContain('/wiring')
    expect(panel).toContain('conflictPreview')
    expect(panel).toContain('findWiringConflict')
    const detail = readFileSync(join(process.cwd(), 'src/views/SensorDetail.vue'), 'utf8')
    expect(detail).toContain('HardwareWiringPanel')
  })

  it('Go wiring package validates and generates pi config', () => {
    const wiring = readFileSync(join(repoRoot, 'internal/hardware/wiring.go'), 'utf8')
    expect(wiring).toContain('ValidateSensorWiring')
    expect(wiring).toContain('MergeWiring')
    const piconfig = readFileSync(join(repoRoot, 'internal/hardware/piconfig.go'), 'utf8')
    expect(piconfig).toContain('GeneratePiConfigYAML')
    expect(readFileSync(join(repoRoot, 'internal/hardware/piconfig_test.go'), 'utf8')).toContain('roundTrip')
    expect(readFileSync(join(repoRoot, 'internal/hardware/conflict_test.go'), 'utf8')).toContain('FindWiringConflict')
    expect(readFileSync(join(repoRoot, 'internal/handler/sensor/wiring_test.go'), 'utf8')).toContain('PatchWiring')
  })

  it('API routes expose wiring and pi-config endpoints', () => {
    const routes = readFileSync(join(repoRoot, 'cmd/api/routes.go'), 'utf8')
    expect(routes).toContain('PATCH /sensors/{id}/wiring')
    expect(routes).toContain('PATCH /actuators/{id}/wiring')
    expect(routes).toContain('GET /devices/{id}/pi-config')
  })

  it('WS4 device wizard downloads generated config', () => {
    const wizard = readFileSync(join(process.cwd(), 'src/views/DeviceSetupWizard.vue'), 'utf8')
    expect(wizard).toContain('pi-config')
    expect(wizard).toContain('Download config.yaml')
  })

  it('WS5 sanity report flags wiring conflicts', () => {
    const sql = readFileSync(join(repoRoot, 'scripts/sql/db_sanity_report.sql'), 'utf8')
    const sh = readFileSync(join(repoRoot, 'scripts/db-sanity-report.sh'), 'utf8')
    expect(sql).toContain('GPIO pin conflicts')
    expect(sql).toContain('derived sensors with missing input')
    expect(sh).toContain('gpio_conflicts')
    expect(sh).toContain('i2c_conflicts')
  })

  it('demo backfill migration exists', () => {
    const mig = readFileSync(
      join(repoRoot, 'db/migrations/20260607_phase50_hardware_wiring_backfill.sql'),
      'utf8',
    )
    expect(mig).toContain('demo-veg-relay-01')
    expect(mig).toContain("config->'wiring'")
  })

  it('pi-integration-guide documents DB-first wiring (§2a) and Phase 51 sync', () => {
    const guide = readFileSync(join(repoDocs, 'pi-integration-guide.md'), 'utf8')
    expect(guide).toContain('DB-first wiring')
    expect(guide).toContain('/devices/{id}/pi-config')
    expect(guide).toContain('/devices/by-uid/{uid}/config')
    expect(guide).toContain('manual fallback')
  })

  it('high-impact nav-hint affordances are wired', () => {
    const actuators = readFileSync(join(process.cwd(), 'src/views/Actuators.vue'), 'utf8')
    const card = readFileSync(join(process.cwd(), 'src/components/ActuatorCard.vue'), 'utf8')
    const operator = readFileSync(join(process.cwd(), 'src/views/OperatorGuide.vue'), 'utf8')
    const piGuide = readFileSync(join(process.cwd(), 'src/views/PiSetupGuide.vue'), 'utf8')
    const hints = readFileSync(join(process.cwd(), 'src/lib/emptyStateHint.js'), 'utf8')
    expect(actuators).toContain("v-nav-hint=\"'/pi-setup'\"")
    expect(card).toContain('syncBadgeNavHint')
    expect(operator).toContain("v-nav-hint=\"'/pi-setup'\"")
    expect(piGuide).toContain("v-nav-hint=\"'/actuators'\"")
    expect(hints).toContain("actionTo: '/comfort-targets'")
  })

  it('pi-sequent-hat-setup guide and /pi-setup redirect to hardware workspace exist', () => {
    const doc = readFileSync(join(repoDocs, 'pi-sequent-hat-setup.md'), 'utf8')
    expect(doc).toContain('Sequent Microsystems')
    expect(doc).toContain('DIP switch')
    const router = readFileSync(join(process.cwd(), 'src/router/index.js'), 'utf8')
    const nav = readFileSync(join(process.cwd(), 'src/lib/navGroups.js'), 'utf8')
    const workspaces = readFileSync(join(process.cwd(), 'src/lib/workspaces.js'), 'utf8')
    expect(router).toContain('buildLegacyRedirectRoutes')
    expect(workspaces).toContain("'/pi-setup': { tab: 'reference' }")
    expect(nav).toContain("to: '/hardware'")
  })

  it('architecture §7.0o documents Phase 50 wiring shipped', () => {
    const arch = readFileSync(join(repoDocs, 'farm-guardian-architecture.md'), 'utf8')
    expect(arch).toContain('### 7.0o Hardware wiring visibility (Phase 50')
    expect(arch).toContain('phase-50-closure.test.js')
    expect(arch).toContain('piconfig.go')
  })

  it('operator-tour mentions wiring badges and generated Pi config', () => {
    const tour = readFileSync(join(repoDocs, 'operator-tour.md'), 'utf8')
    expect(tour).toContain('wiring badges')
    expect(tour).toContain('generated config.yaml')
  })

  it('phase 50 plan marks all workstreams completed', () => {
    const plan = readFileSync(
      join(repoDocs, 'plans/phase_50_hardware_wiring_visibility.plan.md'),
      'utf8',
    )
    for (const id of [
      'ws1-wiring-model',
      'ws2-api-read-write',
      'ws3-ui-surface',
      'ws4-config-generator',
      'ws5-validation-conflicts',
      'ws6-docs-tests',
    ]) {
      expect(plan).toMatch(new RegExp(`${id}[\\s\\S]*status: completed`))
    }
    expect(plan).toContain('**Shipped.**')
  })

  it('OC-50 marked completed in operational closure plan', () => {
    const closure = readFileSync(
      join(repoDocs, 'plans/phase_35_37_operational_closure.plan.md'),
      'utf8',
    )
    expect(closure).toContain('oc-50-closure')
    expect(closure).toMatch(/oc-50-closure[\s\S]*status: completed/)
    expect(closure).toContain('## Phase 50 — Hardware wiring visibility')
  })
})
