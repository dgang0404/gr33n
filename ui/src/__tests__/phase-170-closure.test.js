/**
 * Phase 170 — Today Guardian one-tap farm counsel closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')
const uiSrc = join(process.cwd(), 'src')

describe('Phase 170 — one-tap farm counsel wiring', () => {
  it('ships guardianStarterEntry lib and panel store flags', () => {
    const lib = readFileSync(join(uiSrc, 'lib/guardianStarterEntry.js'), 'utf8')
    expect(lib).toContain('starterPrefersFarmCounsel')
    expect(lib).toContain('starterShouldAutoSend')

    const store = readFileSync(join(uiSrc, 'stores/guardianPanel.js'), 'utf8')
    expect(store).toContain('preferFarmCounsel')
    expect(store).toContain('autoSendOnOpen')
  })

  it('wires chips, quick actions, and chat panel auto-send', () => {
    const chips = readFileSync(join(uiSrc, 'components/GuardianStarterChips.vue'), 'utf8')
    expect(chips).toContain('guardianStarterEntry')
    expect(chips).toContain('farmCounsel')

    const sheet = readFileSync(join(uiSrc, 'components/ZoneQuickActions.vue'), 'utf8')
    expect(sheet).toContain('autoSend: true')
  })

  it('documents phase in current-state', () => {
    const state = readFileSync(join(repoRoot, 'docs/current-state.md'), 'utf8')
    expect(state).toContain('Phase 170')
  })
})
