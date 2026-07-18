#!/usr/bin/env node
// Phase 205 — fails the build only on *new* UI test failures, not on the
// pre-existing debt tracked in ui/test-baseline-known-failures.json.
//
// Why: `npm test -- --run` already failed on 14 files before this script
// existed, and nobody noticed because CI/reviewers only ran the new
// phase's own test file. This makes "did I just break something" a single
// exit code instead of a manual diff against memory.
//
// Usage: node scripts/check-ui-test-baseline.mjs

import { spawnSync } from 'node:child_process'
import { readFileSync, writeFileSync, unlinkSync } from 'node:fs'
import { join, dirname } from 'node:path'
import { fileURLToPath } from 'node:url'

const repoRoot = join(dirname(fileURLToPath(import.meta.url)), '..')
const uiDir = join(repoRoot, 'ui')
const baselinePath = join(uiDir, 'test-baseline-known-failures.json')
const tmpReport = join(uiDir, '.vitest-baseline-report.json')

const baseline = JSON.parse(readFileSync(baselinePath, 'utf8')).failures

const result = spawnSync(
  'npx',
  ['vitest', 'run', '--reporter=json', `--outputFile=${tmpReport}`],
  { cwd: uiDir, stdio: ['ignore', 'pipe', 'inherit'] },
)

let report
try {
  report = JSON.parse(readFileSync(tmpReport, 'utf8'))
} finally {
  try { unlinkSync(tmpReport) } catch { /* best effort cleanup */ }
}

const currentFailures = {}
for (const fileResult of report.testResults ?? []) {
  const rel = fileResult.name.replace(uiDir + '/', '')
  const fails = (fileResult.assertionResults ?? [])
    .filter((a) => a.status === 'failed')
    .map((a) => a.fullName)
  if (fails.length) currentFailures[rel] = fails
}

const newFailures = []
for (const [file, tests] of Object.entries(currentFailures)) {
  const known = new Set(baseline[file] ?? [])
  for (const t of tests) {
    if (!known.has(t)) newFailures.push(`${file} :: ${t}`)
  }
}

const nowFixed = []
for (const [file, tests] of Object.entries(baseline)) {
  const stillFailing = new Set(currentFailures[file] ?? [])
  for (const t of tests) {
    if (!stillFailing.has(t)) nowFixed.push(`${file} :: ${t}`)
  }
}

if (nowFixed.length) {
  console.log(`\n${nowFixed.length} baseline failure(s) now pass — remove from test-baseline-known-failures.json when convenient:`)
  for (const f of nowFixed) console.log(`  fixed: ${f}`)
}

if (newFailures.length) {
  console.error(`\n✗ ${newFailures.length} NEW UI test failure(s) not in the known baseline:`)
  for (const f of newFailures) console.error(`  new: ${f}`)
  console.error('\nIf this is a real regression, fix it. If it is an intentional behavior change,')
  console.error('update the test in the same commit — do not add it to the baseline to silence it.')
  process.exit(1)
}

console.log(`\n✓ No new UI test failures (${Object.values(baseline).flat().length} pre-existing tracked in baseline, unchanged or improved).`)
process.exit(0)
