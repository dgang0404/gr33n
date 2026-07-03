/**
 * Phase 121 — per-driver physical hookup steps (mirrors API driver_hookups).
 */

/** @typedef {{ wire: string, to: string, role: string }} HookupStep */

/** @type {Record<string, HookupStep[]>} */
export const FALLBACK_DRIVER_HOOKUPS = {
  dht22: [
    { wire: 'VCC', to: 'Physical pin 1 or 17 (3.3 V)', role: 'power3v3' },
    { wire: 'DATA', to: 'Your chosen GPIO (BCM)', role: 'gpio' },
    { wire: 'GND', to: 'Any GND pin (e.g. pin 6)', role: 'gnd' },
  ],
  ads1115: [
    { wire: 'VDD', to: 'Physical pin 1 or 17 (3.3 V)', role: 'power3v3' },
    { wire: 'GND', to: 'Any GND pin', role: 'gnd' },
    { wire: 'SDA', to: 'Physical pin 3 (I²C SDA)', role: 'i2c_sda' },
    { wire: 'SCL', to: 'Physical pin 5 (I²C SCL)', role: 'i2c_scl' },
    { wire: 'A0–A3', to: 'Analog sensor signal (per channel)', role: 'analog_in' },
  ],
  bh1750: [
    { wire: 'VCC', to: 'Physical pin 1 or 17 (3.3 V)', role: 'power3v3' },
    { wire: 'GND', to: 'Any GND pin', role: 'gnd' },
    { wire: 'SDA', to: 'Physical pin 3 (I²C SDA)', role: 'i2c_sda' },
    { wire: 'SCL', to: 'Physical pin 5 (I²C SCL)', role: 'i2c_scl' },
  ],
  mhz19: [
    { wire: 'VIN', to: 'Physical pin 2 or 4 (5 V)', role: 'power5v' },
    { wire: 'GND', to: 'Any GND pin', role: 'gnd' },
    { wire: 'TX', to: 'Physical pin 10 (Pi RX / GPIO15)', role: 'uart_rx' },
    { wire: 'RX', to: 'Physical pin 8 (Pi TX / GPIO14)', role: 'uart_tx' },
  ],
  gpio_digital: [
    { wire: 'Signal', to: 'Your chosen GPIO (BCM)', role: 'gpio' },
    { wire: 'VCC', to: 'Sensor supply (3.3 V or 5 V per datasheet)', role: 'power3v3' },
    { wire: 'GND', to: 'Any GND pin', role: 'gnd' },
  ],
  gpio_relay: [
    { wire: 'IN / coil', to: 'GPIO pin or relay HAT channel output', role: 'gpio' },
    { wire: 'COM / NO', to: 'Load wiring (mains or low-voltage per relay)', role: 'load' },
  ],
  relay_hat: [
    { wire: 'HAT stack', to: '40-pin header — seats on pins 1–40', role: 'hat' },
    { wire: 'I²C', to: 'Uses pins 3 (SDA) and 5 (SCL) on the bus', role: 'i2c_sda' },
    { wire: 'DIP switches', to: 'Set ID0–ID2 per stack level (see relay stack view)', role: 'dip' },
    { wire: 'Channel', to: 'Assign relay channel 0–63 in wiring panel', role: 'relay_channel' },
  ],
}

/** @param {object|null|undefined} taxonomy */
export function driverHookupsFromTaxonomy(taxonomy) {
  return taxonomy?.driver_hookups || FALLBACK_DRIVER_HOOKUPS
}

/** @param {Record<string, HookupStep[]>} hookups @param {string} driver */
export function hookupStepsForDriver(hookups, driver) {
  const key = String(driver || '').toLowerCase()
  return hookups[key] || []
}

/** Resolve wiring source for an entity assignment. */
export function wiringSourceForEntity(kind, entity) {
  if (kind === 'sensor') {
    return entity?.wiring?.source || entity?.config?.wiring?.source || 'dht22'
  }
  const w = entity?.config?.wiring || entity?.wiring
  if (w?.source === 'relay_hat' || w?.channel != null) return 'relay_hat'
  return 'gpio_relay'
}
