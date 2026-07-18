/**
 * Phase 157 — docs consolidation (current-state + archive).
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')
const docs = join(repoRoot, 'docs')

describe('Phase 157 — docs consolidation', () => {
  it('ships current-state.md with core sections', () => {
    const cs = readFileSync(join(docs, 'current-state.md'), 'utf8')
    for (const needle of [
      'Farm Guardian',
      'make guardian-qa-smoke',
      'gr33ncore',
      'plans/archive',
    ]) {
      expect(cs).toContain(needle)
    }
  })

  it('archives phase 88-92 under docs/plans/archive/ (Phase 206 removed stubs)', () => {
    expect(existsSync(join(docs, 'plans/archive/phase_88_domain_enums_api.plan.md'))).toBe(true)
    expect(existsSync(join(docs, 'plans/phase_88_domain_enums_api.plan.md'))).toBe(false)
  })

  it('phase-14 index links current-state and archive', () => {
    const index = readFileSync(join(docs, 'phase-14-operator-documentation.md'), 'utf8')
    expect(index).toContain('current-state.md')
    expect(index).toContain('plans/archive/')
    expect(index).toContain('Start here (Phase 157)')
  })

  it('README and INSTALL link current-state', () => {
    expect(readFileSync(join(repoRoot, 'README.md'), 'utf8')).toContain('current-state.md')
    expect(readFileSync(join(repoRoot, 'INSTALL.md'), 'utf8')).toContain('current-state.md')
  })

  it('Makefile exposes docs-current-state-hint', () => {
    const makefile = readFileSync(join(repoRoot, 'Makefile'), 'utf8')
    expect(makefile).toMatch(/^docs-current-state-hint:/m)
  })
})
