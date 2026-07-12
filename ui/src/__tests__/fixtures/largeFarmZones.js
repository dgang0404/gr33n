/**
 * Phase 173 WS6 — synthetic large-farm fixture for filter/paging tests.
 * 24 zones: mix of indoor/outdoor/greenhouse, 3 flagged for attention.
 */

const TYPES = ['indoor', 'outdoor', 'greenhouse']

export function buildLargeFarmZones(count = 24) {
  return Array.from({ length: count }, (_, i) => ({
    id: i + 1,
    name: `Zone ${i + 1}`,
    zone_type: TYPES[i % TYPES.length],
  }))
}

/**
 * Status resolver for tests: zones at index 0, 5, 10 (1-indexed ids 1, 6, 11)
 * are flagged as needing attention; everything else is healthy.
 */
export function largeFarmStatusFor(zone) {
  const attentionIds = new Set([1, 6, 11])
  if (attentionIds.has(zone.id)) {
    return { health: 'warn', attention: [{ label: 'Needs a look' }], plants: { state: 'growing' } }
  }
  return { health: 'ok', plants: { state: 'growing' } }
}

export const LARGE_FARM_ZONES = buildLargeFarmZones()
