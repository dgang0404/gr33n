/**
 * Phase 207 — Natural farming studio arc closure (Phases 207–211).
 * Roadmap sign-off: child phases shipped, closure tests present, north-star wiring.
 */
import { describe, it, expect } from 'vitest'
import { existsSync, readFileSync } from 'node:fs'
import { join } from 'node:path'
import { buildNavGroups } from '../lib/navGroups.js'
import { WORKSPACES } from '../lib/workspaces.js'

const repoRoot = join(process.cwd(), '..')
const repoDocs = join(repoRoot, 'docs')
const uiTests = join(process.cwd(), 'src/__tests__')

const ARC_PLANS = [
  'phase_207_natural_farming_studio.plan.md',
  'phase_208_natural_farming_process_knowledge.plan.md',
  'phase_209_natural_farming_studio_ui.plan.md',
  'phase_210_natural_farming_guardian_integration.plan.md',
  'phase_211_natural_farming_switchover_commons.plan.md',
]

const CLOSURE_TESTS = [
  'phase-208-closure.test.js',
  'phase-209-closure.test.js',
  'phase-210-closure.test.js',
  'phase-211-closure.test.js',
]

describe('Phase 207 — natural farming arc closure (207–211)', () => {
  it('roadmap and child plans mark the arc shipped', () => {
    const roadmap = readFileSync(join(repoDocs, 'plans', ARC_PLANS[0]), 'utf8')
    expect(roadmap).toMatch(/Shipped \(207–211\)/)
    for (const file of ARC_PLANS.slice(1)) {
      const plan = readFileSync(join(repoDocs, 'plans', file), 'utf8')
      expect(plan).toMatch(/Shipped|shipped/)
    }
  })

  it('each phase has a closure test on disk', () => {
    for (const name of CLOSURE_TESTS) {
      expect(existsSync(join(uiTests, name))).toBe(true)
    }
    expect(existsSync(join(uiTests, 'phase-207-closure.test.js'))).toBe(true)
  })

  it('sidebar placement matches roadmap (Grow & operate → Natural farming)', () => {
    const grow = buildNavGroups().find((g) => g.label === 'Grow & operate')
    const idxZones = grow.items.findIndex((i) => i.to === '/zones')
    const idxNf = grow.items.findIndex((i) => i.to === '/natural-farming')
    const idxComfort = grow.items.findIndex((i) => i.to === '/comfort-targets')
    expect(idxNf).toBeGreaterThan(idxZones)
    expect(idxNf).toBeLessThan(idxComfort)
    expect(WORKSPACES.naturalfarming.route).toBe('/natural-farming')
  })

  it('208 process catalog + 210 Guardian tools + 211 packs are wired', () => {
    expect(readFileSync(join(repoRoot, 'data/process-material-catalog.yaml'), 'utf8')).toContain(
      'goldenrod',
    )
    const readtools = readFileSync(join(repoRoot, 'internal/farmguardian/readtools_naturalfarming.go'), 'utf8')
    expect(readtools).toContain('SuggestProcessFromMaterial')
    expect(readFileSync(join(repoRoot, 'cmd/api/routes.go'), 'utf8')).toContain(
      'POST /farms/{id}/naturalfarming/apply-pack',
    )
    expect(readFileSync(join(repoDocs, 'pattern-playbooks.md'), 'utf8')).toContain(
      'Natural farming recipe packs (Phase 211',
    )
  })

  it('smoke suite keeps cherry-forest step 1 and adds smoke-cherry-jlf step 5', () => {
    const smoke = readFileSync(join(repoRoot, 'internal/farmguardian/eval/fixtures_smoke.go'), 'utf8')
    const score = readFileSync(join(repoRoot, 'internal/farmguardian/eval/score.go'), 'utf8')
    expect(smoke.indexOf('smoke-cherry-forest')).toBeLessThan(smoke.indexOf('smoke-cherry-jlf'))
    expect(score).toContain(`in.Question.ID == "smoke-cherry-forest"`)
    expect(score).toContain('smoke-cherry-jlf')
  })

  it('operator tour and current-state document the studio arc', () => {
    const tour = readFileSync(join(repoDocs, 'operator-tour.md'), 'utf8')
    const state = readFileSync(join(repoDocs, 'current-state.md'), 'utf8')
    expect(tour).toMatch(/7u\. Natural farming studio/i)
    expect(tour).toContain('phase_207_natural_farming_studio.plan.md')
    expect(state).toContain('207–211')
  })
})
