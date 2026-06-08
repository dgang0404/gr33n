/**
 * Phase 57 WS5 / OC-57 — per-device Pi API keys closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')
const repoDocs = join(repoRoot, 'docs')

describe('Phase 57 WS5 / OC-57 — device API keys closure', () => {
  it('documents migration, architecture, and plan shipped status', () => {
    const arch = readFileSync(join(repoDocs, 'farm-guardian-architecture.md'), 'utf8')
    const plan = readFileSync(join(repoDocs, 'plans/phase_57_pi_device_api_keys.plan.md'), 'utf8')
    const tour = readFileSync(join(repoDocs, 'operator-tour.md'), 'utf8')
    const piGuide = readFileSync(join(repoDocs, 'pi-sequent-hat-setup.md'), 'utf8')
    expect(existsSync(join(repoRoot, 'db/migrations/20260609_phase57_device_api_keys.sql'))).toBe(true)
    expect(arch).toContain('### 7.0u Per-device Pi API keys (Phase 57 — shipped)')
    expect(arch).toContain('device_api_keys')
    expect(arch).toContain('X-Device-Key')
    expect(plan).toMatch(/ws5-docs-tests[\s\S]*status: completed/)
    expect(plan).toContain('**Shipped.**')
    expect(tour).toContain('### 6l. Per-device Pi API keys (Phase 57 — shipped)')
    expect(piGuide).toContain('GR33N_DEVICE_API_KEY')
    expect(piGuide).toContain('/etc/gr33n/device.key')
  })

  it('OC-57 row is closed in operational closure doc', () => {
    const oc = readFileSync(
      join(repoDocs, 'plans/phase_35_37_operational_closure.plan.md'),
      'utf8',
    )
    expect(oc).toContain('## Phase 57 — Per-device Pi API keys')
    expect(oc).toMatch(/oc-57-closure[\s\S]*status: completed/)
    expect(oc).toContain('phase-57-closure.test.js')
  })

  it('API routes and edge auth middleware accept device keys', () => {
    const routes = readFileSync(join(repoRoot, 'cmd/api/routes.go'), 'utf8')
    expect(routes).toContain('GET /devices/{id}/api-keys')
    expect(routes).toContain('POST /devices/{id}/api-keys')
    expect(routes).toContain('POST /devices/{id}/api-keys/{key_id}/revoke')
    const auth = readFileSync(join(repoRoot, 'cmd/api/pi_edge_auth.go'), 'utf8')
    expect(auth).toContain('X-Device-Key')
    expect(auth).toContain('WithDeviceKeyAuth')
  })

  it('Pi client resolves device key header before legacy shared key', () => {
    const client = readFileSync(join(repoRoot, 'pi_client/gr33n_client.py'), 'utf8')
    expect(client).toContain('resolve_edge_api_credential')
    expect(client).toContain('GR33N_DEVICE_API_KEY')
    expect(client).toContain('X-Device-Key')
  })

  it('UI surfaces issue, show-once, revoke, and legacy badge', () => {
    const panel = readFileSync(join(process.cwd(), 'src/components/DeviceApiKeyPanel.vue'), 'utf8')
    const wizard = readFileSync(join(process.cwd(), 'src/views/DeviceSetupWizard.vue'), 'utf8')
    const card = readFileSync(join(process.cwd(), 'src/components/ActuatorCard.vue'), 'utf8')
    expect(panel).toContain('device-key-show-once')
    expect(panel).toContain('device-legacy-auth-badge')
    expect(wizard).toContain('DeviceApiKeyPanel')
    expect(card).toContain('device-key-toggle')
  })

  it('smoke_phase57_test covers issue, auth, and revoke', () => {
    const smoke = readFileSync(join(repoRoot, 'cmd/api/smoke_phase57_test.go'), 'utf8')
    expect(smoke).toContain('TestPhase57_DeviceAPIKeyIssueAuthRevoke')
    expect(smoke).toContain('X-Device-Key')
    expect(smoke).toContain('StatusForbidden')
  })
})
