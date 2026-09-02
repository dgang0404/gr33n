/**
 * Phase 217 WS8 — clickable automation label wiring (static-analysis style,
 * matching the existing phase-80-closure.test.js convention for these same
 * two panel files rather than a full component mount).
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const uiSrc = join(process.cwd(), 'src')

describe('Phase 217 WS3 — automation label resource linking', () => {
  it('TargetsSchedulesPanel renders the schedule name as a v-nav-hint link when a target resolves', () => {
    const panel = readFileSync(join(uiSrc, 'components/TargetsSchedulesPanel.vue'), 'utf8')
    expect(panel).toContain("import { resolveScheduleTarget } from '../lib/automationTargetLink.js'")
    expect(panel).toContain('scheduleTarget(schedule)')
    expect(panel).toContain('v-nav-hint="scheduleTarget(schedule).navHint"')
    // Unresolvable targets must fall back to plain text, never a dead link.
    expect(panel).toMatch(/v-else[^>]*>\{\{\s*schedule\.name\s*\}\}/)
  })

  it('TargetsRulesPanel renders the rule name as a v-nav-hint link when a target resolves', () => {
    const panel = readFileSync(join(uiSrc, 'components/TargetsRulesPanel.vue'), 'utf8')
    expect(panel).toContain("import { resolveRuleTarget } from '../lib/automationTargetLink.js'")
    expect(panel).toContain('ruleTarget(rule)')
    expect(panel).toContain('v-nav-hint="ruleTarget(rule).navHint"')
    expect(panel).toMatch(/v-else[^>]*>\{\{\s*rule\.name\s*\}\}/)
  })
})
