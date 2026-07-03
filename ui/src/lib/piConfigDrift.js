/**
 * Phase 121 — compare Pi-reported wiring hash vs platform-generated config.
 */

/** @param {object|null|undefined} device */
export function reportedConfigSha256(device) {
  const cfg = device?.config
  if (!cfg || typeof cfg !== 'object') return ''
  return String(cfg.config_sha256 || '').trim().toLowerCase()
}

/**
 * @param {object|null|undefined} device
 * @param {string} expectedSha256 from GET /devices/{id}/pi-config
 * @returns {'unknown'|'synced'|'stale'}
 */
export function wiringDriftStatus(device, expectedSha256) {
  const expected = String(expectedSha256 || '').trim().toLowerCase()
  const reported = reportedConfigSha256(device)
  if (!expected) return 'unknown'
  if (!reported) return 'unknown'
  return reported === expected ? 'synced' : 'stale'
}

/** @param {'unknown'|'synced'|'stale'} status */
export function wiringDriftLabel(status) {
  switch (status) {
    case 'synced':
      return 'Pi wiring matches platform'
    case 'stale':
      return 'Pi running stale wiring — update config on the Pi'
    default:
      return 'Wiring drift unknown — Pi has not reported a config hash yet'
  }
}
