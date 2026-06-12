/**
 * Phase 104 — group crop cycles by catalog crop_key for compare / money views.
 */

/**
 * @param {object[]} cycles
 * @returns {{ key: string, label: string, cycles: object[] }[]}
 */
export function groupCyclesByCropKey(cycles) {
  const groups = new Map()
  const unlinked = []
  for (const c of cycles || []) {
    const key = String(c?.crop_key || '').trim()
    if (!key) {
      unlinked.push(c)
      continue
    }
    if (!groups.has(key)) {
      const label = String(c.catalog_display_name || key).trim() || key
      groups.set(key, { key, label, cycles: [] })
    }
    groups.get(key).cycles.push(c)
  }
  const out = [...groups.values()].sort((a, b) => a.label.localeCompare(b.label))
  for (const g of out) {
    g.cycles.sort((a, b) => Number(b.id) - Number(a.id))
  }
  if (unlinked.length) {
    out.push({
      key: '',
      label: 'No catalog crop',
      cycles: unlinked.sort((a, b) => Number(b.id) - Number(a.id)),
    })
  }
  return out
}

/**
 * @param {object} cycle
 */
export function cyclePickerLabel(cycle) {
  const crop = cycle?.catalog_display_name || cycle?.crop_key
  const batch = cycle?.batch_label
  const parts = [cycle?.name || `Grow #${cycle?.id}`]
  if (crop) parts.push(String(crop))
  if (batch) parts.push(String(batch))
  return parts.filter(Boolean).join(' · ')
}
