/**
 * Phase 155 — automated backups (scripts + Makefile targets).
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')

describe('Phase 155 — automated backups', () => {
  it('ships backup and verify scripts', () => {
    expect(existsSync(join(repoRoot, 'scripts/backup-gr33n.sh'))).toBe(true)
    expect(existsSync(join(repoRoot, 'scripts/verify-backup-gr33n.sh'))).toBe(true)
    const backup = readFileSync(join(repoRoot, 'scripts/backup-gr33n.sh'), 'utf8')
    expect(backup).toContain('pg_dump')
    expect(backup).toContain('manifest.json')
    const verify = readFileSync(join(repoRoot, 'scripts/verify-backup-gr33n.sh'), 'utf8')
    expect(verify).toContain('createdb')
    expect(verify).toContain('gr33ncore.farms')
  })

  it('Makefile exposes make backup and make verify-backup', () => {
    const makefile = readFileSync(join(repoRoot, 'Makefile'), 'utf8')
    expect(makefile).toMatch(/^backup:/m)
    expect(makefile).toMatch(/^verify-backup:/m)
  })

  it('runbook documents Phase 155 automation', () => {
    const runbook = readFileSync(join(repoRoot, 'docs/backup-restore-runbook.md'), 'utf8')
    expect(runbook).toContain('make backup')
    expect(runbook).toContain('make verify-backup')
  })
})
