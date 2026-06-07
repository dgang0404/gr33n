/**
 * Phase 50 WS3 — hardware wiring visibility closure guards.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')

describe('Phase 50 — hardware wiring visibility', () => {
  it('hardware wiring lib exists', () => {
    const lib = readFileSync(join(process.cwd(), 'src/lib/hardwareWiring.js'), 'utf8')
    expect(lib).toContain('formatWiringLabel')
    expect(lib).toContain('resolveWiring')
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
  })

  it('Sensor detail includes wiring panel with PATCH', () => {
    const panel = readFileSync(join(process.cwd(), 'src/components/HardwareWiringPanel.vue'), 'utf8')
    expect(panel).toContain('/wiring')
    const detail = readFileSync(join(process.cwd(), 'src/views/SensorDetail.vue'), 'utf8')
    expect(detail).toContain('HardwareWiringPanel')
  })

  it('Go wiring package validates sensor sources', () => {
    const wiring = readFileSync(join(repoRoot, 'internal/hardware/wiring.go'), 'utf8')
    expect(wiring).toContain('ValidateSensorWiring')
    expect(wiring).toContain('MergeWiring')
  })

  it('API routes expose PATCH wiring endpoints', () => {
    const routes = readFileSync(join(repoRoot, 'cmd/api/routes.go'), 'utf8')
    expect(routes).toContain('PATCH /sensors/{id}/wiring')
    expect(routes).toContain('PATCH /actuators/{id}/wiring')
  })

  it('WS4 pi-config generator endpoint and wizard download', () => {
    const routes = readFileSync(join(repoRoot, 'cmd/api/routes.go'), 'utf8')
    expect(routes).toContain('GET /devices/{id}/pi-config')
    const gen = readFileSync(join(repoRoot, 'internal/hardware/piconfig.go'), 'utf8')
    expect(gen).toContain('GeneratePiConfigYAML')
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

  it('wiring panel previews conflicts before save', () => {
    const panel = readFileSync(join(process.cwd(), 'src/components/HardwareWiringPanel.vue'), 'utf8')
    expect(panel).toContain('conflictPreview')
    expect(panel).toContain('findWiringConflict')
  })

  it('demo backfill migration exists', () => {
    const mig = readFileSync(
      join(repoRoot, 'db/migrations/20260607_phase50_hardware_wiring_backfill.sql'),
      'utf8',
    )
    expect(mig).toContain('demo-veg-relay-01')
    expect(mig).toContain("config->'wiring'")
  })
})
