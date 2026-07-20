/**
 * Phase 209 WS3 — Make a batch flow (canon inputs + field guide sections).
 */
import api from '../api'
import { sectionBodyByPrefix, extractGuideSections } from './naturalFarmingGuideSections.js'

/** Mirrors data/natural_farming_process_vocabulary.yaml process_types (read API has no vocab endpoint). */
export const PROCESS_TYPE_LABELS = {
  jms: { label: 'JMS', expand: 'JADAM Microbial Solution', tradition: 'jadam' },
  jlf: { label: 'JLF', expand: 'JADAM Liquid Fertilizer', tradition: 'jadam' },
  fpj: { label: 'FPJ', expand: 'Fermented Plant Juice', tradition: 'knf' },
  ffj: { label: 'FFJ', expand: 'Fermented Fruit Juice', tradition: 'knf' },
  lab: { label: 'LAB', expand: 'Lactic Acid Bacteria serum', tradition: 'knf' },
  ohn: { label: 'OHN', expand: 'Oriental Herbal Nutrient', tradition: 'knf' },
  faa: { label: 'FAA', expand: 'Fish Amino Acid', tradition: 'knf' },
  jwa: { label: 'JWA', expand: 'JADAM Wetting Agent', tradition: 'jadam' },
  js: { label: 'JS', expand: 'JADAM Sulfur concentrate', tradition: 'jadam' },
  jhs: { label: 'JHS', expand: 'JADAM Herbal Solution', tradition: 'jadam' },
  wca: { label: 'WCA', expand: 'Water-Soluble Calcium', tradition: 'knf' },
  wcs: { label: 'WCS', expand: 'Water-Soluble Calcium Phosphate', tradition: 'knf' },
  brv: { label: 'BRV', expand: 'Brown Rice Vinegar', tradition: 'other' },
  compost_tea_aact: { label: 'Compost tea', expand: 'Actively aerated compost tea', tradition: 'other' },
}

/**
 * @param {Record<string, unknown>} canon
 */
export function processTypesFromCanon(canon) {
  const inputs = /** @type {Array<Record<string, unknown>>} */ (canon?.inputs ?? [])
  const seen = new Map()
  for (const inp of inputs) {
    const id = String(inp.process_type || '').trim()
    if (!id || seen.has(id)) continue
    const meta = PROCESS_TYPE_LABELS[id] ?? { label: id.toUpperCase(), expand: id, tradition: inp.tradition }
    seen.set(id, {
      id,
      label: meta.label,
      expand: meta.expand,
      tradition: inp.tradition ?? meta.tradition,
    })
  }
  return [...seen.values()].sort((a, b) => a.label.localeCompare(b.label))
}

/**
 * @param {string} processType
 * @param {Record<string, unknown>} canon
 */
export function variantsForProcess(processType, canon) {
  return (canon?.inputs ?? []).filter((i) => i.process_type === processType)
}

/**
 * @param {string} guideFile
 */
export function guideSlug(guideFile) {
  return String(guideFile || '').replace(/\.md$/, '').trim()
}

/**
 * @param {string} guideFile
 */
export async function loadFieldGuideBody(guideFile) {
  const slug = guideSlug(guideFile)
  if (!slug) return ''
  const { data } = await api.get(`/commons/agronomy-field-guides/${encodeURIComponent(slug)}`)
  return data?.body_md ?? ''
}

/**
 * @param {Record<string, unknown>} canonInput
 */
export function canonDilutionHint(canonInput) {
  const parts = []
  if (canonInput?.dilution_start) parts.push(`Start ${canonInput.dilution_start}`)
  if (canonInput?.dilution_strong && canonInput.dilution_strong !== canonInput.dilution_start) {
    parts.push(`stronger ${canonInput.dilution_strong}`)
  }
  return parts.join(' · ')
}

/**
 * @param {Record<string, unknown>} canonInput
 * @param {string} bodyMd
 */
export function buildInputPayload(canonInput, bodyMd) {
  const sections = extractGuideSections(bodyMd)
  const clip = (s, n) => (s ? String(s).slice(0, n) : '')
  return {
    name: canonInput.seed_name,
    category: canonInput.schema_category,
    description: clip(sectionBodyByPrefix(sections, 'What it is'), 500),
    typical_ingredients: clip(sectionBodyByPrefix(sections, 'Ingredients'), 500),
    preparation_summary: clip(sectionBodyByPrefix(sections, 'Step-by-step preparation'), 500),
    storage_guidelines: clip(sectionBodyByPrefix(sections, 'Storage'), 300),
    safety_precautions: clip(sectionBodyByPrefix(sections, 'Safety'), 300),
    reference_source: canonInput.reference_source || '',
  }
}

/**
 * @param {Array<Record<string, unknown>>} farmInputs
 * @param {string} seedName
 */
export function findFarmInputByName(farmInputs, seedName) {
  return farmInputs.find((i) => i.name === seedName) ?? null
}

/**
 * @param {Record<string, unknown>} canonInput
 * @param {string} bodyMd
 */
export function batchCreatePayload(canonInput, bodyMd, form) {
  const sections = extractGuideSections(bodyMd)
  const today = new Date().toISOString().slice(0, 10)
  return {
    batch_identifier: form.batch_identifier?.trim() || undefined,
    status: form.status || 'fermenting_brewing',
    creation_start_date: today,
    ingredients_used: sectionBodyByPrefix(sections, 'Ingredients')?.slice(0, 1000) || undefined,
    procedure_followed: `field-guides/${guideSlug(canonInput.guide)}.md`,
    observations_notes: form.observations_notes?.trim() || undefined,
    quantity_produced: form.quantity_produced != null ? Number(form.quantity_produced) : undefined,
    current_quantity_remaining:
      form.current_quantity_remaining != null ? Number(form.current_quantity_remaining) : 0,
  }
}

/**
 * @param {Record<string, unknown>} canonInput
 * @param {string} bodyMd
 */
export function prepTaskPayload(canonInput, bodyMd) {
  const sections = extractGuideSections(bodyMd)
  const timeline = sectionBodyByPrefix(sections, 'Ferment / wait timeline')
  return {
    title: `Prepare ${canonInput.seed_name}`,
    description: [timeline, sectionBodyByPrefix(sections, 'Ready signs')].filter(Boolean).join('\n\n').slice(0, 800),
    task_type: 'jadam_prep',
    status: 'todo',
    priority: 2,
    due_date: new Date().toISOString().slice(0, 10),
  }
}
