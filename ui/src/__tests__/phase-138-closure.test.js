/**
 * Phase 138 — Guardian inference policy closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { resolveEffectiveChatModelName } from '../lib/guardianModelGrounded.js'

const repoDocs = join(process.cwd(), '..', 'docs')
const repoRoot = join(process.cwd(), '..')

describe('Phase 138 — inference policy closure', () => {
  it('plan is shipped', () => {
    const plan = readFileSync(join(repoDocs, 'plans/archive/phase_138_guardian_inference_policy.plan.md'), 'utf8')
    expect(plan).toContain('**Shipped.**')
  })

  it('resolveEffectiveChatModelName picks counsel vs quick farm policy', () => {
    expect(
      resolveEffectiveChatModelName({
        farmCounselModel: 'phi3:mini',
        farmQuickModel: 'tinyllama',
        serverDefault: 'llama3.1:8b',
        grounded: true,
      }),
    ).toBe('phi3:mini')
    expect(
      resolveEffectiveChatModelName({
        farmCounselModel: 'phi3:mini',
        farmQuickModel: 'tinyllama',
        serverDefault: 'llama3.1:8b',
        grounded: false,
      }),
    ).toBe('tinyllama')
  })

  it('UI wires Settings model policy and chat cost hint', () => {
    const settings = readFileSync(join(process.cwd(), 'src/views/Settings.vue'), 'utf8')
    const policy = readFileSync(join(process.cwd(), 'src/components/GuardianSettingsModelPolicyCard.vue'), 'utf8')
    const chat = readFileSync(join(process.cwd(), 'src/stores/guardianChat.js'), 'utf8')
    expect(settings).toContain('GuardianSettingsModelPolicyCard')
    expect(policy).toContain('settings-counsel-model')
    expect(policy).toContain('settings-quick-model')
    expect(chat).toContain('body.grounded')
  })

  it('Go exposes split-host health and model policy columns', () => {
    const offline = readFileSync(join(repoRoot, 'internal/farmguardian/offline.go'), 'utf8')
    const policy = readFileSync(join(repoRoot, 'internal/farmguardian/inference_policy.go'), 'utf8')
    const settings = readFileSync(join(repoRoot, 'internal/handler/farm/guardian_settings.go'), 'utf8')
    expect(offline).toContain('embedding_reachable')
    expect(offline).toContain('split_inference_hosts')
    expect(policy).toContain('FarmCounselModel')
    expect(policy).toContain('FarmQuickModel')
    expect(settings).toContain('guardian_counsel_model')
    expect(settings).toContain('guardian_quick_model')
  })
})
