/**
 * Phase 181 — Guardian composer diet + single Ask gr33n badge closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const uiSrc = join(process.cwd(), 'src')

describe('Phase 181 — single pending badge + composer diet', () => {
  it('TopBar carries the pending count badge; sidebar launch does not', () => {
    const topbar = readFileSync(join(uiSrc, 'components/TopBar.vue'), 'utf8')
    const nav = readFileSync(join(uiSrc, 'components/GuardianNavLaunch.vue'), 'utf8')
    expect(topbar).toContain('data-test="topbar-guardian-pending-badge"')
    expect(nav).not.toContain('guardian-nav-pending-badge')
    expect(nav).toContain('guardian-readiness-dot')
  })

})
