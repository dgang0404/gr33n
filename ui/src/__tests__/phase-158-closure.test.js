/**
 * Phase 158 — accessibility pass (core workspaces).
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'
import { citationLinkAriaLabel } from '../lib/guardianCitationLabels.js'

const uiSrc = join(process.cwd(), 'src')
const repoDocs = join(process.cwd(), '..', 'docs')

describe('Phase 158 — accessibility closure', () => {
  it('ships a11y audit doc', () => {
    expect(existsSync(join(repoDocs, 'a11y-audit-2026-07-11.md'))).toBe(true)
    const audit = readFileSync(join(repoDocs, 'a11y-audit-2026-07-11.md'), 'utf8')
    expect(audit).toContain('Phase 158')
    expect(audit).toContain('chat-accuracy-banner')
  })

  it('Guardian chat wires citation aria-label and accuracy alert', () => {
    const chat = readFileSync(join(uiSrc, 'components/GuardianChatPanel.vue'), 'utf8')
    expect(chat).toContain('citationLinkAriaLabel')
    expect(chat).toContain(':aria-label="citationLinkAriaLabel(c)"')
    expect(chat).toContain('data-test="chat-accuracy-banner"')
    expect(chat).toMatch(/data-test="chat-accuracy-banner"[\s\S]*role="alert"/)
    expect(chat).toContain('id="chat-message-input"')
    expect(chat).toContain('for="chat-message-input"')
    expect(chat).toContain('guardian-proposal-confirm')
  })

  it('citationLinkAriaLabel includes source and excerpt', () => {
    const label = citationLinkAriaLabel({
      source_type: 'field_guide',
      source_id: 42,
      excerpt: 'Check pH daily in veg.',
    })
    expect(label).toContain('Field guide')
    expect(label).toContain('#42')
    expect(label).toContain('Check pH daily')
  })

  it('Guardian drawer uses focus trap composable', () => {
    const drawer = readFileSync(join(uiSrc, 'components/GuardianDrawer.vue'), 'utf8')
    expect(drawer).toContain('useDialogFocusTrap')
    expect(drawer).toContain('role="dialog"')
    expect(existsSync(join(uiSrc, 'composables/useDialogFocusTrap.js'))).toBe(true)
  })

  it('workspace shell has skip link and aria-current nav', () => {
    const app = readFileSync(join(uiSrc, 'App.vue'), 'utf8')
    const side = readFileSync(join(uiSrc, 'components/SideNav.vue'), 'utf8')
    expect(app).toContain('Skip to main content')
    expect(app).toContain('id="main-content"')
    expect(app).toContain('aria-current')
    expect(side).toContain('aria-current')
  })

  it('zone detail tabs use tablist semantics', () => {
    const zone = readFileSync(join(uiSrc, 'views/ZoneDetail.vue'), 'utf8')
    expect(zone).toContain('role="tablist"')
    expect(zone).toContain('role="tab"')
    expect(zone).toContain(':aria-selected="activeTab === tab.id"')
  })

  it('proposal high warning and settings model feedback are announced', () => {
    const proposal = readFileSync(join(uiSrc, 'components/GuardianActionProposal.vue'), 'utf8')
    const settings = readFileSync(join(uiSrc, 'components/GuardianSettingsModelPolicyCard.vue'), 'utf8')
    expect(proposal).toMatch(/guardian-proposal-high-warning[\s\S]*role="alert"/)
    expect(settings).toContain('aria-live')
  })
})
