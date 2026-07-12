import { describe, it, expect } from 'vitest'
import { starterPrefersFarmCounsel, starterShouldAutoSend } from '../lib/guardianStarterEntry.js'
import { buildMorningWalkthroughStarters, buildTodayAttentionStarters } from '../lib/guardianStarters.js'
import { buildSetupStarters } from '../lib/guardianStarters.js'

describe('Phase 170 — guardianStarterEntry', () => {
  it('prefers farm counsel for morning and attention starters', () => {
    const morning = buildMorningWalkthroughStarters({ surface: 'dashboard' })[0]
    expect(starterPrefersFarmCounsel(morning)).toBe(true)
    expect(starterShouldAutoSend(morning)).toBe(true)

    const attention = buildTodayAttentionStarters({
      zones: [{ id: 2, name: 'Flower Room' }],
      getStatus: () => ({ health: 'warn', attention: [{ label: 'Humidity high' }] }),
    })[0]
    expect(starterPrefersFarmCounsel(attention)).toBe(true)
    expect(starterShouldAutoSend(attention)).toBe(true)
  })

  it('prefills setup starters in farm counsel without auto-send', () => {
    const setup = buildSetupStarters({ surface: 'first_run_dashboard', farmId: 1, zoneCount: 0 })[0]
    expect(starterPrefersFarmCounsel(setup)).toBe(true)
    expect(starterShouldAutoSend(setup)).toBe(false)
  })

  it('does not auto-send inline panel picks', () => {
    const morning = buildMorningWalkthroughStarters({ surface: 'dashboard' })[0]
    expect(starterShouldAutoSend(morning, { inline: true })).toBe(false)
  })

  it('auto-sends zone-scoped starters from quick actions', () => {
    const zoneStarter = {
      id: 'zone-why',
      label: 'Why?',
      message: 'Why humidity high?',
      contextRef: { type: 'zone', id: 2, name: 'Flower Room' },
    }
    expect(starterPrefersFarmCounsel(zoneStarter)).toBe(true)
    expect(starterShouldAutoSend(zoneStarter)).toBe(true)
  })
})
