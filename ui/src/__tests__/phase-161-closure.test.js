/**
 * Phase 161 — Guardian ec-ph smoke closure (trim + crop drift).
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')

describe('Phase 161 — ec-ph smoke closure', () => {
  it('TrimUncitedTail ships in chat finalize', () => {
    const finalize = readFileSync(
      join(repoRoot, 'internal/handler/chat/answer_finalize.go'),
      'utf8',
    )
    expect(finalize).toContain('TrimUncitedTail')
    expect(finalize).toContain('applyUncitedTailTrim')
    expect(finalize).toContain('uncited_tail_trimmed')
  })

  it('EcphCropDriftNote ships in farmguardian', () => {
    const align = readFileSync(
      join(repoRoot, 'internal/farmguardian/answer_citation_align.go'),
      'utf8',
    )
    expect(align).toContain('EcphCropDriftNote')
    expect(align).toContain('blueberry')
  })

  it('phase 161 plan marked shipped', () => {
    const plan = readFileSync(
      join(repoRoot, 'docs/plans/phase_161_guardian_ecph_smoke_closure.plan.md'),
      'utf8',
    )
    expect(plan).toContain('Status:** shipped')
  })
})
