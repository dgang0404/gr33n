/**
 * Phase 45 WS1 — farmer sit-in protocol closure (docs + PR path UI anchors).
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'
import { buildSetupStarters } from '../lib/guardianStarters.js'
import { impactForProposal } from '../lib/guardianImpact.js'

const repoDocs = join(process.cwd(), '..', 'docs')
const uiSrc = join(process.cwd(), 'src')

describe('Phase 45 WS1 — sit-in protocol closure', () => {
  it('farmer-sit-in-protocol is shipped with prep, three paths, and session log link', () => {
    const protocol = readFileSync(join(repoDocs, 'workstreams/farmer-sit-in-protocol.md'), 'utf8')
    expect(protocol).toContain('status: shipped')
    expect(protocol).toContain('## 0. Facilitator prep')
    expect(protocol).toContain('### 4.1 Path 1 — `ack_alert`')
    expect(protocol).toContain('### 4.2 Path 2 — `apply_grow_setup_pack`')
    expect(protocol).toContain('### 4.3 Path 3 — Dismiss')
    expect(protocol).toContain('### 4.4 UI anchors')
    expect(protocol).toContain('sit-in-45-session-log-template.md')
    expect(protocol).toContain('sit-in-45')
  })

  it('session log template exists with PR scorecard and closure checklist', () => {
    const log = readFileSync(join(repoDocs, 'workstreams/sit-in-45-session-log-template.md'), 'utf8')
    expect(log).toContain('status: shipped')
    expect(log).toContain('ack_alert')
    expect(log).toContain('apply_grow_setup_pack')
    expect(log).toContain('dismiss')
    expect(log).toContain('sit-in-46-backlog')
    expect(log).toContain('≥2 sessions A')
  })

  it('setup starters match protocol ack and setup-pack phrases', () => {
    const ackStarters = buildSetupStarters({
      surface: 'first_run_dashboard',
      farmId: 1,
      zoneCount: 1,
      zones: [{ id: 2, name: 'Flower Room' }],
      unreadAlerts: [{ id: 42, subject: 'Humidity high', is_read: false }],
    })
    const ack = ackStarters.find((s) => s.id === 'handle-alert')
    expect(ack).toBeTruthy()
    expect(ack.message).toContain('Acknowledge alert #42')

    const growStarters = buildSetupStarters({
      surface: 'empty_zone_grow',
      farmId: 1,
      zoneCount: 1,
      zones: [{ id: 5, name: 'Veg Room' }],
      zoneName: 'Veg Room',
      activeCycles: [],
    })
    const grow = growStarters.find((s) => s.id === 'start-grow')
    expect(grow.message).toBe(
      'Add my philodendron to Veg Room with a light fertigation program',
    )
  })

  it('ack impact line is readable before Confirm', () => {
    const { lines } = impactForProposal({
      tool: 'ack_alert',
      args: { alert_id: 7 },
    })
    expect(lines.some((l) => /acknowledge the alert/i.test(l))).toBe(true)
  })

  it('Dismiss is client-side only in GuardianActionProposal', () => {
    const vue = readFileSync(join(uiSrc, 'components/GuardianActionProposal.vue'), 'utf8')
    expect(vue).toContain('data-test="guardian-proposal-dismiss"')
    expect(vue).toContain('data-test="guardian-proposal-dismissed"')
    const dismissFn = vue.match(/async function onDismiss\(\) \{[\s\S]*?\n\}/)
    expect(dismissFn).toBeTruthy()
    expect(dismissFn[0]).toContain('api.post')
    expect(dismissFn[0]).toContain('/dismiss')
  })

  it('phase 45 parent plan marks WS1 complete', () => {
    const plan = readFileSync(
      join(repoDocs, 'plans/archive/phase_45_farmer_validation_whole_app_polish.plan.md'),
      'utf8',
    )
    expect(plan).toContain('id: ws1-sit-in-protocol')
    expect(plan).toMatch(/ws1-sit-in-protocol[\s\S]*status: completed/)
  })

  it('operator-tour §9 references shipped protocol kit', () => {
    const tour = readFileSync(join(repoDocs, 'operator-tour.md'), 'utf8')
    expect(tour).toContain('## 9. Farmer validation sit-in (Phase 45')
    expect(tour).toContain('farmer-sit-in-protocol.md')
    expect(tour).toContain('sit-in-45-session-log-template.md')
    expect(tour).not.toContain('## 9. Farmer validation sit-in (Phase 45 — planned)')
  })

  it('closure test file exists', () => {
    expect(existsSync(join(uiSrc, '__tests__/phase-45-ws1-protocol.test.js'))).toBe(true)
  })
})
