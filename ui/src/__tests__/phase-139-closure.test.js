/**
 * Phase 139 — Guardian docs, turn debugger & engineering CI closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoDocs = join(process.cwd(), '..', 'docs')
const repoRoot = join(process.cwd(), '..')

describe('Phase 139 — docs & turn debugger closure', () => {
  it('plan is shipped', () => {
    const plan = readFileSync(join(repoDocs, 'plans/phase_139_guardian_docs_and_engineering.plan.md'), 'utf8')
    expect(plan).toContain('**Shipped.**')
  })

  it('architecture doc profiles laptop vs server and links roadmap', () => {
    const arch = readFileSync(join(repoDocs, 'farm-guardian-architecture.md'), 'utf8')
    expect(arch).toContain('Profile A — Laptop dev')
    expect(arch).toContain('Profile D — Server')
    expect(arch).toContain('phase_129_139_guardian_next_level_roadmap.plan.md')
    expect(arch).not.toMatch(/Llama 3\.1 70B Q4 training weights/)
  })

  it('UI wires dev turn inspector', () => {
    const panel = readFileSync(join(process.cwd(), 'src/components/GuardianChatPanel.vue'), 'utf8')
    const debug = readFileSync(join(process.cwd(), 'src/components/GuardianTurnDebug.vue'), 'utf8')
    expect(panel).toContain('GuardianTurnDebug')
    expect(panel).toContain('showTurnDebug')
    expect(panel).toContain('lastTurnDebug')
    expect(panel).toContain('finalEvent.debug')
    expect(debug).toContain('data-test="guardian-turn-debug"')
  })

  it('Go exposes turn debug endpoint and builder', () => {
    const routes = readFileSync(join(repoRoot, 'cmd/api/routes.go'), 'utf8')
    const handler = readFileSync(join(repoRoot, 'internal/handler/chat/turn_debug.go'), 'utf8')
    const builder = readFileSync(join(repoRoot, 'internal/farmguardian/turn_debug.go'), 'utf8')
    expect(routes).toContain('/turns/{turn_index}/debug')
    expect(handler).toContain('GetTurnDebug')
    expect(handler).toContain('debugModeEnabled')
    expect(builder).toContain('BuildTurnDebug')
  })

  it('CI QA doc and closure checklist exist', () => {
    const ci = readFileSync(join(repoDocs, 'ci-guardian-qa.md'), 'utf8')
    const closure = readFileSync(join(repoDocs, 'plans/phase-129-139-closure.md'), 'utf8')
    expect(ci).toContain('guardian-qa-smoke')
    expect(ci).toContain('guardian_qa_runs')
    expect(closure).toContain('Phase 139')
  })

  it('INSTALL links Guardian roadmap', () => {
    const install = readFileSync(join(repoRoot, 'INSTALL.md'), 'utf8')
    expect(install).toContain('phase_129_139_guardian_next_level_roadmap.plan.md')
  })
})
