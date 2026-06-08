/** Phase 51 WS4 — Pi platform config sync staleness on device cards. */

export const CONFIG_STALE_MINUTES_DEFAULT = 10
export const CONFIG_FRESH_POLL_MULTIPLIER = 2

/** Decode device.config whether API returns object or base64 JSON string. */
export function deviceConfigObject(device) {
  const raw = device?.config
  if (!raw) return {}
  if (typeof raw === 'object' && !Array.isArray(raw)) return raw
  if (typeof raw === 'string') {
    try {
      const text = atob(raw)
      return JSON.parse(text)
    } catch {
      try {
        return JSON.parse(raw)
      } catch {
        return {}
      }
    }
  }
  return {}
}

export function deviceLastConfigFetchAt(device) {
  const cfg = deviceConfigObject(device)
  const ts = cfg?.last_config_fetch_at
  return typeof ts === 'string' && ts.trim() ? ts.trim() : null
}

/** Whether the device is set up for platform config sync (has uid + version). */
export function deviceUsesPlatformSync(device) {
  const uid = (device?.device_uid || '').trim()
  if (!uid) return false
  const version = device?.config_version
  return typeof version === 'number' && version > 0
}

/**
 * @returns {{ tone: 'ok'|'warn'|'muted', label: string } | null}
 * null when badge should not show (local-YAML Pi or non-sync device).
 */
export function configSyncBadge(device, opts = {}) {
  if (!deviceUsesPlatformSync(device)) return null
  const staleMinutes = opts.staleMinutes ?? CONFIG_STALE_MINUTES_DEFAULT
  const pollSeconds = opts.pollIntervalSeconds ?? 30
  const fetchedAt = deviceLastConfigFetchAt(device)
  if (!fetchedAt) {
    return { tone: 'muted', label: 'Never fetched' }
  }
  const ageMs = Date.now() - new Date(fetchedAt).getTime()
  if (Number.isNaN(ageMs) || ageMs < 0) {
    return { tone: 'muted', label: 'Never fetched' }
  }
  const freshMs = pollSeconds * CONFIG_FRESH_POLL_MULTIPLIER * 1000
  const staleMs = staleMinutes * 60 * 1000
  if (ageMs <= freshMs) {
    return { tone: 'ok', label: 'Config synced' }
  }
  if (ageMs > staleMs) {
    return { tone: 'warn', label: 'Config stale' }
  }
  return { tone: 'ok', label: 'Config synced' }
}
