/**
 * Phase 156 — dependency & vulnerability scanning.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')

describe('Phase 156 — dependency scanning', () => {
  it('ships Dependabot config for Go and npm', () => {
    const yml = readFileSync(join(repoRoot, '.github/dependabot.yml'), 'utf8')
    expect(yml).toContain('package-ecosystem: gomod')
    expect(yml).toContain('package-ecosystem: npm')
    expect(yml).toContain('directory: /ui')
  })

  it('CI runs govulncheck and npm audit', () => {
    const ci = readFileSync(join(repoRoot, '.github/workflows/ci.yml'), 'utf8')
    expect(ci).toContain('govulncheck')
    expect(ci).toContain('npm audit --audit-level=high')
  })

  it('Makefile and script expose make vuln-check', () => {
    expect(existsSync(join(repoRoot, 'scripts/vuln-check.sh'))).toBe(true)
    const makefile = readFileSync(join(repoRoot, 'Makefile'), 'utf8')
    expect(makefile).toMatch(/^vuln-check:/m)
  })

  it('documents triage in SECURITY.md and vuln allowlist', () => {
    const sec = readFileSync(join(repoRoot, 'SECURITY.md'), 'utf8')
    expect(sec).toContain('Dependency vulnerabilities')
    expect(sec).toContain('vuln-allowlist.md')
    expect(existsSync(join(repoRoot, 'docs/vuln-allowlist.md'))).toBe(true)
  })
})
