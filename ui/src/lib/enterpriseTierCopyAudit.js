/**
 * Phase 59 — banned ERP jargon on farmer-facing Vue surfaces.
 */
import { readdirSync, readFileSync } from 'node:fs'
import { join } from 'node:path'

/** Regexes that must not appear in farmer UI copy (case-insensitive). */
export const BANNED_FARMER_COPY = [
  /purchase order/i,
  /\bMETRC\b/i,
  /general ledger/i,
  /SKU master/i,
  /\bwarehouse\b/i,
]

/** Advanced / power-user views — GL export and account codes allowed. */
export const COPY_AUDIT_EXCLUDE = new Set(['Costs.vue'])

/**
 * @param {string} uiSrcRoot absolute or cwd-relative ui/src path
 * @returns {{ file: string, term: string }[]}
 */
export function findBannedCopyInFarmerVue(uiSrcRoot) {
  const viewsDir = join(uiSrcRoot, 'views')
  const hits = []
  for (const name of readdirSync(viewsDir)) {
    if (!name.endsWith('.vue') || COPY_AUDIT_EXCLUDE.has(name)) continue
    const text = readFileSync(join(viewsDir, name), 'utf8')
    for (const pattern of BANNED_FARMER_COPY) {
      if (pattern.test(text)) {
        hits.push({ file: `views/${name}`, term: pattern.source })
      }
    }
  }
  const libFiles = ['zoneSetupWizard.js']
  for (const name of libFiles) {
    const text = readFileSync(join(uiSrcRoot, 'lib', name), 'utf8')
    for (const pattern of BANNED_FARMER_COPY) {
      if (pattern.test(text)) {
        hits.push({ file: `lib/${name}`, term: pattern.source })
      }
    }
  }
  return hits
}
