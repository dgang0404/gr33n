/** Phase 50 — Pi wiring metadata helpers (sensors / actuators). */

export const SENSOR_WIRING_SOURCES = [
  { value: 'dht22', label: 'DHT22 (temp / humidity)' },
  { value: 'ads1115', label: 'ADS1115 (analog)' },
  { value: 'mhz19', label: 'MH-Z19 (CO₂ serial)' },
  { value: 'bh1750', label: 'BH1750 (light I2C)' },
  { value: 'gpio_digital', label: 'GPIO digital' },
  { value: 'derived', label: 'Derived (computed)' },
]

/** Prefer API top-level wiring; fall back to config.wiring. */
export function resolveWiring(entity) {
  if (!entity) return null
  if (entity.wiring) return entity.wiring
  let cfg = entity.config
  if (!cfg) return null
  if (typeof cfg === 'string') {
    try { cfg = JSON.parse(cfg) } catch { return null }
  }
  return cfg?.wiring ?? null
}

/** Farmer-facing one-line wiring summary. */
export function formatWiringLabel(wiring) {
  if (!wiring) return ''
  const parts = []
  if (wiring.source) parts.push(String(wiring.source).toUpperCase())
  if (wiring.gpio_pin != null) parts.push(`BCM GPIO ${wiring.gpio_pin}`)
  if (wiring.i2c_channel != null) parts.push(`I2C ch ${wiring.i2c_channel}`)
  if (wiring.serial_port) parts.push(wiring.serial_port)
  return parts.join(' · ')
}

export function wiringIsEmpty(wiring) {
  if (!wiring) return true
  return !wiring.source
    && wiring.gpio_pin == null
    && wiring.i2c_channel == null
    && !wiring.serial_port
    && wiring.device_id == null
}

/** One-line GPIO / relay label for a sensor or actuator entity. */
export function formatEntityHardwareLabel(entity) {
  if (!entity) return ''
  const fromWiring = formatWiringLabel(resolveWiring(entity))
  if (fromWiring) return fromWiring
  const hi = entity.hardware_identifier
  if (hi != null && hi !== '') return `Relay ch ${hi}`
  return ''
}

function intVal(v) {
  if (v == null || v === '') return null
  const n = Number(v)
  return Number.isFinite(n) ? n : null
}

function sharedDht22Gpio(a, b) {
  return a === 'dht22' && b === 'dht22'
}

function entityWiring(entity) {
  return resolveWiring(entity)
}

function sharesGpio(wiring, deviceId, pin) {
  if (!wiring || wiring.device_id == null || wiring.gpio_pin == null) return false
  return Number(wiring.device_id) === deviceId && Number(wiring.gpio_pin) === pin
}

function sharesI2c(wiring, deviceId, channel) {
  if (!wiring || wiring.device_id == null || wiring.i2c_channel == null) return false
  return Number(wiring.device_id) === deviceId && Number(wiring.i2c_channel) === channel
}

/** Client-side pin/channel conflict preview (mirrors API checks). */
export function findWiringConflict({ wiring, entityType, entityId, sensors = [], actuators = [] }) {
  if (!wiring || wiring.device_id == null) return null
  const deviceId = Number(wiring.device_id)
  const pin = intVal(wiring.gpio_pin)
  const channel = intVal(wiring.i2c_channel)
  const source = wiring.source || ''

  if (pin != null) {
    for (const s of sensors) {
      if (entityType === 'sensor' && Number(s.id) === Number(entityId)) continue
      const w = entityWiring(s)
      if (!sharesGpio(w, deviceId, pin)) continue
      if (sharedDht22Gpio(source, w?.source)) continue
      return {
        entity_type: 'sensor',
        entity_id: s.id,
        entity_name: s.name,
        message: `sensor ${s.id} (${s.name}) already uses this pin/channel on the device`,
      }
    }
    for (const a of actuators) {
      if (entityType === 'actuator' && Number(a.id) === Number(entityId)) continue
      const w = entityWiring(a)
      if (sharesGpio(w, deviceId, pin)) {
        return {
          entity_type: 'actuator',
          entity_id: a.id,
          entity_name: a.name,
          message: `actuator ${a.id} (${a.name}) already uses this pin/channel on the device`,
        }
      }
    }
  }

  if (channel != null && entityType === 'sensor') {
    for (const s of sensors) {
      if (Number(s.id) === Number(entityId)) continue
      const w = entityWiring(s)
      if (sharesI2c(w, deviceId, channel)) {
        return {
          entity_type: 'sensor',
          entity_id: s.id,
          entity_name: s.name,
          message: `sensor ${s.id} (${s.name}) already uses this pin/channel on the device`,
        }
      }
    }
  }
  return null
}
