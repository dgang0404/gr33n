/**
 * Phase 51 WS6 / OC-51 — Pi config platform sync closure bundle.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')
const repoDocs = join(repoRoot, 'docs')

describe('Phase 51 WS6 / OC-51 — Pi config sync closure', () => {
  it('API routes expose by-uid config sync endpoints', () => {
    const routes = readFileSync(join(repoRoot, 'cmd/api/routes.go'), 'utf8')
    expect(routes).toContain('GET /devices/by-uid/{device_uid}/config/version')
    expect(routes).toContain('GET /devices/by-uid/{device_uid}/config')
  })

  it('Go runtime config builder and migration exist', () => {
    const jsonCfg = readFileSync(join(repoRoot, 'internal/hardware/piconfig_json.go'), 'utf8')
    expect(jsonCfg).toContain('BuildPiRuntimeConfig')
    expect(jsonCfg).toContain('PiRuntimeSensor')
    const mig = readFileSync(
      join(repoRoot, 'db/migrations/20260608_phase51_device_config_version.sql'),
      'utf8',
    )
    expect(mig).toContain('config_version')
    expect(mig).toContain('bump_device_config_version_from_entity')
    expect(readFileSync(join(repoRoot, 'internal/handler/device/config_sync.go'), 'utf8')).toContain(
      'GetConfigByUID',
    )
  })

  it('Pi client has bootstrap, fetch, reload, and import script', () => {
    const client = readFileSync(join(repoRoot, 'pi_client/gr33n_client.py'), 'utf8')
    expect(client).toContain('load_bootstrap')
    expect(client).toContain('fetch_remote_config')
    expect(client).toContain('resolve_startup_config')
    expect(client).toContain('_reload_config')
    expect(client).toContain('_poll_config_version')
    expect(readFileSync(join(repoRoot, 'pi_client/import_config_to_platform.py'), 'utf8')).toContain(
      'import_wiring',
    )
    expect(readFileSync(join(repoRoot, 'pi_client/config.bootstrap.example.yaml'), 'utf8')).toContain(
      'device:',
    )
  })

  it('device card shows config sync staleness badge', () => {
    const lib = readFileSync(join(process.cwd(), 'src/lib/deviceConfigSync.js'), 'utf8')
    expect(lib).toContain('configSyncBadge')
    expect(lib).toContain('Never fetched')
    const card = readFileSync(join(process.cwd(), 'src/components/ActuatorCard.vue'), 'utf8')
    expect(card).toContain('configSyncBadge')
    expect(card).toContain('device-config-sync-badge')
  })

  it('status PATCH accepts last_config_fetch_at', () => {
    const handler = readFileSync(join(repoRoot, 'internal/handler/device/handler.go'), 'utf8')
    expect(handler).toContain('last_config_fetch_at')
    const sql = readFileSync(join(repoRoot, 'internal/db/devices.sql.go'), 'utf8')
    expect(sql).toContain('last_config_fetch_at')
  })

  it('smoke_phase51_test covers config and version bump', () => {
    const smoke = readFileSync(join(repoRoot, 'cmd/api/smoke_phase51_test.go'), 'utf8')
    expect(smoke).toContain('TestPhase51_ConfigByUIDAndVersion')
    expect(smoke).toContain('TestPhase51_ConfigVersionBumpsOnWiringPatch')
    expect(smoke).toContain('TestPhase51_StatusPatchStoresLastConfigFetchAt')
  })

  it('pi-integration-guide documents platform sync §2 and legacy §2b', () => {
    const guide = readFileSync(join(repoDocs, 'pi-integration-guide.md'), 'utf8')
    expect(guide).toContain('## 2. Platform sync (Phase 51')
    expect(guide).toContain('## 2b. Legacy full local YAML (opt-out)')
    expect(guide).toContain('import_config_to_platform.py')
    expect(guide).toContain('/devices/by-uid/{uid}/config')
  })

  it('architecture §7.0p documents Phase 51 sync loop', () => {
    const arch = readFileSync(join(repoDocs, 'farm-guardian-architecture.md'), 'utf8')
    expect(arch).toContain('### 7.0p Pi config platform sync (Phase 51 — shipped)')
    expect(arch).toContain('phase-51-closure.test.js')
    expect(arch).toContain('config_version')
  })

  it('operator-tour mentions platform sync badge', () => {
    const tour = readFileSync(join(repoDocs, 'operator-tour.md'), 'utf8')
    expect(tour).toContain('Config synced')
  })

  it('phase 51 plan marks all workstreams completed and shipped', () => {
    const plan = readFileSync(join(repoDocs, 'plans/phase_51_pi_config_sync.plan.md'), 'utf8')
    for (const id of [
      'ws1-api-config-endpoint',
      'ws2-pi-bootstrap-rewrite',
      'ws3-live-reload',
      'ws4-offline-safety',
      'ws5-backward-compat',
      'ws6-docs-tests',
    ]) {
      expect(plan).toMatch(new RegExp(`${id}[\\s\\S]*status: completed`))
    }
    expect(plan).toContain('**Shipped.**')
  })

  it('OC-51 marked completed in operational closure plan', () => {
    const closure = readFileSync(
      join(repoDocs, 'plans/phase_35_37_operational_closure.plan.md'),
      'utf8',
    )
    expect(closure).toContain('oc-51-closure')
    expect(closure).toMatch(/oc-51-closure[\s\S]*status: completed/)
    expect(closure).toContain('## Phase 51 — Pi config platform sync')
  })
})
