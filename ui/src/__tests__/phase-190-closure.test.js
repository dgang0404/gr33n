/**
 * Phase 190 — dangling list-intro truncation detection + completion budget bump.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')

describe('Phase 190 — Guardian dangling list-intro truncation', () => {
  it('adds DanglingListIntroNote wired into AnswerAccuracyNote', () => {
    const src = readFileSync(join(repoRoot, 'internal/farmguardian/answer_accuracy.go'), 'utf8')
    expect(src).toContain('func DanglingListIntroNote')
    expect(src).toContain('DanglingListIntroNote(answer)')
  })

  it('has regression tests built from live truncated-answer turns', () => {
    const test = readFileSync(join(repoRoot, 'internal/farmguardian/answer_accuracy_test.go'), 'utf8')
    expect(test).toContain('TestDanglingListIntroNote_liveCalciumNitrateTask')
    expect(test).toContain('TestDanglingListIntroNote_liveFeedVolumeSetup')
  })

  it('bumps the default completion token budget 1024 -> 1536', () => {
    const chat = readFileSync(join(repoRoot, 'internal/rag/llm/chat.go'), 'utf8')
    expect(chat).toContain('return 1536')
    expect(chat).not.toContain('return 1024')
  })
})
