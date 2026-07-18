#!/usr/bin/env node
/**
 * Phase 206 — one-shot docs/plans → docs/plans/archive migration.
 * Run from repo root: node scripts/migrate-plans-to-archive.mjs
 */
import { readdirSync, readFileSync, writeFileSync, unlinkSync, existsSync } from 'node:fs'
import { join, dirname } from 'node:path'
import { fileURLToPath } from 'node:url'
import { execSync } from 'node:child_process'

const repoRoot = join(dirname(fileURLToPath(import.meta.url)), '..')
const plansDir = join(repoRoot, 'docs/plans')
const archiveDir = join(plansDir, 'archive')

const KEEP_AT_TOP = new Set([
  'product_backlog_operator_runtime.plan.md',
  'pre_development_gaps_index.plan.md',
  'phase_84_100_master_roadmap.plan.md',
  'phase_68_73_spa_workspace_roadmap.plan.md',
  'farmer_ux_roadmap_40_plus.plan.md',
  'phase_53_59_roadmap.plan.md',
  'phase_173_177_today_excellence_roadmap.plan.md',
  'phase_205_pre_existing_test_debt.plan.md',
  'phase_206_docs_plans_archive_migration.plan.md',
])

// Phase 157 stubs — full content already in archive/; delete after link sweep.
const STUBS_TO_DELETE = [
  'phase_88_domain_enums_api.plan.md',
  'phase_89_lighting_presets_api_wiring.plan.md',
  'phase_90_device_taxonomy_registry.plan.md',
  'phase_91_bootstrap_template_catalog.plan.md',
  'phase_92_zone_greenhouse_vocabulary.plan.md',
  'phase_88_92_platform_data_gaps_roadmap.plan.md',
]

function walkFiles(dir, exts, out = []) {
  for (const ent of readdirSync(dir, { withFileTypes: true })) {
    const p = join(dir, ent.name)
    if (ent.name === 'node_modules' || ent.name === '.git' || ent.name === 'data') continue
    if (ent.isDirectory()) walkFiles(p, exts, out)
    else if (exts.some((e) => ent.name.endsWith(e))) out.push(p)
  }
  return out
}

const toMove = readdirSync(plansDir)
  .filter((n) => n.endsWith('.plan.md') && !KEEP_AT_TOP.has(n) && !STUBS_TO_DELETE.includes(n))

console.log(`Moving ${toMove.length} plan files to archive/…`)
for (const name of toMove.sort()) {
  const src = join(plansDir, name)
  const dest = join(archiveDir, name)
  if (existsSync(dest)) {
    console.log(`  skip (already in archive): ${name}`)
    unlinkSync(src)
    continue
  }
  execSync(`git mv ${JSON.stringify(src)} ${JSON.stringify(dest)}`, { cwd: repoRoot })
}

console.log(`Deleting ${STUBS_TO_DELETE.length} Phase 157 redirect stubs…`)
for (const name of STUBS_TO_DELETE) {
  const p = join(plansDir, name)
  if (existsSync(p)) {
    execSync(`git rm ${JSON.stringify(p)}`, { cwd: repoRoot })
  }
}

const archivedNames = new Set([
  ...readdirSync(archiveDir).filter((n) => n.endsWith('.plan.md')),
])

const refFiles = walkFiles(repoRoot, ['.md', '.js', '.go', '.vue', '.yaml', '.yml'])
let touched = 0

for (const file of refFiles) {
  if (file.startsWith(archiveDir)) continue
  let text = readFileSync(file, 'utf8')
  let changed = false
  for (const name of archivedNames) {
    if (KEEP_AT_TOP.has(name)) continue
    const patterns = [
      [`docs/plans/${name}`, `docs/plans/archive/${name}`],
      [`plans/${name}`, `plans/archive/${name}`],
    ]
    for (const [from, to] of patterns) {
      if (text.includes(from) && !text.includes(to)) {
        text = text.split(from).join(to)
        changed = true
      }
    }
  }
  // Fix accidental double archive/
  if (text.includes('plans/archive/archive/')) {
    text = text.split('plans/archive/archive/').join('plans/archive/')
    changed = true
  }
  if (changed) {
    writeFileSync(file, text)
    touched++
  }
}

console.log(`Updated references in ${touched} files.`)
console.log(`Remaining at docs/plans/: ${readdirSync(plansDir).filter((n) => n.endsWith('.plan.md')).join(', ')}`)
console.log(`Archive count: ${archivedNames.size} plan files.`)
