/**
 * Phase 45 WS5 — edge module empty shells closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'

const uiSrc = join(process.cwd(), 'src')
const repoDocs = join(process.cwd(), '..', 'docs')

describe('Phase 45 WS5 — module shells closure', () => {
  it('moduleEmptyShell lib and component exist', () => {
    expect(existsSync(join(uiSrc, 'lib/moduleEmptyShell.js'))).toBe(true)
    expect(existsSync(join(uiSrc, 'components/ModuleEmptyShell.vue'))).toBe(true)
    expect(existsSync(join(uiSrc, '__tests__/module-empty-shell.test.js'))).toBe(true)
  })

  it('Animals and Aquaponics views use ModuleEmptyShell', () => {
    const animals = readFileSync(join(uiSrc, 'views/Animals.vue'), 'utf8')
    const aquaponics = readFileSync(join(uiSrc, 'views/Aquaponics.vue'), 'utf8')
    expect(animals).toContain('ModuleEmptyShell')
    expect(animals).toContain('module-id="animals"')
    expect(aquaponics).toContain('ModuleEmptyShell')
    expect(aquaponics).toContain('module-id="aquaponics"')
    expect(animals).not.toContain('No animal groups yet. Create one')
    expect(aquaponics).not.toContain('No aquaponics loops yet. Create one')
  })

  it('shell copy links to workflow-guide and pattern playbooks', () => {
    const lib = readFileSync(join(uiSrc, 'lib/moduleEmptyShell.js'), 'utf8')
    expect(lib).toContain('docs/workflow-guide.md')
    expect(lib).toContain('docs/pattern-playbooks.md')
    expect(lib).toContain('chicken_coop_v1')
    expect(lib).toContain('small_aquaponics_v1')
  })

  it('operator-tour documents livestock module shells', () => {
    const tour = readFileSync(join(repoDocs, 'operator-tour.md'), 'utf8')
    expect(tour).toContain('## 10a. Livestock modules (Phase 45 WS5')
    expect(tour).toContain('/animals')
    expect(tour).toContain('/aquaponics')
    expect(tour).toContain('workflow-guide.md')
  })

  it('phase 45 parent plan marks WS5 complete', () => {
    const plan = readFileSync(
      join(repoDocs, 'plans/phase_45_farmer_validation_whole_app_polish.plan.md'),
      'utf8',
    )
    expect(plan).toMatch(/ws5-module-shells[\s\S]*status: completed/)
  })
})
