/**
 * Phase 73 — Guardian PR discoverability closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

describe('Phase 73 — guardian discoverability closure', () => {
  it('TopBar shows numeric pending badge and opens pending tab when count > 0', () => {
    const topbar = readFileSync(join(process.cwd(), 'src/components/TopBar.vue'), 'utf8')
    expect(topbar).toContain('data-test="topbar-guardian-pending-badge"')
    expect(topbar).toContain('proposalsStore.pendingCount')
    expect(topbar).toContain('openPendingTab')
    expect(topbar).toContain('refreshPendingCount')
  })

  it('GuardianNavLaunch opens pending inbox without duplicate pending badge (Phase 181)', () => {
    const nav = readFileSync(join(process.cwd(), 'src/components/GuardianNavLaunch.vue'), 'utf8')
    expect(nav).not.toContain('data-test="guardian-nav-pending-badge"')
    expect(nav).toContain('proposalsStore.pendingCount')
    expect(nav).toContain('openPendingTab')
    expect(nav).toContain('guardian-readiness-dot')
  })

  it('dismiss persists via POST /v1/chat/proposals/{id}/dismiss', () => {
    const vue = readFileSync(join(process.cwd(), 'src/components/GuardianActionProposal.vue'), 'utf8')
    expect(vue).toContain('/v1/chat/proposals/${local.proposal_id}/dismiss')
    const store = readFileSync(join(process.cwd(), 'src/stores/guardianProposals.js'), 'utf8')
    expect(store).toContain('dismissProposal')
    expect(store).toContain('/v1/chat/proposals/${proposalId}/dismiss')
  })

  it('empty zone nudge offers setup via suggest-empty-zone API', () => {
    const nudge = readFileSync(join(process.cwd(), 'src/components/EmptyZoneGrowNudge.vue'), 'utf8')
    expect(nudge).toContain('data-test="empty-zone-grow-nudge"')
    expect(nudge).toContain('/v1/chat/proposals/suggest-empty-zone')
    const zone = readFileSync(join(process.cwd(), 'src/views/ZoneDetail.vue'), 'utf8')
    expect(zone).toContain('EmptyZoneGrowNudge')
  })

  it('read-tool weather intent includes supplemental-light phrasing', () => {
    const weather = readFileSync(
      join(process.cwd(), '../internal/farmguardian/readtools_weather.go'),
      'utf8',
    )
    expect(weather).toContain('bright enough')
    expect(weather).toContain('Settings → Farm site')
  })

  it('registers dismiss and suggest-empty-zone routes', () => {
    const routes = readFileSync(join(process.cwd(), '../cmd/api/routes.go'), 'utf8')
    expect(routes).toContain('PostDismissProposal')
    expect(routes).toContain('PostSuggestEmptyZoneProposal')
    expect(routes).toContain('/v1/chat/proposals/{id}/dismiss')
    expect(routes).toContain('/v1/chat/proposals/suggest-empty-zone')
  })
})
