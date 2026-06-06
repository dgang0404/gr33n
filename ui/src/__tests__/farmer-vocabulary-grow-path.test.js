import { describe, it, expect } from 'vitest'
import { readFileSync, readdirSync, statSync } from 'node:fs'
import { join, relative } from 'node:path'
import {
  findGrowPathVocabularyViolations,
  extractGrowPathScanText,
  GROW_PATH_ZONE_LABELS,
  GROW_PATH_GENERIC_ROOM_BANS,
} from '../lib/farmerVocabulary.js'

const uiSrc = join(process.cwd(), 'src')

const GROW_PATH_VUE_VIEWS = [
  'Zones.vue',
  'FeedingHub.vue',
  'Dashboard.vue',
  'ComfortTargetsHub.vue',
  'Alerts.vue',
  'Tasks.vue',
  'FarmSetupWizard.vue',
  'ZoneSetupWizard.vue',
  'DeviceSetupWizard.vue',
  'LightingPrograms.vue',
]

const GROW_PATH_JS_LIBS = [
  'lib/plantNeeds.js',
  'lib/guardianStarters.js',
  'lib/guardianContextPrompts.js',
  'lib/navGroups.js',
  'lib/zoneFeedingPlan.js',
  'lib/zoneWaterGrowStory.js',
  'lib/farmGrowSummary.js',
  'lib/firstRunChecklist.js',
  'lib/farmSetupWizard.js',
  'lib/guardianRouteRef.js',
  'lib/feedingAdminHub.js',
]

function growPathSourceFiles() {
  const files = []
  collectVue(join(uiSrc, 'views'), (rel) => {
    if (rel.startsWith('Zone') || GROW_PATH_VUE_VIEWS.includes(rel)) {
      files.push(join(uiSrc, 'views', rel))
    }
  })
  collectVue(join(uiSrc, 'components'), (rel) => {
    if (rel.startsWith('Zone')) files.push(join(uiSrc, 'components', rel))
  })
  for (const rel of GROW_PATH_JS_LIBS) {
    files.push(join(uiSrc, rel))
  }
  return files.sort()
}

function collectVue(dir, accept) {
  for (const name of readdirSync(dir)) {
    const p = join(dir, name)
    if (statSync(p).isDirectory()) continue
    if (name.endsWith('.vue')) accept(name)
  }
}

function scanFile(file) {
  const raw = readFileSync(file, 'utf8')
  const text = file.endsWith('.vue') ? extractGrowPathScanText(raw) : raw
  const violations = findGrowPathVocabularyViolations(text)
  return { file, violations }
}

function scanGrowPathVocabulary() {
  return growPathSourceFiles()
    .map(scanFile)
    .filter((r) => r.violations.length)
}

describe('Phase 47 WS5 + Phase 45 WS3 — grow-path farmer vocabulary', () => {
  it('exports zone label map and room ban patterns', () => {
    expect(GROW_PATH_ZONE_LABELS.navMyZones).toBe('My zones')
    expect(GROW_PATH_ZONE_LABELS.mobileZones).toBe('Zones')
    expect(GROW_PATH_GENERIC_ROOM_BANS.some((b) => b.id === 'my-rooms')).toBe(true)
    expect(GROW_PATH_GENERIC_ROOM_BANS.some((b) => b.id === 'this-room')).toBe(true)
  })

  it('scans zone, feeding hub, dashboard, wizards, and farmer copy libs', () => {
    const files = growPathSourceFiles()
    expect(files.some((f) => f.endsWith('FeedingHub.vue'))).toBe(true)
    expect(files.some((f) => f.endsWith('guardianStarters.js'))).toBe(true)
    expect(files.some((f) => f.endsWith('navGroups.js'))).toBe(true)
    expect(files.length).toBeGreaterThan(15)
  })

  it('has no banned phrases on grow routes', () => {
    const failures = scanGrowPathVocabulary()
    if (failures.length) {
      const msg = failures.map(({ file, violations }) => {
        const rel = relative(uiSrc, file)
        const detail = violations.map((v) => `  - ${v.id}: "${v.match}" (${v.hint})`).join('\n')
        return `${rel}\n${detail}`
      }).join('\n\n')
      expect.fail(`Grow-path vocabulary violations:\n\n${msg}`)
    }
    expect(failures).toEqual([])
  })

  it('flags Setpoints → as a violation', () => {
    const hits = findGrowPathVocabularyViolations('Open Setpoints → for bands')
    expect(hits.some((h) => h.id === 'setpoints-arrow')).toBe(true)
  })

  it('flags generic My rooms and this room', () => {
    expect(findGrowPathVocabularyViolations('Sidebar My rooms')).toEqual(
      expect.arrayContaining([expect.objectContaining({ id: 'my-rooms' })]),
    )
    expect(findGrowPathVocabularyViolations('Alerts for this room')).toEqual(
      expect.arrayContaining([expect.objectContaining({ id: 'for-this-room' })]),
    )
  })

  it('allows zone display names containing Room', () => {
    expect(findGrowPathVocabularyViolations('When is the next feed for Flower Room?')).toEqual([])
    expect(findGrowPathVocabularyViolations('Start a grow in Veg Room')).toEqual([])
  })

  it('only scans vue templates, not script blocks', () => {
    const vue = `<template><p>Feed & water</p></template>
<script>
const x = { cron_expression: '0 8 * * *' }
</script>`
    expect(findGrowPathVocabularyViolations(extractGrowPathScanText(vue))).toEqual([])
  })
})
