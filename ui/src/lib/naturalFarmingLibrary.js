/**
 * Phase 209 WS3b — recipe library helpers.
 */

export const LIBRARY_TABS = [
  { id: 'inputs', label: 'Inputs' },
  { id: 'application', label: 'Application' },
  { id: 'programs', label: 'Programs' },
]

/** Bootstrap program explainers (field guide backed). */
export const LIBRARY_PROGRAMS = [
  {
    id: 'jadam_indoor_photoperiod_v1',
    title: 'Indoor photoperiod starter',
    guide: 'natural-farming-indoor-photoperiod-program.md',
    bootstrapTemplate: 'jadam_indoor_photoperiod_v1',
    summary: 'Veg 18/6, flower 12/12, and outdoor JLF programs from bootstrap template.',
  },
]

/** Phase 211 WS3 — livestock feed templates (Animals module). */
export const LIVESTOCK_FEED_TEMPLATES = [
  {
    id: 'livestock_comfrey_feed_v1',
    title: 'Comfrey & sprouted grain supplements',
    guide: 'natural-farming-livestock-plant-feed.md',
    packKey: 'livestock_comfrey_feed_v1',
    summary: 'animal_feed inputs with simple flock supplement examples — not ration math.',
    moduleSchema: 'gr33nanimals',
  },
]

/**
 * @param {string} tradition
 */
export function traditionBadge(tradition) {
  const t = String(tradition || '').trim()
  if (t === 'knf') return { text: 'KNF', class: 'text-amber-400/90' }
  if (t === 'jadam') return { text: 'JADAM', class: 'text-green-400/90' }
  if (t === 'extension') return { text: 'Extension', class: 'text-sky-400/90' }
  if (t === 'other') return { text: 'Other', class: 'text-zinc-400' }
  return t ? { text: t, class: 'text-zinc-400' } : null
}

/**
 * @param {string} targetType
 */
export function formatApplicationType(targetType) {
  return String(targetType || '')
    .replace(/_/g, ' ')
    .replace(/\b\w/g, (c) => c.toUpperCase())
}

/**
 * @param {string[]} componentNames
 * @param {Array<{ name: string, id: number }>} farmInputs
 * @param {Array<{ input_definition_id: number, status: string, batch_identifier?: string, id: number }>} farmBatches
 */
export function readyBatchesForComponents(componentNames, farmInputs, farmBatches) {
  const ready = new Set(['ready_for_use', 'partially_used'])
  const out = []
  for (const name of componentNames || []) {
    const def = farmInputs.find((i) => i.name === name)
    if (!def) continue
    const batches = farmBatches.filter(
      (b) => b.input_definition_id === def.id && ready.has(String(b.status)),
    )
    for (const b of batches) {
      out.push({ inputName: name, batch: b })
    }
  }
  return out
}

/**
 * @param {string} seedName
 */
export function libraryCardSlug(seedName) {
  return String(seedName)
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-|-$/g, '')
    .slice(0, 48)
}
