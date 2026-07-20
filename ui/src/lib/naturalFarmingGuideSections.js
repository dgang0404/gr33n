/**
 * Phase 209 WS3 — extract instructional sections from natural-farming field guides.
 */

/** Cards shown during Make a batch (matches WS2 guide headings). */
export const BATCH_STEP_SECTION_PREFIXES = [
  { key: 'ingredients', prefix: 'Ingredients' },
  { key: 'steps', prefix: 'Step-by-step preparation' },
  { key: 'timeline', prefix: 'Ferment / wait timeline' },
  { key: 'ready', prefix: 'Ready signs' },
  { key: 'safety', prefix: 'Safety & water' },
]

export function stripFrontmatter(md) {
  const text = String(md || '')
  if (!text.startsWith('---\n')) return text
  const end = text.indexOf('\n---\n', 3)
  return end >= 0 ? text.slice(end + 5) : text
}

/**
 * @param {string} bodyMd
 * @returns {Record<string, string>}
 */
export function extractGuideSections(bodyMd) {
  const body = stripFrontmatter(bodyMd)
  /** @type {Record<string, string>} */
  const sections = {}
  let current = null
  /** @type {string[]} */
  let buf = []
  for (const line of body.split('\n')) {
    const m = line.match(/^## (.+)$/)
    if (m) {
      if (current) sections[current] = buf.join('\n').trim()
      current = m[1].trim()
      buf = []
    } else if (current) {
      buf.push(line)
    }
  }
  if (current) sections[current] = buf.join('\n').trim()
  return sections
}

/**
 * @param {Record<string, string>} sections
 * @param {string} prefix
 */
export function sectionBodyByPrefix(sections, prefix) {
  if (!sections || !prefix) return ''
  const key = Object.keys(sections).find((k) => k.startsWith(prefix))
  return key ? sections[key] : ''
}

/**
 * @param {string} bodyMd
 */
export function batchStepCards(bodyMd) {
  const sections = extractGuideSections(bodyMd)
  return BATCH_STEP_SECTION_PREFIXES.map(({ key, prefix }) => ({
    key,
    title: prefix,
    body: sectionBodyByPrefix(sections, prefix),
  })).filter((c) => c.body)
}
