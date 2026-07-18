/**
 * Phase 204 — Docs navigation cleanup guardrails.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')
const repoDocs = join(repoRoot, 'docs')

describe('Phase 204 — docs navigation cleanup', () => {
  it('ships a single narrative roadmap doc', () => {
    const roadmap = readFileSync(join(repoDocs, 'roadmap/README.md'), 'utf8')
    expect(roadmap).toContain('## Eras, in order')
    expect(roadmap).toContain('Janitorial consolidation')
    expect(roadmap).toContain('204')
  })

  it('README explains the product before citing phase history', () => {
    const readme = readFileSync(join(repoRoot, 'README.md'), 'utf8')
    const introEnd = readme.indexOf('## What You Can Do')
    const intro = readme.slice(0, introEnd)
    expect(intro).toMatch(/Farm Guardian/)
    expect(intro).not.toMatch(/Phases? \d/)
    expect(readme).toContain('docs/roadmap/README.md')
  })

  it('Roadmap & history section leads with the roadmap doc', () => {
    const readme = readFileSync(join(repoRoot, 'README.md'), 'utf8')
    const section = readme.slice(readme.indexOf('## Roadmap & history'))
    expect(section.indexOf('docs/roadmap/README.md')).toBeLessThan(section.indexOf('phase-14-operator-documentation.md'))
  })

  it('duplicate Phase 60 doc pile stays deleted', () => {
    for (const name of [
      'OC-60-CLOSURE.md',
      'PHASE-60-BUILD-SUMMARY.md',
      'PHASE-60-QUICK-REFERENCE.md',
      'phase-60-implementation-checklist.md',
      'PHASE-60-IMPLEMENTATION-COMPLETE.md',
    ]) {
      expect(existsSync(join(repoDocs, name))).toBe(false)
    }
  })

  it('existing README-dependent closure tests still have their anchor strings', () => {
    const readme = readFileSync(join(repoRoot, 'README.md'), 'utf8')
    expect(readme).toContain('current-state.md')
    expect(readme).toContain('Product tiers')
    expect(readme).toContain('enterprise-tier-boundary.md')
    expect(readme).toContain('Phase 45')
    expect(readme).toContain('Phase 46')
    expect(readme).toContain('GUARDIAN_LLM_PROPOSALS')
  })
})
