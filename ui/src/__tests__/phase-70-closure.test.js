/**
 * Phase 70 — GPIO board + relay-HAT export closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const uiSrc = join(process.cwd(), 'src')

describe('Phase 70 — GPIO board closure', () => {
  it('GpioBoard loads farm data and renders relay rows', () => {
    const board = readFileSync(join(uiSrc, 'views/hardware/GpioBoard.vue'), 'utf8')
    expect(board).toContain('data-test="gpio-board"')
    expect(board).toContain('store.loadAll()')
    expect(board).toContain('gpio-board-relay-row')
    expect(board).toContain('Reference')
  })

  it('Hardware workspace uses GpioBoard instead of stub', () => {
    const hw = readFileSync(join(uiSrc, 'views/workspaces/HardwareWorkspace.vue'), 'utf8')
    expect(hw).toContain('GpioBoard')
    expect(hw).not.toContain('HardwareBoardStub')
    expect(hw).toContain('embedded')
  })

  it('PiSetupGuide uses loadAll and embedded reference banner', () => {
    const guide = readFileSync(join(uiSrc, 'views/PiSetupGuide.vue'), 'utf8')
    expect(guide).toContain('store.loadAll()')
    expect(guide).toContain('GPIO board')
    expect(guide).not.toContain('loadFarm')
  })
})
