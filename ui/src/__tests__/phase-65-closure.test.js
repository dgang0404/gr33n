/**
 * Phase 65 WS4 / OC-65 — Guardian Pi diagnostics closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'

const repoDocs = join(process.cwd(), '..', 'docs')
const repoRoot = join(process.cwd(), '..')

describe('Phase 65 WS4 / OC-65 — Pi diagnostics closure', () => {
  it('documents summarize_device_health read tool and architecture section', () => {
    const arch = readFileSync(join(repoDocs, 'farm-guardian-architecture.md'), 'utf8')
    const plan = readFileSync(join(repoDocs, 'plans/archive/phase_65_guardian_pi_diagnostics.plan.md'), 'utf8')
    expect(existsSync(join(repoRoot, 'internal/farmguardian/readtools_device.go'))).toBe(true)
    expect(arch).toContain('summarize_device_health')
    expect(plan).toContain('**Shipped.**')
  })

  it('grounding no longer claims Guardian is blind to wiring', () => {
    const guardian = readFileSync(join(repoRoot, 'internal/rag/synthesis/guardian.go'), 'utf8')
    const field = readFileSync(join(repoRoot, 'internal/farmguardian/field_assistant.go'), 'utf8')
    expect(guardian).not.toContain('Guardian cannot see the wiring')
    expect(guardian).toContain('summarize_device_health')
    expect(field).not.toContain('cannot see their wiring')
    expect(field).toContain('summarize_device_health')
  })

  it('context_ref routes mention summarize_device_health', () => {
    const ctx = readFileSync(join(repoRoot, 'internal/farmguardian/context_ref.go'), 'utf8')
    expect(ctx).toContain('summarize_device_health')
    expect(ctx).toContain('/pi-setup')
    expect(ctx).toContain('/sensors')
    expect(ctx).toContain('/actuators')
  })

  it('read tool registry includes summarize_device_health', () => {
    const readtools = readFileSync(join(repoRoot, 'internal/farmguardian/readtools.go'), 'utf8')
    expect(readtools).toContain('"summarize_device_health"')
    expect(readtools).toContain('shouldRunSummarizeDeviceHealthReadIntent')
  })
})
