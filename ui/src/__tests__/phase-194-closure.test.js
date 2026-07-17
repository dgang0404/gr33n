/**
 * Phase 194 — Pending proposal "View conversation".
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')

describe('Phase 194 — View conversation wiring', () => {
  it('guardianPanel exposes requestViewConversation without prefill', () => {
    const store = readFileSync(join(repoRoot, 'ui/src/stores/guardianPanel.js'), 'utf8')
    expect(store).toContain('viewConversationTick')
    expect(store).toContain('requestViewConversation(proposal)')
    expect(store).toContain("this.prefilledMessage = ''")
    expect(store).toContain('this.viewConversationTick += 1')
  })

  it('GuardianActionProposal renders View conversation when session_id is set', () => {
    const card = readFileSync(join(repoRoot, 'ui/src/components/GuardianActionProposal.vue'), 'utf8')
    expect(card).toContain('data-test="guardian-proposal-view-conversation"')
    expect(card).toContain('hasLinkedSession')
    expect(card).toContain("'view-conversation'")
    expect(card).toContain('View conversation')
  })

  it('FarmGuardianChat switches to Chat tab on view-conversation', () => {
    const page = readFileSync(join(repoRoot, 'ui/src/views/FarmGuardianChat.vue'), 'utf8')
    expect(page).toContain('@view-conversation="onInboxViewConversation"')
    expect(page).toContain('function onInboxViewConversation')
    expect(page).toContain("activeTab.value = 'chat'")
  })

  it('GuardianChatPanel watches viewConversationTick and clears composer', () => {
    const panel = readFileSync(join(repoRoot, 'ui/src/components/GuardianChatPanel.vue'), 'utf8')
    expect(panel).toContain('guardianPanel.viewConversationTick')
    expect(panel).toContain('onProposalViewConversation')
    expect(panel).toMatch(/viewConversationTick[\s\S]*message\.value = ''/)
  })
})
