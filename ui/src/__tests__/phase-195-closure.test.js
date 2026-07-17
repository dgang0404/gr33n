/**
 * Phase 195 — Pending inbox sticky count bar opaque (no scroll bleed).
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')

describe('Phase 195 — Pending inbox sticky count bleed', () => {
  it('GuardianRequestsInbox count row uses solid bg-zinc-950 without backdrop-blur', () => {
    const inbox = readFileSync(join(repoRoot, 'ui/src/components/GuardianRequestsInbox.vue'), 'utf8')
    expect(inbox).toContain('data-test="guardian-inbox-count"')
    expect(inbox).toContain('bg-zinc-950 border-b border-zinc-800/80')
    expect(inbox).not.toContain('bg-zinc-950/95')
    expect(inbox).not.toContain('backdrop-blur')
  })
})
