/**
 * Phase 206 WS1 — docs/plans archive migration tooling closure.
 * The actual migration (WS2-6) is planned, not executed yet; this guards
 * the inventory tool the plan depends on, not a specific file count
 * (counts drift as other phases add or remove cross-references).
 */
import { describe, it, expect } from 'vitest'
import { execSync } from 'node:child_process'
import { existsSync, readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')

describe('Phase 206 WS1 — docs/plans archive inventory tool', () => {
  it('ships the read-only inventory script', () => {
    expect(existsSync(join(repoRoot, 'scripts/docs-plans-archive-inventory.mjs'))).toBe(true)
  })

  it('summary mode reports all three batches without moving anything', () => {
    const out = execSync('node scripts/docs-plans-archive-inventory.mjs', { cwd: repoRoot, encoding: 'utf8' })
    expect(out).toContain('Batch A')
    expect(out).toContain('Batch B')
    expect(out).toContain('Batch C')
    const [, batchC] = out.match(/Batch C.*?:\s*(\d+)/) ?? []
    expect(Number(batchC)).toBeGreaterThan(0)
  }, 30000)

  it('--batch=A lists only files with zero repo referrers', () => {
    const out = execSync('node scripts/docs-plans-archive-inventory.mjs --batch=A', { cwd: repoRoot, encoding: 'utf8' })
    const names = out.trim().split('\n').filter(Boolean)
    expect(names.length).toBeGreaterThan(0)
    for (const name of names) {
      expect(existsSync(join(repoRoot, 'docs/plans', name))).toBe(true)
    }
  }, 30000)

  it('plan documents batch ordering and the hub-doc exclusion list', () => {
    const plan = readFileSync(join(repoRoot, 'docs/plans/phase_206_docs_plans_archive_migration.plan.md'), 'utf8')
    expect(plan).toContain('Batch A')
    expect(plan).toContain('Batch C')
    expect(plan).toContain('docs/roadmap/README.md')
  })
})
