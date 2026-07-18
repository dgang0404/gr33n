/**
 * Phase 202 — Closure test consolidation guardrails.
 */
import { describe, it, expect } from 'vitest'
import { readdirSync, readFileSync } from 'node:fs'
import { join } from 'node:path'

const testsDir = join(process.cwd(), 'src/__tests__')

function testFilesMatching(pattern) {
  return readdirSync(testsDir)
    .filter((name) => name.endsWith('.test.js'))
    .filter((name) => pattern.test(readFileSync(join(testsDir, name), 'utf8')))
}

describe('Phase 202 — closure test consolidation', () => {
  it('ships canonical GuardianChatPanel source test module', () => {
    const canonical = readFileSync(join(testsDir, 'guardian-chat-panel-source.test.js'), 'utf8')
    expect(canonical).toContain('GuardianChatPanel.vue')
    expect(canonical).toContain('chat-accuracy-banner')
  })

  it('ships testing-ui ownership doc', () => {
    const doc = readFileSync(join(process.cwd(), '..', 'docs/testing-ui.md'), 'utf8')
    expect(doc).toContain('guardian-chat-panel-source.test.js')
    expect(doc).toContain('today-excellence-arc.test.js')
  })

  it('phase-closure files touching GuardianChatPanel stay within budget', () => {
    const phaseGuardian = readdirSync(testsDir)
      .filter((name) => name.startsWith('phase-') && name.endsWith('-closure.test.js'))
      .filter((name) => readFileSync(join(testsDir, name), 'utf8').includes('GuardianChatPanel'))
    // Baseline was 18 phase-closure files; target ≥50% reduction → ≤9.
    expect(phaseGuardian.length).toBeLessThanOrEqual(9)
  })

  it('phase-closure files touching Dashboard.vue stay within budget', () => {
    const phaseDash = readdirSync(testsDir)
      .filter((name) => name.startsWith('phase-') && name.endsWith('-closure.test.js'))
      .filter((name) => readFileSync(join(testsDir, name), 'utf8').includes('Dashboard.vue'))
    // Baseline was 12 phase-closure files; target ≥40% reduction → ≤7.
    expect(phaseDash.length).toBeLessThanOrEqual(7)
  })

  it('only one test file readFileSync-scans GuardianChatPanel template', () => {
    const scanners = readdirSync(testsDir)
      .filter((name) => name.endsWith('.test.js') && name !== 'phase-202-closure.test.js')
      .filter((name) => /readFileSync\([^)]*GuardianChatPanel\.vue/.test(readFileSync(join(testsDir, name), 'utf8')))
    expect(scanners).toEqual(['guardian-chat-panel-source.test.js'])
  })

  it('Dashboard template scans live in today arc + workspace links', () => {
    const scanners = testFilesMatching(/readFileSync\([\s\S]*views\/Dashboard\.vue/)
    expect(scanners.sort()).toEqual([
      'dashboard-workspace-links.test.js',
      'phase-177-today-a11y.test.js',
      'today-excellence-arc.test.js',
    ])
  })
})
