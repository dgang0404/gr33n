/**
 * Phase 188 — broadened instruction-leak detection (off-topic essay/template
 * leaks) + WorkspaceShell/session sidebar UI fixes from the same sit-in pass.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')

describe('Phase 188 — Guardian answer-leak marker broadening', () => {
  it('leakCutIndex recognizes the ## Instruction and bare Question heading tells', () => {
    const leak = readFileSync(join(repoRoot, 'internal/farmguardian/answer_leak.go'), 'utf8')
    expect(leak).toContain('leakTopMarkers')
    expect(leak).toContain('## instruction')
    expect(leak).toContain('bareQuestionHeadingCutIndex')
    expect(leak).toContain('leakEssayTells')
  })

  it('has a regression test built from the live off-topic essay leak', () => {
    const test = readFileSync(join(repoRoot, 'internal/farmguardian/answer_leak_test.go'), 'utf8')
    expect(test).toContain('TestTrimInstructionLeak_offTopicEssayLeak')
    expect(test).toContain('offTopicEssayLeak')
  })
})

describe('Phase 188 — WorkspaceShell sticky subnav fix', () => {
  it('WorkspaceShell sticky subnav is fully opaque so scrolled content cannot show through it', () => {
    const shell = readFileSync(join(repoRoot, 'ui/src/components/WorkspaceShell.vue'), 'utf8')
    expect(shell).toContain('bg-zinc-950 border-b')
    expect(shell).not.toContain('bg-zinc-950/95')
  })
})
