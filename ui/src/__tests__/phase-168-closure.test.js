/**
 * Phase 168 — Today cleanup: checklist removal + polish closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')
const uiSrc = join(process.cwd(), 'src')

describe('Phase 168 WS1 — remove getting-started checklist', () => {
  it('Dashboard uses canvas hero, not GettingStartedChecklist', () => {
    const dash = readFileSync(join(uiSrc, 'views/Dashboard.vue'), 'utf8')
    expect(dash).toContain('FarmCanvas')
    expect(dash).toContain('FarmZoneStack')
    expect(dash).not.toContain('GettingStartedChecklist')
    expect(dash).not.toContain('showFirstRunChecklist')
    expect(dash).not.toContain('firstRunDismissed')
    expect(existsSync(join(uiSrc, 'components/GettingStartedChecklist.vue'))).toBe(false)
  })

  it('keeps firstRunChecklist lib for Guardian chat', () => {
    const guardian = readFileSync(join(uiSrc, 'components/GuardianChatPanel.vue'), 'utf8')
    expect(guardian).toContain('firstRunChecklist.js')
    expect(existsSync(join(uiSrc, 'lib/firstRunChecklist.js'))).toBe(true)
  })
})

describe('Phase 168 WS2 — grower-native empty farm', () => {
  it('shows empty-farm Guardian setup chips on Dashboard', () => {
    const dash = readFileSync(join(uiSrc, 'views/Dashboard.vue'), 'utf8')
    expect(dash).toContain('showEmptyFarmStarters')
    expect(dash).toContain('emptyFarmStarters')
    expect(dash).toContain('buildSetupStarters')
    expect(dash).toContain('dashboard-empty-farm-starters')
  })
})

describe('Phase 168 WS3 — copy sweep', () => {
  it('uses formatZoneTypeLabel on Today tiles and quick actions', () => {
    const tile = readFileSync(join(uiSrc, 'components/FarmCanvasZoneTile.vue'), 'utf8')
    expect(tile).toContain('formatZoneTypeLabel')
    expect(tile).not.toContain("zone.zone_type || 'zone'")

    const sheet = readFileSync(join(uiSrc, 'components/ZoneQuickActions.vue'), 'utf8')
    expect(sheet).toContain('formatZoneTypeLabel')
    expect(sheet).not.toContain("zone.zone_type || 'zone'")

    const lib = readFileSync(join(uiSrc, 'lib/farmVisualStatus.js'), 'utf8')
    expect(lib).toContain('export function formatZoneTypeLabel')
  })
})

describe('Phase 168 WS4 — docs + tests', () => {
  it('documents visual farm cockpit in operator tour and current-state', () => {
    const tour = readFileSync(join(repoRoot, 'docs/operator-tour.md'), 'utf8')
    expect(tour).toContain('### 7k. Visual farm cockpit (Phases 164–168')
    expect(tour).toContain('Guardian setup chips')

    const state = readFileSync(join(repoRoot, 'docs/current-state.md'), 'utf8')
    expect(state).toContain('Phase 168')
    expect(state).toContain('getting-started checklist')
  })

  it('ships phase 168 test bundle', () => {
    expect(readFileSync(join(uiSrc, '__tests__/farmer-vocabulary-grow-path.test.js'), 'utf8')).toContain('FarmCanvas.vue')
    expect(readFileSync(join(uiSrc, '__tests__/first-run-checklist.test.js'), 'utf8')).not.toContain('GettingStartedChecklist')
  })
})
