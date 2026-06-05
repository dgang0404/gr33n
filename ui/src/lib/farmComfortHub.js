/**
 * Phase 42 WS1 — farm-wide comfort targets hub cards.
 */

import {
  buildZoneComfortBands,
  summarizeZoneComfortStatus,
  COMFORT_STATUS_META,
} from './comfortBand.js'

/**
 * @param {object} params
 * @returns {Array<{ zone: object, bands: object[], status: string, statusMeta: object, summaryLine: string }>}
 */
export function buildFarmComfortCards({
  zones = [],
  sensors = [],
  setpoints = [],
  readings = {},
}) {
  return (zones || []).map((zone) => {
    const bands = buildZoneComfortBands({
      zoneId: zone.id,
      sensors,
      setpoints,
      readings,
    })
    const status = summarizeZoneComfortStatus(bands)
    return {
      zone,
      bands,
      status,
      statusMeta: COMFORT_STATUS_META[status] || COMFORT_STATUS_META.no_sensors,
      summaryLine: formatBandSummaryLine(bands),
    }
  })
}

/**
 * @param {Array} bands
 */
function formatBandSummaryLine(bands) {
  if (!bands.length) return 'Add climate sensors to set comfort bands'
  return bands
    .map((b) => `${b.label}: ${COMFORT_STATUS_META[b.status]?.label || b.status}`)
    .join(' · ')
}

/**
 * @param {Array} cards
 * @param {number|null} zoneId
 */
export function filterComfortCardsByZone(cards, zoneId) {
  if (zoneId == null) return cards
  return (cards || []).filter((c) => Number(c.zone.id) === Number(zoneId))
}

/**
 * @param {object} params
 * @returns {number}
 */
export function countMissingComfortBands({ zones = [], sensors = [], setpoints = [], readings = {} }) {
  return buildFarmComfortCards({ zones, sensors, setpoints, readings })
    .reduce((n, card) => n + card.bands.filter((b) => b.status === 'missing').length, 0)
}
