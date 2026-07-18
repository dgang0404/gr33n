/**
 * Canonical GuardianChatPanel.vue source wiring — merged from phase-closure
 * readFileSync blocks (Phase 202). Add new panel template anchors here, not in
 * phase-N-closure.test.js.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'

const uiSrc = join(process.cwd(), 'src')
const panel = readFileSync(join(uiSrc, 'components/GuardianChatPanel.vue'), 'utf8')

describe('GuardianChatPanel — accuracy + citations', () => {
  it('maps accuracy notes to farmer-facing banner and citation deep links', () => {
    expect(panel).toContain('chat-accuracy-banner')
    expect(panel).toContain('accuracyNoteMessage(t.accuracy_note)')
    expect(panel).toContain("accuracy_note: finalEvent.accuracy_note")
    expect(panel).toContain('setTranscript')
    expect(panel).toMatch(/data-test="chat-accuracy-banner"[\s\S]*role="alert"/)
    expect(panel).toContain('chat-citation-link')
    expect(panel).toContain('v-if="c.route"')
    expect(panel).toContain('v-nav-hint="c.route"')
    expect(panel).toContain('citationLinkAriaLabel')
    expect(panel).toContain(':aria-label="citationLinkAriaLabel(c)"')
  })
})

describe('GuardianChatPanel — a11y + composer', () => {
  it('wires labeled composer, proposal confirm, and turn debug', () => {
    expect(panel).toContain('id="chat-message-input"')
    expect(panel).toContain('for="chat-message-input"')
    expect(panel).toContain('guardian-proposal-confirm')
    expect(panel).toContain('GuardianTurnDebug')
    expect(panel).toContain('showTurnDebug')
    expect(panel).toContain('lastTurnDebug')
    expect(panel).toContain('finalEvent.debug')
  })

  it('full-page composer diet hides extras behind more menu', () => {
    expect(panel).toContain('data-test="chat-composer-more"')
    expect(panel).toContain('isFullPageDiet')
    expect(panel).toContain('showComposerExtras')
    expect(panel).toContain('composerExtrasOpen')
  })
})

describe('GuardianChatPanel — counsel, nudges, and starters', () => {
  it('wires farm counsel auto-send and morning walkthrough starters', () => {
    expect(panel).toContain('sendCounselStarter')
    expect(panel).toContain('autoSendOnOpen')
    expect(panel).toContain('chat-morning-starters')
    expect(panel).toContain('buildMorningWalkthroughStarters')
  })

  it('wires nudge review, offline field banner, and nudge strip', () => {
    expect(panel).toContain('onNudgeReview')
    expect(panel).toContain('ensureAwake')
    expect(panel).toContain('guardian-offline-field-banner')
    expect(panel).toContain('chat-offline-procedure-starters')
    expect(panel).toContain('GuardianNudgeStrip')
  })

  it('wires inference cost hint and grounded counsel label', () => {
    expect(panel).toContain('chat-counsel-cost-hint')
    expect(panel).toContain('grounded: counsel')
  })
})

describe('GuardianChatPanel — session memory + proposals', () => {
  it('wires topic chips, session close, and view-conversation handoff', () => {
    expect(panel).toContain('GuardianRecentTopicChip')
    expect(panel).toContain('session-topic-')
    expect(panel).toContain('/close')
    expect(panel).toContain('guardianPanel.viewConversationTick')
    expect(panel).toContain('onProposalViewConversation')
    expect(panel).toMatch(/viewConversationTick[\s\S]*message\.value = ''/)
    const chipsRowIdx = panel.indexOf("v-if=\"(s.topics || []).length\"")
    expect(chipsRowIdx).toBeGreaterThan(-1)
  })

  it('uses sessionDisplayLabel for pending session chip', () => {
    expect(panel).toContain('sessionDisplayLabel')
    expect(panel).toContain('sessionHasPendingProposal')
    expect(panel).toContain('data-test="session-pending-chip"')
    expect(panel).toContain('guardianProposals.fetch')
  })
})

describe('GuardianChatPanel — field assistant', () => {
  it('wires mic, camera, zone picker, and vision disclaimer', () => {
    expect(panel).toContain('chat-mic-button')
    expect(panel).toContain('chat-camera-button')
    expect(panel).toContain('chat-photo-zone-picker')
    expect(panel).toContain('chat-vision-disclaimer')
    expect(panel).toContain('chat-field-empty-hint')
  })
})

describe('GuardianChatPanel — first-run checklist lib', () => {
  it('imports firstRunChecklist for Guardian chat (not Dashboard checklist)', () => {
    expect(panel).toContain('firstRunChecklist.js')
    expect(existsSync(join(uiSrc, 'lib/firstRunChecklist.js'))).toBe(true)
  })
})
