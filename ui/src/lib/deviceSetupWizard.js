/**
 * Phase 44 WS3 — edge device wizard helpers.
 */

export const DEVICE_TYPE_OPTIONS = [
  { value: 'raspberry_pi_edge', label: 'Raspberry Pi edge client' },
  { value: 'relay_controller', label: 'Relay controller' },
  { value: 'mqtt_bridge', label: 'MQTT bridge' },
]

/** Suggested quick-add actuators after device registration. */
export const DEVICE_ACTUATOR_TEMPLATES = [
  { id: 'grow_light', actuator_type: 'grow_light', label: 'Grow light', defaultName: 'Grow light' },
  { id: 'pump', actuator_type: 'pump', label: 'Irrigation pump', defaultName: 'Feed pump' },
  { id: 'exhaust_fan', actuator_type: 'exhaust_fan', label: 'Exhaust fan', defaultName: 'Exhaust fan' },
]

/** Short field checklist from pi-integration-guide §8.3 (in-app, not PDF-only). */
export const PI_FIELD_CHECKLIST = [
  { id: 'api', label: 'API reachable on LAN; per-device key issued in wizard (or legacy PI_API_KEY on server)' },
  { id: 'pi-os', label: 'Pi OS 64-bit, SSH, NTP/chrony (UTC timestamps)' },
  { id: 'deps', label: 'Edge deps installed — ./scripts/install-pi-edge-deps.sh on the Pi' },
  { id: 'device-key', label: 'GR33N_DEVICE_API_KEY or /etc/gr33n/device.key on the Pi (preferred over shared api_key)' },
  { id: 'config', label: 'pi_client/config.yaml — download from wizard; device.uid + farm_id match dashboard' },
  { id: 'systemd', label: 'systemd gr33n service enabled — journalctl -u gr33n -f' },
  { id: 'readings', label: 'Dashboard Live Sensors update; device shows online after heartbeat' },
  { id: 'relay-test', label: 'One-relay safe bench test (LED) before mains loads' },
]

export function deviceSetupRoute(farmId, zoneId = null) {
  const base = `/farms/${farmId}/devices/new`
  if (zoneId == null) return base
  return `${base}?zone_id=${zoneId}`
}

/**
 * @param {number} farmId
 */
export function suggestDeviceUid(farmId) {
  const stamp = Date.now().toString(36)
  return `pi-farm${farmId}-${stamp}`
}

/**
 * @param {object} form
 */
export function buildDeviceCreatePayload(form) {
  const name = String(form.name || '').trim()
  const deviceUid = String(form.deviceUid || '').trim()
  if (!name) throw new Error('Device name is required')
  if (!deviceUid) throw new Error('Device UID is required')
  return {
    name,
    device_uid: deviceUid,
    device_type: form.deviceType || 'raspberry_pi_edge',
    zone_id: form.zoneId || null,
    status: 'offline',
  }
}

/**
 * @param {object} params
 */
export function buildActuatorCreatePayload({
  farmId,
  deviceId,
  zoneId,
  template,
  customName = '',
  hardwareId = 'BCM17',
}) {
  const name = String(customName || template.defaultName || template.label).trim()
  return {
    farmId,
    body: {
      name,
      actuator_type: template.actuator_type,
      device_id: deviceId,
      zone_id: zoneId || null,
      hardware_identifier: hardwareId || null,
    },
  }
}

/**
 * @param {object|null} device
 */
export function isDeviceOnline(device) {
  if (!device) return false
  return String(device.status || '').toLowerCase() === 'online'
}

/**
 * Build pi_client config snippet for copy-paste.
 */
export function buildPiConfigSnippet({ baseUrl, farmId, deviceId, deviceUid }) {
  const lines = [
    '# pi_client/config.yaml (excerpt)',
    'api:',
    `  base_url: ${baseUrl || 'http://<api-lan-ip>:8080'}`,
    '  api_key: <optional legacy PI_API_KEY — prefer GR33N_DEVICE_API_KEY on Pi>',
    'farm:',
    `  farm_id: ${farmId}`,
    'device:',
    `  device_id: ${deviceId}`,
    `  device_uid: ${deviceUid}`,
  ]
  return lines.join('\n')
}

export function formatDeviceStatusLabel(device) {
  if (!device) return 'Unknown'
  const status = String(device.status || 'offline').replace(/_/g, ' ')
  if (device.last_heartbeat) {
    return `${status} — last heartbeat ${new Date(device.last_heartbeat).toLocaleString()}`
  }
  return status
}
