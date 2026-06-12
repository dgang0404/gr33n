/**
 * Phase 54 WS4 / OC-54 — zone connection nav closure (Vitest bundle guard).
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'
import { relatedTo, NAV_RELATIONS } from '../lib/navRelations.js'
import { buildNavGroups, collectSidebarRoutes } from '../lib/navGroups.js'
import { resolvePipelineDeviceHint } from '../lib/zoneConnectionPipeline.js'

const uiSrc = join(process.cwd(), 'src')
const repoDocs = join(process.cwd(), '..', 'docs')

describe('Phase 54 WS4 / OC-54 — connection nav closure', () => {
  const navRoutes = new Set(collectSidebarRoutes(buildNavGroups()))

  it('documents operator-tour and architecture for connection pipeline', () => {
    const tour = readFileSync(join(repoDocs, 'operator-tour.md'), 'utf8')
    const arch = readFileSync(join(repoDocs, 'farm-guardian-architecture.md'), 'utf8')
    const plan = readFileSync(
      join(repoDocs, 'plans/phase_54_zone_connection_nav.plan.md'),
      'utf8',
    )
    expect(tour).toContain('connection pipeline')
    expect(tour).toContain('Phase 54')
    expect(arch).toContain('### 7.0r Zone connection nav (Phase 54 — shipped)')
    expect(plan).toMatch(/ws4-docs-tests[\s\S]*status: completed/)
    expect(plan).toContain('**Shipped.**')
  })

  it('OC-54 row is closed in operational closure doc', () => {
    const oc = readFileSync(
      join(repoDocs, 'plans/phase_35_37_operational_closure.plan.md'),
      'utf8',
    )
    expect(oc).toContain('## Phase 54 — Zone connection nav')
    expect(oc).toMatch(/oc-54-closure[\s\S]*status: completed/)
    expect(oc).toContain('phase-54-closure.test.js')
  })

  it('interactive pipeline segments use v-nav-hint', () => {
    const pipeline = readFileSync(join(uiSrc, 'components/ZoneConnectionPipeline.vue'), 'utf8')
    const needSection = readFileSync(join(uiSrc, 'components/ZoneNeedSection.vue'), 'utf8')
    const zoneDetail = readFileSync(join(uiSrc, 'views/ZoneDetail.vue'), 'utf8')
    expect(pipeline).toContain('v-nav-hint="seg.hint"')
    expect(pipeline).toContain('data-test="zone-connection-pipeline"')
    expect(needSection).toContain('ZoneConnectionPipeline')
    expect(zoneDetail).toContain('ZoneConnectionPipeline')
  })

  it('orphan link pass adds v-nav-hint on audited surfaces', () => {
    const audits = [
      ['views/Actuators.vue', 'actuator-zone-link', 'v-nav-hint'],
      ['views/Tasks.vue', 'task-zone-link', 'v-nav-hint'],
      ['components/ZoneNeedConnectionCard.vue', 'connection-card-details', 'v-nav-hint'],
      ['components/ZoneWaterGrowStory.vue', 'feeding-history-link', 'v-nav-hint'],
      ['components/ZoneAutomationPanel.vue', 'zone-rule-edit-automation', 'v-nav-hint'],
      ['components/TargetsRulesPanel.vue', 'greenhouse-templates-zone-link', 'v-nav-hint'],
    ]
    for (const [file, testId, hint] of audits) {
      const src = readFileSync(join(uiSrc, file), 'utf8')
      expect(src).toContain(`data-test="${testId}"`)
      expect(src).toContain(hint)
    }
  })

  it('navRelations expansion links tasks, alerts, fertigation, and grow paths', () => {
    expect(relatedTo('/tasks')).toContain('/zones')
    expect(relatedTo('/alerts')).toContain('/zones')
    expect(relatedTo('/fertigation')).toContain('/zones')
    expect(relatedTo('/fertigation')).toContain('/feed-water')
    expect(relatedTo('/operations/money')).toContain('/money')
    expect(relatedTo('/plants')).toContain('/zones')
  })

  it('navRelations only points at sidebar routes', () => {
    const legacyOk = new Set([
      '/feeding', '/fertigation', '/operations/feeding', '/operations/supplies', '/operations/money',
      '/sensors', '/actuators', '/lighting', '/pi-setup', '/costs', '/inventory',
      '/tasks', '/alerts', '/plants', '/schedules', '/automation', '/setpoints',
      '/feed-water', '/hardware',
    ])
    const workspaceTargets = new Set(['/feed-water', '/hardware'])
    for (const [from, targets] of Object.entries(NAV_RELATIONS)) {
      if (!navRoutes.has(from) && !legacyOk.has(from)) {
        expect(navRoutes.has(from), `missing nav route ${from}`).toBe(true)
      }
      for (const to of targets) {
        expect(navRoutes.has(to) || workspaceTargets.has(to), `${from} → ${to} not in sidebar`).toBe(true)
      }
    }
  })

  it('device hint prefers pi-setup when offline', () => {
    expect(resolvePipelineDeviceHint([{ status: 'offline' }])).toBe('/pi-setup')
  })

  it('closure Vitest files exist', () => {
    for (const f of [
      '__tests__/phase-54-closure.test.js',
      '__tests__/zone-connection-pipeline.test.js',
      'lib/zoneConnectionPipeline.js',
      'components/ZoneConnectionPipeline.vue',
    ]) {
      expect(existsSync(join(uiSrc, f)), `missing ${f}`).toBe(true)
    }
  })
})
