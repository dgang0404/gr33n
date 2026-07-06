/**
 * Phase 142 — Virtual Pi field validation closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { computeVirtualPiValidation } from '../lib/virtualPiValidation.js'

const repoDocs = join(process.cwd(), '..', 'docs')
const repoRoot = join(process.cwd(), '..')

describe('Phase 142 — Virtual Pi field validation closure', () => {
  it('plan is shipped', () => {
    const plan = readFileSync(join(repoDocs, 'plans/phase_142_virtual_pi_field_validation.plan.md'), 'utf8')
    expect(plan).toContain('**Shipped.**')
  })

  it('computeVirtualPiValidation ready_dry_run when wired + config', () => {
    const v = computeVirtualPiValidation({
      device: { id: 1, config: {} },
      sensors: [{ id: 1, device_id: 1, wiring: { gpio_pin: 4, source: 'dht22' } }],
      actuators: [],
      expectedConfigSha: 'abc123',
    })
    expect(v.status).toBe('ready_dry_run')
    expect(v.checklist.find((c) => c.id === 'wiring')?.ok).toBe(true)
    expect(v.checklist.find((c) => c.id === 'config')?.ok).toBe(true)
  })

  it('computeVirtualPiValidation needs_wiring when empty', () => {
    const v = computeVirtualPiValidation({
      device: { id: 1 },
      sensors: [],
      actuators: [],
      expectedConfigSha: 'abc',
    })
    expect(v.status).toBe('needs_wiring')
  })

  it('VirtualPi and Settings wire validation UI', () => {
    const vpi = readFileSync(join(process.cwd(), 'src/views/VirtualPi.vue'), 'utf8')
    const settings = readFileSync(join(process.cwd(), 'src/views/Settings.vue'), 'utf8')
    const edge = readFileSync(join(process.cwd(), 'src/components/SettingsEdgeValidationCard.vue'), 'utf8')
    expect(vpi).toContain('virtual-pi-validation-banner')
    expect(vpi).toContain('computeVirtualPiValidation')
    expect(settings).toContain('SettingsEdgeValidationCard')
    expect(edge).toContain('data-test="settings-edge-validation"')
  })

  it('Go smoke covers pi-config export', () => {
    const smoke = readFileSync(join(repoRoot, 'cmd/api/smoke_phase142_test.go'), 'utf8')
    expect(smoke).toContain('TestPhase142_VirtualPiConfigExport')
    expect(smoke).toContain('demo-veg-relay-01')
  })
})
