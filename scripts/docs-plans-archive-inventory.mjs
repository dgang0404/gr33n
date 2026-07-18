#!/usr/bin/env node
// Phase 206 — read-only inventory for archiving docs/plans/*.plan.md.
//
// Splits every plan file into one of three batches by blast radius:
//   A  zero referrers anywhere in the repo         -> move only
//   B  referenced by non-test docs/code            -> move + fix those links
//   C  referenced by a ui/src/__tests__ *.test.js  -> move + fix links + fix
//      the test's literal path string + re-run that test file
//
// This script only reads and reports — it does not move anything. See
// docs/plans/phase_206_docs_plans_archive_migration.plan.md for the
// workstreams that consume this report.
//
// Usage:
//   node scripts/docs-plans-archive-inventory.mjs                 # summary
//   node scripts/docs-plans-archive-inventory.mjs --batch=A        # list batch A filenames
//   node scripts/docs-plans-archive-inventory.mjs --batch=B
//   node scripts/docs-plans-archive-inventory.mjs --batch=C        # also shows which test files touch each

import { readdirSync, readFileSync } from 'node:fs'
import { join, dirname } from 'node:path'
import { fileURLToPath } from 'node:url'
import { execSync } from 'node:child_process'

const repoRoot = join(dirname(fileURLToPath(import.meta.url)), '..')
const plansDir = join(repoRoot, 'docs/plans')
const testsDir = join(repoRoot, 'ui/src/__tests__')

// Hub / index docs kept at the top level of docs/plans/ regardless of
// reference count — they're navigation aids (era summaries), not closed
// single-phase records, and the new docs/roadmap/README.md links past them.
const KEEP_AT_TOP_LEVEL = new Set([
  'product_backlog_operator_runtime.plan.md',
  'pre_development_gaps_index.plan.md',
  'phase_84_100_master_roadmap.plan.md',
  'phase_68_73_spa_workspace_roadmap.plan.md',
  'farmer_ux_roadmap_40_plus.plan.md',
  'phase_53_59_roadmap.plan.md',
  'phase_173_177_today_excellence_roadmap.plan.md',
])

function planFiles() {
  return readdirSync(plansDir, { withFileTypes: true })
    .filter((d) => d.isFile() && d.name.endsWith('.plan.md'))
    .map((d) => d.name)
    .filter((name) => !KEEP_AT_TOP_LEVEL.has(name))
}

function testReferrers() {
  const refs = new Map() // plan filename -> Set(test filenames)
  for (const name of readdirSync(testsDir)) {
    if (!name.endsWith('.test.js')) continue
    const content = readFileSync(join(testsDir, name), 'utf8')
    for (const m of content.matchAll(/([\w.-]+\.plan\.md)/g)) {
      if (!refs.has(m[1])) refs.set(m[1], new Set())
      refs.get(m[1]).add(name)
    }
  }
  return refs
}

// One grep pass across the whole repo (excluding docs/plans self-refs and
// the tests dir, which is handled separately as batch C) instead of one
// subprocess per plan file — 199 spawns was the difference between this
// running in seconds vs. ~15s, which matters since a closure test shells
// out to this script.
function otherReferrersIndex() {
  let out
  try {
    out = execSync(
      `grep -rlE '[[:alnum:]_.-]+\\.plan\\.md' ` +
        `--include=*.md --include=*.go --include=*.js --include=*.vue ` +
        `--exclude-dir=plans --exclude-dir=__tests__ .`,
      { cwd: repoRoot, encoding: 'utf8', maxBuffer: 1024 * 1024 * 16 },
    )
  } catch {
    return new Map() // grep exits 1 on no matches
  }
  const index = new Map() // plan filename -> [referrer file, ...]
  for (const referrer of out.split('\n').filter(Boolean)) {
    const content = readFileSync(join(repoRoot, referrer), 'utf8')
    for (const m of content.matchAll(/([\w.-]+\.plan\.md)/g)) {
      if (!index.has(m[1])) index.set(m[1], [])
      if (!index.get(m[1]).includes(referrer)) index.get(m[1]).push(referrer)
    }
  }
  return index
}

const all = planFiles()
const testRefs = testReferrers()
const otherRefs = otherReferrersIndex()

const batchC = [] // test-referenced
const batchB = [] // other-doc-referenced only
const batchA = [] // zero referrers

for (const name of all) {
  if (testRefs.has(name)) {
    batchC.push(name)
  } else if ((otherRefs.get(name) ?? []).length > 0) {
    batchB.push(name)
  } else {
    batchA.push(name)
  }
}

const args = process.argv.slice(2)
const batchArg = args.find((a) => a.startsWith('--batch='))?.split('=')[1]

if (!batchArg) {
  console.log(`docs/plans/ total (excluding kept-at-top-level hubs): ${all.length}`)
  console.log(`  Batch A (zero referrers anywhere)        : ${batchA.length}`)
  console.log(`  Batch B (other docs/code, no tests)      : ${batchB.length}`)
  console.log(`  Batch C (referenced by closure tests)    : ${batchC.length} (touching ${new Set([...testRefs.values()].flatMap((s) => [...s])).size} test files)`)
  console.log(`  Kept at top level (hub/roadmap docs)      : ${KEEP_AT_TOP_LEVEL.size}`)
  console.log('\nRun with --batch=A, --batch=B, or --batch=C to list filenames.')
} else if (batchArg === 'A') {
  batchA.sort().forEach((n) => console.log(n))
} else if (batchArg === 'B') {
  for (const name of batchB.sort()) {
    console.log(`${name}  <-  ${(otherRefs.get(name) ?? []).sort().join(', ')}`)
  }
} else if (batchArg === 'C') {
  for (const name of batchC.sort()) {
    console.log(`${name}  <-  ${[...testRefs.get(name)].sort().join(', ')}`)
  }
} else {
  console.error(`Unknown --batch=${batchArg}, expected A, B, or C`)
  process.exit(1)
}
