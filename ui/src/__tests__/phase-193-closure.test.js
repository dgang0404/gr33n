/**
 * Phase 193 — Help Library sticky section pills opaque (no scroll bleed).
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')

describe('Phase 193 — Help Library sticky nav bleed', () => {
  it('HelpLibraryHub section pills use solid bg via WorkspaceShell subnav (Phase 199)', () => {
    const hub = readFileSync(join(repoRoot, 'ui/src/views/HelpLibraryHub.vue'), 'utf8')
    expect(hub).not.toContain('help-library-jump')
    expect(hub).not.toContain('backdrop-blur')
    expect(hub).not.toContain('bg-zinc-950/95')

    const nav = readFileSync(join(repoRoot, 'ui/src/components/HelpLibrarySectionNav.vue'), 'utf8')
    expect(nav).toContain('data-test="help-library-jump"')

    const shell = readFileSync(join(repoRoot, 'ui/src/components/WorkspaceShell.vue'), 'utf8')
    expect(shell).toContain('workspace-shell__subnav')
    expect(shell).toContain('bg-zinc-950 border-b')
    expect(shell).not.toContain('bg-zinc-950/95')
    expect(shell).not.toMatch(/workspace-shell__subnav[\s\S]*backdrop-blur/)
  })
})
