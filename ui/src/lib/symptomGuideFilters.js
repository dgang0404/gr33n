/** Distinct crop keys and categories from agronomy symptom catalog rows. */
export function uniqueCropKeys(symptoms) {
  const out = new Set()
  for (const row of symptoms || []) {
    for (const k of row.crop_keys || []) {
      const key = String(k || '').trim()
      if (key) out.add(key)
    }
  }
  return [...out].sort((a, b) => a.localeCompare(b))
}

export function uniqueCategories(symptoms) {
  const out = new Set()
  for (const row of symptoms || []) {
    for (const c of row.categories || []) {
      const cat = String(c || '').trim()
      if (cat) out.add(cat)
    }
  }
  return [...out].sort((a, b) => a.localeCompare(b))
}

export function formatCropKeys(cropKeys) {
  const keys = (cropKeys || []).map((k) => String(k).trim()).filter(Boolean)
  return keys.length ? keys.join(', ') : 'All crops'
}
