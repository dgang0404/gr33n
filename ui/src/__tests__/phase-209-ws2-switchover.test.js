/**
 * Phase 209 WS2 — switchover wizard UI wiring.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')
const workspace = readFileSync(
  join(process.cwd(), 'src/views/workspaces/NaturalFarmingWorkspace.vue'),
  'utf8',
)
const wizard = readFileSync(
  join(process.cwd(), 'src/components/naturalfarming/SwitchoverWizard.vue'),
  'utf8',
)
const canonLib = readFileSync(join(process.cwd(), 'src/lib/naturalFarmingCanon.js'), 'utf8')

describe('Phase 209 WS2 — switchover wizard component', () => {
  it('switchover wizard component still exists for pack/bootstrap helpers', () => {
    expect(wizard).toContain('loadRecipeCanon')
    expect(wizard).toContain('nf-cta-apply-bootstrap')
  })

  it('workspace no longer mounts switchover wizard', () => {
    expect(workspace).not.toContain('SwitchoverWizard')
  })

  it('wizard implements five steps and canonical mapping UI', () => {
    expect(wizard).toContain('data-test="nf-switchover-wizard"')
    expect(wizard).toContain('nf-switchover-step-context')
    expect(wizard).toContain('nf-switchover-step-pattern')
    expect(wizard).toContain('nf-switchover-step-mapping')
    expect(wizard).toContain('nf-switchover-step-first-batch')
    expect(wizard).toContain('nf-switchover-step-actions')
    expect(wizard).toContain('resolveSwitchoverMapping')
    expect(wizard).toContain('firstBatchSuggestions')
  })

  it('CTAs link to batch tab and bootstrap apply', () => {
    expect(wizard).toContain('nf-cta-make-batch')
    expect(wizard).toContain('Ready to ferment? → Make a batch')
    expect(wizard).not.toContain('nf-cta-make-jms')
    expect(wizard).toContain('nf-cta-apply-bootstrap')
    expect(wizard).toContain('applyBootstrapTemplate')
    expect(wizard).toContain('bootstrapTemplateForContext')
  })

  it('Learn how expander links to field guides', () => {
    expect(wizard).toContain('LearnHowExpander')
    const expander = readFileSync(
      join(process.cwd(), 'src/components/naturalfarming/LearnHowExpander.vue'),
      'utf8',
    )
    expect(expander).toContain('fieldGuideLearnRoute')
    const switchover = readFileSync(join(process.cwd(), 'src/lib/naturalFarmingSwitchover.js'), 'utf8')
    expect(switchover).toContain('cited_doc')
  })

  it('step 5 is seed farm optional not Apply', () => {
    expect(wizard).toContain('Seed farm (optional)')
    expect(wizard).toContain("actions: 'Seed farm (optional)'")
  })

  it('links to recipe library tab from switchover guide', () => {
    expect(wizard).toContain('nf-switchover-library-link')
    expect(wizard).toContain("tab: 'library'")
  })
})
