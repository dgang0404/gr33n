/**
 * Phase 63 WS5 / OC-63 — Guardian session memory closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'
import { buildContinueTopicPayload, topicChipLabel } from '../lib/guardianSessionMemory.js'

const repoDocs = join(process.cwd(), '..', 'docs')
const repoRoot = join(process.cwd(), '..')

describe('Phase 63 WS5 / OC-63 — session memory closure', () => {
  it('documents session_summaries and architecture section', () => {
    const arch = readFileSync(join(repoDocs, 'farm-guardian-architecture.md'), 'utf8')
    const plan = readFileSync(join(repoDocs, 'plans/phase_63_guardian_session_memory.plan.md'), 'utf8')
    expect(existsSync(join(repoRoot, 'internal/farmguardian/session_memory.go'))).toBe(true)
    expect(existsSync(join(repoRoot, 'db/migrations/20260611_phase63_session_summaries.sql'))).toBe(true)
    expect(arch).toContain('Session memory')
    expect(arch).toContain('session_summaries')
    expect(plan).toContain('**Shipped.**')
  })

  it('buildContinueTopicPayload frames a continue message', () => {
    const payload = buildContinueTopicPayload({
      summary_text: 'You asked about VPD in Flower Room.',
      topics: ['grow'],
      prompt: 'You recently asked about Grow — continue?',
    })
    expect(payload.message).toContain('continue')
    expect(payload.message).toContain('Grow')
    expect(topicChipLabel('comfort')).toBe('Comfort')
  })

  it('UI wires topic chips, recent chip, settings memory, and close endpoint', () => {
    const panel = readFileSync(join(process.cwd(), 'src/components/GuardianChatPanel.vue'), 'utf8')
    const settings = readFileSync(join(process.cwd(), 'src/views/Settings.vue'), 'utf8')
    const routes = readFileSync(join(repoRoot, 'cmd/api/routes.go'), 'utf8')
    expect(panel).toContain('GuardianRecentTopicChip')
    expect(panel).toContain('session-topic-')
    expect(panel).toContain('/close')
    expect(settings).toContain('settings-guardian-memory')
    expect(routes).toContain('POST /v1/chat/sessions/{session_id}/close')
    expect(routes).toContain('DELETE /farms/{id}/guardian-memory')
  })

  it('session memory helpers infer topics and prior context block', () => {
    const mem = readFileSync(join(repoRoot, 'internal/farmguardian/session_memory.go'), 'utf8')
    const handler = readFileSync(join(repoRoot, 'internal/handler/chat/session_memory.go'), 'utf8')
    expect(mem).toContain('InferSessionTopics')
    expect(mem).toContain('PriorSessionContextBlock')
    expect(handler).toContain('injectPriorSessionMemory')
  })
})
