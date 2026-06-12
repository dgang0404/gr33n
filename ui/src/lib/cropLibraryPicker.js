/**
 * Phase 82 WS4f — crop library picker helpers (search + grouping).
 */

export const CATEGORY_ORDER = [
  'fruiting',
  'leafy',
  'herb',
  'flower',
  'epiphyte',
  'industrial',
  'ornamental',
  'custom',
]

/** @param {string} q */
export function normalizeSearchQuery(q) {
  return String(q || '').trim().toLowerCase()
}

/**
 * @param {object} item
 */
export function itemMatchesQuery(item, q) {
  if (!q) return true
  const terms = item.search_terms || []
  if (terms.some((t) => t.includes(q))) return true
  return (
    String(item.display_name || '').toLowerCase().includes(q) ||
    String(item.crop_key || '').toLowerCase().includes(q)
  )
}

/**
 * @param {{ groups?: Array<{ key: string, label: string, items: Array<object> }> }} picker
 * @param {string} query
 */
export function filterPickerGroups(picker, query) {
  const q = normalizeSearchQuery(query)
  if (!picker?.groups?.length) return []
  if (!q) return picker.groups
  return picker.groups
    .map((g) => ({
      ...g,
      items: (g.items || []).filter((item) => itemMatchesQuery(item, q)),
    }))
    .filter((g) => g.items.length > 0)
}

/**
 * @param {Array<{ groups?: Array<{ items: Array<object> }> }>|{ groups?: Array<{ items: Array<object> }> }} picker
 * @param {number|null|undefined} profileId
 */
export function findPickerItemByProfileId(picker, profileId) {
  if (!profileId) return null
  const groups = Array.isArray(picker) ? picker : picker?.groups
  if (!groups) return null
  for (const g of groups) {
    for (const item of g.items || []) {
      if (Number(item.crop_profile_id) === Number(profileId)) return item
    }
  }
  return null
}

/**
 * @param {object|null} item
 */
export function pickerItemLabel(item) {
  if (!item) return ''
  let s = item.display_name || item.crop_key || ''
  if (item.is_custom) s += ' (custom)'
  return s
}

/**
 * @param {object|null} item
 */
export function pickerItemHint(item) {
  if (!item) return ''
  if (item.has_targets && item.substrate) {
    return `${item.substrate}${item.watering_style ? ` · ${item.watering_style.replace(/_/g, ' ')}` : ''}`
  }
  if (!item.has_targets && item.cousin_label) {
    return `No built-in targets yet — start from ${item.cousin_label}`
  }
  if (!item.has_targets) return 'Catalog entry — not in knowledge base yet'
  return ''
}

/** Fallback when GET …/crop-library/picker is missing (older API). */
export function buildPickerFallbackFromProfiles(profiles) {
  const builtins = (profiles || []).filter((p) => p.is_builtin)
  const items = builtins.map((p) => ({
    crop_key: p.crop_key,
    display_name: p.display_name,
    category: p.category || 'fruiting',
    crop_profile_id: p.id,
    has_targets: true,
    is_custom: false,
    search_terms: [p.crop_key, p.display_name].filter(Boolean).map((s) => String(s).toLowerCase()),
  }))
  items.sort((a, b) => a.display_name.localeCompare(b.display_name))
  return {
    version: 1,
    counts: { with_targets: items.length, catalog_only: 0, total: items.length },
    groups: [{ key: 'fruiting', label: 'Crops (knowledge base)', items }],
  }
}

function num(v) {
  if (v == null || v === '') return null
  const n = Number(v)
  return Number.isFinite(n) ? n : null
}

/** @param {object} stage crop_profile_stages row */
export function formatStageTargetLine(stage) {
  if (!stage) return ''
  const parts = []
  const ecMin = num(stage.ec_min)
  const ecMax = num(stage.ec_max)
  const ecT = num(stage.ec_target)
  if (ecMin != null || ecMax != null) {
    const range = ecMin != null && ecMax != null ? `${ecMin}–${ecMax}` : String(ecMin ?? ecMax)
    parts.push(`EC ${range} mS/cm${ecT != null ? ` (t ${ecT})` : ''}`)
  }
  const dli = num(stage.dli_target)
  if (dli != null) parts.push(`DLI ${dli} mol/m²/d`)
  const photo = num(stage.photoperiod_hrs)
  if (photo != null) parts.push(`${photo}h photoperiod`)
  const label = String(stage.stage || '').replace(/_/g, ' ')
  return parts.length ? `${label}: ${parts.join(' · ')}` : label
}
