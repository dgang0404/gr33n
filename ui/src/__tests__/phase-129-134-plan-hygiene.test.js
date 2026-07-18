/**
 * Phases 129–134 — plan doc hygiene (shipped status).
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const plans = join(process.cwd(), '..', 'docs', 'plans', 'archive')

const phaseFiles = [
  'phase_129_guardian_awakening.plan.md',
  'phase_130_guardian_runtime_orchestration.plan.md',
  'phase_131_guardian_qa_harness.plan.md',
  'phase_132_guardian_read_tool_router.plan.md',
  'phase_133_guardian_answer_grounding_honesty.plan.md',
  'phase_134_guardian_answer_feedback.plan.md',
]

describe('Phases 129–134 — plan hygiene', () => {
  for (const file of phaseFiles) {
    it(`${file} is shipped with completed todos`, () => {
      const text = readFileSync(join(plans, file), 'utf8')
      expect(text).toContain('**Shipped.**')
      expect(text).not.toMatch(/status: pending/)
    })
  }
})
