/**
 * Phase 45 WS4 — mobile sit-in path closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')
const repoDocs = join(repoRoot, 'docs')
const uiPublic = join(process.cwd(), 'public')

describe('Phase 45 WS4 — mobile sit-in path closure', () => {
  it('PWA manifest icons exist on disk', () => {
    const manifest = JSON.parse(
      readFileSync(join(uiPublic, 'manifest.webmanifest'), 'utf8'),
    )
    for (const icon of manifest.icons.filter((i) => i.type === 'image/png')) {
      const file = icon.src.replace(/^\//, '')
      expect(existsSync(join(uiPublic, file)), `missing ${file}`).toBe(true)
    }
  })

  it('mobile prep and cap LAN build scripts exist', () => {
    expect(existsSync(join(repoRoot, 'scripts/mobile-sit-in-prep.sh'))).toBe(true)
    expect(existsSync(join(repoRoot, 'scripts/cap-lan-build.sh'))).toBe(true)
    const prep = readFileSync(join(repoRoot, 'scripts/mobile-sit-in-prep.sh'), 'utf8')
    expect(prep).toContain('CORS_ORIGIN')
    expect(prep).toContain('phase-45-ws4-mobile-sit-in-path.md')
  })

  it('WS4 sit-in path doc documents PWA primary and store deferral', () => {
    const doc = readFileSync(
      join(repoDocs, 'workstreams/phase-45-ws4-mobile-sit-in-path.md'),
      'utf8',
    )
    expect(doc).toContain('PWA')
    expect(doc).toContain('Deferred')
    expect(doc).toContain('mobile-sit-in-prep.sh')
    expect(doc).toContain('farmer-sit-in-protocol.md')
  })

  it('operator-tour §10c documents mobile WS4', () => {
    const tour = readFileSync(join(repoDocs, 'operator-tour.md'), 'utf8')
    expect(tour).toContain('## 10c. Mobile distribution (Phase 45 WS4')
    expect(tour).toContain('mobile-sit-in-prep.sh')
    expect(tour).toContain('phase-45-ws4-mobile-sit-in-path.md')
  })

  it('mobile-distribution.md links Phase 45 WS4 path', () => {
    const mobile = readFileSync(join(repoDocs, 'mobile-distribution.md'), 'utf8')
    expect(mobile).toContain('Phase 45 WS4')
    expect(mobile).toContain('phase-45-ws4-mobile-sit-in-path.md')
  })

  it('phase 45 parent plan marks WS4 complete', () => {
    const plan = readFileSync(
      join(repoDocs, 'plans/archive/phase_45_farmer_validation_whole_app_polish.plan.md'),
      'utf8',
    )
    expect(plan).toMatch(/ws4-mobile-b4[\s\S]*status: completed/)
  })
})
