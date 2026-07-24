/**
 * Phase 80 — routing anchors, zones tab labels, workspace route helpers.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { WORKSPACES } from '../lib/workspaces.js'
import {
  COMFORT_ADVANCED_SCHEDULES_HASH,
  ZONE_HARDWARE_HASH,
  ZONE_WATER_PLAN_HASH,
  comfortAdvancedSchedulesRoute,
  naturalFarmingManageRoute,
  zoneHardwareRoute,
  zoneWaterPlanRoute,
} from '../lib/workspaceRoutes.js'

const uiSrc = join(process.cwd(), 'src')

describe('Phase 80 — routing & zones tab labels', () => {
  it('zones workspace uses farmer-friendly tab labels', () => {
    const tabs = WORKSPACES.zones.tabs
    expect(tabs.find((t) => t.id === 'rooms')?.label).toBe('My zones')
    expect(tabs.find((t) => t.id === 'fleet')?.label).toBe('Hardware & devices')
    expect(tabs.find((t) => t.id === 'plants')?.label).toBe('Plants')
    expect(WORKSPACES.zones.subtitle).toMatch(/hardware/i)
  })

  it('workspaceRoutes build hash targets for in-page navigation', () => {
    expect(comfortAdvancedSchedulesRoute().hash).toBe(COMFORT_ADVANCED_SCHEDULES_HASH)
    expect(comfortAdvancedSchedulesRoute().query.tab).toBe('schedules')
    expect(zoneWaterPlanRoute(2)).toMatchObject({
      path: '/zones/2',
      query: { tab: 'water' },
      hash: ZONE_WATER_PLAN_HASH,
    })
    expect(zoneHardwareRoute(5)).toMatchObject({
      path: '/zones/5',
      query: {},
      hash: ZONE_HARDWARE_HASH,
    })
    expect(naturalFarmingManageRoute({ inv: 'batches', batchId: 12 })).toEqual({
      path: '/natural-farming',
      query: { tab: 'manage', inv: 'batches', batch_id: '12' },
    })
  })

  it('Advanced schedules link targets comfort hash not legacy /schedules', () => {
    const panel = readFileSync(join(uiSrc, 'components/TargetsSchedulesPanel.vue'), 'utf8')
    expect(panel).toContain('comfortAdvancedSchedulesRoute')
    expect(panel).not.toMatch(/to="\/schedules"/)
    const comfort = readFileSync(join(uiSrc, 'views/workspaces/ComfortWorkspace.vue'), 'utf8')
    expect(comfort).toContain('id="comfort-advanced-schedules"')
  })

  it('feeding plan Details scroll target exists on zone water tab', () => {
    const story = readFileSync(join(uiSrc, 'components/ZoneWaterGrowStory.vue'), 'utf8')
    expect(story).toContain('id="zone-water-plan"')
    const need = readFileSync(join(uiSrc, 'components/ZoneNeedSection.vue'), 'utf8')
    expect(need).toContain('zoneWaterPlanRoute')
    expect(need).toContain('zoneHardwareRoute')
  })

  it('router enables scrollBehavior for hash navigation', () => {
    const router = readFileSync(join(uiSrc, 'router/index.js'), 'utf8')
    expect(router).toContain('scrollBehavior')
    expect(router).toContain('to.hash')
  })

  it('symptom guide lives in Help workspace, not as a standalone router page', () => {
    const router = readFileSync(join(uiSrc, 'router/index.js'), 'utf8')
    expect(router).not.toContain("component: SymptomGuide")
    const help = readFileSync(join(uiSrc, 'views/workspaces/HelpWorkspace.vue'), 'utf8')
    expect(help).toContain("activeTab === 'symptoms'")
    expect(help).toContain('SymptomGuide')
  })
})
