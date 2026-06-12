/**
 * Phase 71 — Feed & Water workspace tabs.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { WORKSPACES, resolveWorkspaceTab } from '../lib/workspaces.js'
import router from '../router/index.js'

describe('Phase 71 — feed-water tabs', () => {
  it('declares four progressive-disclosure tabs', () => {
    const tabs = WORKSPACES.feedwater.tabs.map((t) => t.id)
    expect(tabs).toEqual(['daily', 'programs', 'nutrients', 'advanced'])
  })

  it('defaults to daily tab', () => {
    expect(resolveWorkspaceTab('feedwater', undefined)).toBe('daily')
    expect(resolveWorkspaceTab('feedwater', 'bogus')).toBe('daily')
  })

  it('FeedWaterWorkspace hosts hub views per tab', () => {
    const src = readFileSync(join(process.cwd(), 'src/views/workspaces/FeedWaterWorkspace.vue'), 'utf8')
    expect(src).toContain("activeTab === 'daily'")
    expect(src).toContain("activeTab === 'programs'")
    expect(src).toContain("activeTab === 'nutrients'")
    expect(src).toContain("activeTab === 'advanced'")
    expect(src).toContain('FeedingHub')
    expect(src).toContain('FeedingAdminHub')
    expect(src).toContain('Fertigation')
  })

  it('deep-links ?tab=programs on /feed-water', () => {
    const resolved = router.resolve({ path: '/feed-water', query: { tab: 'programs' } })
    expect(resolved.name).toBe('feed-water')
    expect(resolved.query.tab).toBe('programs')
  })
})
