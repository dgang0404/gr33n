/**
 * Phase 81 — Pi setup guide route restoration.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { WORKSPACES } from '../lib/workspaces.js'

describe('Phase 81 — Pi setup guide routing', () => {
  it('registers dedicated /pi-setup route (not zones fleet redirect)', () => {
    const router = readFileSync(join(process.cwd(), 'src/router/index.js'), 'utf8')
    expect(router).toMatch(/path:\s*'\/pi-setup'/)
    expect(router).toContain('PiSetupGuide')
  })

  it('help workspace includes Pi + HAT setup tab', () => {
    expect(WORKSPACES.help.tabs.some((t) => t.id === 'pi-setup')).toBe(true)
  })

  it('zones fleet tab links to pi-setup guide', () => {
    const zones = readFileSync(join(process.cwd(), 'src/views/workspaces/ZonesWorkspace.vue'), 'utf8')
    expect(zones).toContain('fleet-pi-setup-link')
    expect(zones).toContain("to=\"/pi-setup\"")
  })
})
