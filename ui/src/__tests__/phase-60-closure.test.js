/**
 * Phase 60 WS5 / OC-60 — morning walkthrough closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'
import { buildMorningWalkthroughStarters } from '../lib/guardianStarters.js'

const repoDocs = join(process.cwd(), '..', 'docs')
const repoRoot = join(process.cwd(), '..')

describe('Phase 60 WS5 / OC-60 — morning walkthrough closure', () => {
  it('documents walk_farm read tool and architecture section', () => {
    const arch = readFileSync(join(repoDocs, 'farm-guardian-architecture.md'), 'utf8')
    const plan = readFileSync(join(repoDocs, 'plans/archive/phase_60_guardian_morning_walkthrough.plan.md'), 'utf8')
    expect(existsSync(join(repoRoot, 'internal/farmguardian/readtools_walk.go'))).toBe(true)
    expect(arch).toContain('walk_farm')
    expect(plan).toContain('**Shipped.**')
  })

  it('dashboard morning check starter uses walk_farm and morning_walkthrough mode', () => {
    const starters = buildMorningWalkthroughStarters({ surface: 'dashboard', farmName: 'Demo Farm' })
    expect(starters).toHaveLength(1)
    expect(starters[0].label).toBe('Morning check')
    expect(starters[0].message).toContain('walk_farm')
    expect(starters[0].message).toContain('Demo Farm')
    expect(starters[0].contextRef.guardian_mode).toBe('morning_walkthrough')
    expect(starters[0].contextRef.path).toBe('/')
  })

  it('chat page exposes morning walkthrough starter', () => {
    const starters = buildMorningWalkthroughStarters({ surface: 'chat' })
    expect(starters[0].label).toBe('Morning walkthrough')
    expect(starters[0].contextRef.path).toBe('/chat')
  })

  it('UI wires morning walkthrough starters via lib helpers', () => {
    const starters = readFileSync(join(process.cwd(), 'src/lib/guardianStarters.js'), 'utf8')
    expect(starters).toContain('buildMorningWalkthroughStarters')
  })

  it('context_ref supports guardian_mode morning walkthrough', () => {
    const ctx = readFileSync(join(repoRoot, 'internal/farmguardian/context_ref.go'), 'utf8')
    expect(ctx).toContain('GuardianMode')
    expect(ctx).toContain('morning_walkthrough')
    expect(ctx).toContain('morningWalkthroughContextBlock')
  })
})
