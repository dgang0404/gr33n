#!/usr/bin/env node
/**
 * Phase 206 follow-up — sweep docs/plans/*-closure.md into docs/plans/archive/.
 * Phase 206 only moved *.plan.md; these shorter closure-summary docs (all
 * marked Shipped) kept accumulating at docs/plans/ root afterward. Same
 * one-shot mover + repo-wide link sweep as scripts/migrate-plans-to-archive.mjs.
 * Run from repo root: node scripts/migrate-closures-to-archive.mjs
 */
import { readdirSync, readFileSync, writeFileSync, existsSync } from 'node:fs'
import { join, dirname } from 'node:path'
import { fileURLToPath } from 'node:url'
import { execSync } from 'node:child_process'

const repoRoot = join(dirname(fileURLToPath(import.meta.url)), '..')
const plansDir = join(repoRoot, 'docs/plans')
const archiveDir = join(plansDir, 'archive')

function walkFiles(dir, exts, out = []) {
  for (const ent of readdirSync(dir, { withFileTypes: true })) {
    const p = join(dir, ent.name)
    if (ent.name === 'node_modules' || ent.name === '.git') continue
    if (ent.isDirectory()) walkFiles(p, exts, out)
    else if (exts.some((e) => ent.name.endsWith(e))) out.push(p)
  }
  return out
}

const EXTRA = [
  'local_dev_bugfix_todo.md',
  'phase_27_farm_guardian_ai_layer.md',
  'phase_28_crop_intelligence_guardian_depth.md',
  'phase_29_guardian_agent_layer.md',
  'phase_42_guardian_pr_spec.md',
  'phase_43_guardian_pr_spec.md',
  'phase_44_guardian_pr_spec.md',
  'phase_45_guardian_pr_spec.md',
  'phase_55_guardian_pr_spec.md',
]

const toMove = [
  ...readdirSync(plansDir).filter((n) => n.endsWith('-closure.md')),
  ...EXTRA.filter((n) => existsSync(join(plansDir, n))),
]

console.log(`Moving ${toMove.length} closure files to archive/…`)
for (const name of toMove.sort()) {
  execSync(`git mv ${JSON.stringify(join(plansDir, name))} ${JSON.stringify(join(archiveDir, name))}`, {
    cwd: repoRoot,
  })
}

const refFiles = walkFiles(repoRoot, ['.md', '.js', '.go', '.vue', '.yaml', '.yml'])
let touched = 0

for (const file of refFiles) {
  if (file.startsWith(archiveDir)) continue
  let text = readFileSync(file, 'utf8')
  let changed = false
  for (const name of toMove) {
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
console.log(`Remaining at docs/plans/: ${readdirSync(plansDir).filter((n) => n.endsWith('.md')).length} files`)
