/**
 * Phase 45 WS6 — light accessibility closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'

const uiSrc = join(process.cwd(), 'src')
const repoDocs = join(process.cwd(), '..', 'docs')

describe('Phase 45 WS6 — a11y closure', () => {
  it('farmerA11y lib and tests exist', () => {
    expect(existsSync(join(uiSrc, 'lib/farmerA11y.js'))).toBe(true)
    expect(existsSync(join(uiSrc, '__tests__/farmer-a11y.test.js'))).toBe(true)
  })

  it('global focus-visible styles ship in style.css', () => {
    const css = readFileSync(join(uiSrc, 'style.css'), 'utf8')
    expect(css).toContain('Phase 45 WS6')
    expect(css).toContain('button:focus-visible')
    expect(css).toContain('ring-green-500')
  })

  it('Guardian proposal and starter chips wire aria-labels', () => {
    const proposal = readFileSync(join(uiSrc, 'components/GuardianActionProposal.vue'), 'utf8')
    const chips = readFileSync(join(uiSrc, 'components/GuardianStarterChips.vue'), 'utf8')
    expect(proposal).toContain('guardianProposalAriaLabel')
    expect(proposal).toContain(':aria-label="dismissAriaLabel"')
    expect(chips).toContain('Ask Guardian:')
    expect(chips).toContain('min-h-[44px]')
  })

  it('zone water grow story Run feed now has aria-label', () => {
    const story = readFileSync(join(uiSrc, 'components/ZoneWaterGrowStory.vue'), 'utf8')
    expect(story).toContain('runFeedNowAriaLabel')
    expect(story).toContain(':aria-label="runNowLabel"')
  })

  it('wizards mark current step with aria-current', () => {
    for (const f of ['views/FarmSetupWizard.vue', 'views/ZoneSetupWizard.vue', 'views/DeviceSetupWizard.vue']) {
      const vue = readFileSync(join(uiSrc, f), 'utf8')
      expect(vue).toContain("aria-current")
    }
  })

  it('operator-tour documents WS6 light a11y pass', () => {
    const tour = readFileSync(join(repoDocs, 'operator-tour.md'), 'utf8')
    expect(tour).toContain('## 10b. Light accessibility (Phase 45 WS6')
    expect(tour).toContain('focus-visible')
  })

  it('phase 45 parent plan marks WS6 complete', () => {
    const plan = readFileSync(
      join(repoDocs, 'plans/phase_45_farmer_validation_whole_app_polish.plan.md'),
      'utf8',
    )
    expect(plan).toMatch(/ws6-accessibility-pass[\s\S]*status: completed/)
  })
})
