/**
 * Phase 193 — Help Library sticky section pills opaque (no scroll bleed).
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')

describe('Phase 193 — Help Library sticky nav bleed', () => {
  it('HelpLibraryHub section pills use solid bg-zinc-950 without backdrop-blur', () => {
    const hub = readFileSync(join(repoRoot, 'ui/src/views/HelpLibraryHub.vue'), 'utf8')
    expect(hub).toContain('data-test="help-library-jump"')
    expect(hub).toContain('bg-zinc-950 border-b border-zinc-800/80')
    expect(hub).not.toContain('bg-zinc-950/95')
    expect(hub).not.toContain('backdrop-blur')
  })

  it('WorkspaceShell subnav stays opaque (Phase 188 regression)', () => {
    const shell = readFileSync(join(repoRoot, 'ui/src/components/WorkspaceShell.vue'), 'utf8')
    expect(shell).toContain('workspace-shell__subnav')
    expect(shell).toContain('bg-zinc-950 border-b')
    expect(shell).not.toContain('bg-zinc-950/95')
    expect(shell).not.toMatch(/workspace-shell__subnav[\s\S]*backdrop-blur/)
  })
})
