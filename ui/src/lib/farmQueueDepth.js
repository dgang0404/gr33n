/**
 * Phase 41 WS1 — sum pending device commands across farm devices.
 */

import api from '../api'

/**
 * @param {object[]} devices
 * @returns {Promise<number>}
 */
export async function sumFarmPendingQueueDepth(devices) {
  const ids = (devices || [])
    .map((d) => d.id)
    .filter((id) => id != null)

  if (!ids.length) return 0

  const counts = await Promise.all(
    ids.map(async (deviceId) => {
      try {
        const r = await api.get(`/devices/${deviceId}/commands`, {
          params: { status: 'pending' },
        })
        const list = r.data?.commands ?? r.data ?? []
        return Array.isArray(list) ? list.length : 0
      } catch {
        return 0
      }
    }),
  )

  return counts.reduce((a, b) => a + b, 0)
}
