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
  if (!item.has_targets) return 'Catalog entry — clone a profile to customize'
  return ''
}
