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
