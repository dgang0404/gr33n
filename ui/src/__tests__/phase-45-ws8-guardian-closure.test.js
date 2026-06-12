/**
 * Phase 45 WS8 — Guardian PR sit-in paths closure (ack · setup pack · dismiss).
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'
import { buildSetupStarters } from '../lib/guardianStarters.js'
import { impactForProposal } from '../lib/guardianImpact.js'

const repoDocs = join(process.cwd(), '..', 'docs')
const uiSrc = join(process.cwd(), 'src')

describe('Phase 45 WS8 — Guardian PR path closure', () => {
  it('dry-run log documents pass on all three paths for two personas', () => {
    const log = readFileSync(
      join(repoDocs, 'workstreams/sit-in-45-dry-run-log.md'),
      'utf8',
    )
    expect(log).toContain('DR-A')
    expect(log).toContain('DR-B')
    expect(log).toContain('ack_alert')
    expect(log).toContain('apply_grow_setup_pack')
    expect(log).toContain('dismiss')
    expect(log).toContain('| pass | pass |')
  })

  it('ack_alert impact line is farmer-readable', () => {
    const { lines } = impactForProposal({ tool: 'ack_alert', args: { alert_id: 4 } })
    expect(lines.join(' ')).toMatch(/acknowledge/i)
  })

  it('setup pack starter phrase matches Phase 32 matcher (Session B path)', () => {
    const starters = buildSetupStarters({
      surface: 'empty_zone_grow',
      farmId: 1,
      zoneCount: 1,
      zones: [{ id: 5, name: 'Bench A' }],
      zoneName: 'Bench A',
      activeCycles: [],
    })
    const grow = starters.find((s) => s.id === 'start-grow')
    expect(grow.message).toContain('light fertigation program')
    expect(grow.message).toContain('Bench A')
  })

  it('GuardianActionProposal dismiss persists via API (Phase 73)', () => {
    const vue = readFileSync(join(uiSrc, 'components/GuardianActionProposal.vue'), 'utf8')
    expect(vue).toContain('data-test="guardian-proposal-dismiss"')
    expect(vue).toContain("uiStatus.value = 'dismissed'")
    const dismissFn = vue.match(/async function onDismiss\(\) \{[\s\S]*?\n\}/)
    expect(dismissFn[0]).toContain('api.post')
    expect(dismissFn[0]).toContain('/dismiss')
  })

  it('setup pack card shows high-tier warning before Confirm', () => {
    const proposal = readFileSync(join(uiSrc, 'components/GuardianActionProposal.vue'), 'utf8')
    const setupPack = readFileSync(join(uiSrc, 'components/SetupPackProposalCard.vue'), 'utf8')
    expect(proposal).toContain('SetupPackProposalCard')
    expect(proposal).toContain('data-test="guardian-proposal-high-warning"')
    expect(setupPack).toContain('data-test="setup-pack-proposal-card"')
  })

  it('guardian PR spec marks WS8 definition of done complete', () => {
    const spec = readFileSync(join(repoDocs, 'plans/phase_45_guardian_pr_spec.md'), 'utf8')
    expect(spec).toMatch(/status: completed/)
    expect(spec).toContain('- [x] ack + setup pack + dismiss **pass** documented')
  })

  it('dry-run script exists', () => {
    expect(existsSync(join(repoDocs, '../scripts/sit-in-dry-run.sh'))).toBe(true)
  })
})
