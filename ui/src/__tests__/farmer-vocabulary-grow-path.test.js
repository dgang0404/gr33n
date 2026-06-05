import { describe, it, expect } from 'vitest'
import { readFileSync, readdirSync, statSync } from 'node:fs'
import { join, relative } from 'node:path'
import {
  findGrowPathVocabularyViolations,
  extractGrowPathScanText,
} from '../lib/farmerVocabulary.js'

const uiSrc = join(process.cwd(), 'src')

function growPathSourceFiles() {
  const files = []
  collectVue(join(uiSrc, 'views'), (rel) => {
    if (rel.startsWith('Zone') || rel === 'FeedingHub.vue' || rel === 'Dashboard.vue') {
      files.push(join(uiSrc, 'views', rel))
    }
  })
  collectVue(join(uiSrc, 'components'), (rel) => {
    if (rel.startsWith('Zone')) files.push(join(uiSrc, 'components', rel))
  })
  files.push(join(uiSrc, 'lib', 'plantNeeds.js'))
  return files.sort()
}

function collectVue(dir, accept) {
  for (const name of readdirSync(dir)) {
    const p = join(dir, name)
    if (statSync(p).isDirectory()) continue
    if (name.endsWith('.vue')) accept(name)
  }
}

function scanGrowPathVocabulary() {
  return growPathSourceFiles().map((file) => {
    const raw = readFileSync(file, 'utf8')
    const text = extractGrowPathScanText(raw)
    const violations = findGrowPathVocabularyViolations(text)
    return { file, violations }
  }).filter((r) => r.violations.length)
}

describe('Phase 47 WS5 — grow-path farmer vocabulary', () => {
  it('scans zone, feeding hub, dashboard, and plantNeeds copy', () => {
    const files = growPathSourceFiles()
    expect(files.some((f) => f.endsWith('FeedingHub.vue'))).toBe(true)
    expect(files.some((f) => f.endsWith('plantNeeds.js'))).toBe(true)
    expect(files.length).toBeGreaterThan(5)
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

  it('only scans vue templates, not script blocks', () => {
    const vue = `<template><p>Feed & water</p></template>
<script>
const x = { cron_expression: '0 8 * * *' }
</script>`
    expect(findGrowPathVocabularyViolations(extractGrowPathScanText(vue))).toEqual([])
  })
})
