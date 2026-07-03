/**
 * Phase 119 — Raspberry Pi 40-pin header map (physical ↔ BCM).
 * Layout: pin 1 top-left (3.3V), USB ports toward the bottom.
 */

import { resolveWiring } from './hardwareWiring.js'

/** @typedef {'gpio'|'power3v3'|'power5v'|'ground'|'reserved'} PinRole */

/**
 * @typedef {object} PinDef
 * @property {number} physical
 * @property {number|null} bcm
 * @property {PinRole} role
 * @property {string[]} buses
 * @property {string} [label]
 */

/** Standard 40-pin header (Pi 2/3/4/5, 40-pin HATs). */
export const PI_HEADER_PINS = [
  { physical: 1, bcm: null, role: 'power3v3', buses: [], label: '3.3V' },
  { physical: 2, bcm: null, role: 'power5v', buses: [], label: '5V' },
  { physical: 3, bcm: 2, role: 'gpio', buses: ['i2c1'], label: 'SDA' },
  { physical: 4, bcm: null, role: 'power5v', buses: [], label: '5V' },
  { physical: 5, bcm: 3, role: 'gpio', buses: ['i2c1'], label: 'SCL' },
  { physical: 6, bcm: null, role: 'ground', buses: [], label: 'GND' },
  { physical: 7, bcm: 4, role: 'gpio', buses: [], label: 'GPIO4' },
  { physical: 8, bcm: 14, role: 'gpio', buses: ['uart0'], label: 'TXD' },
  { physical: 9, bcm: null, role: 'ground', buses: [], label: 'GND' },
  { physical: 10, bcm: 15, role: 'gpio', buses: ['uart0'], label: 'RXD' },
  { physical: 11, bcm: 17, role: 'gpio', buses: [], label: 'GPIO17' },
  { physical: 12, bcm: 18, role: 'gpio', buses: [], label: 'GPIO18' },
  { physical: 13, bcm: 27, role: 'gpio', buses: [], label: 'GPIO27' },
  { physical: 14, bcm: null, role: 'ground', buses: [], label: 'GND' },
  { physical: 15, bcm: 22, role: 'gpio', buses: [], label: 'GPIO22' },
  { physical: 16, bcm: 23, role: 'gpio', buses: [], label: 'GPIO23' },
  { physical: 17, bcm: null, role: 'power3v3', buses: [], label: '3.3V' },
  { physical: 18, bcm: 24, role: 'gpio', buses: [], label: 'GPIO24' },
  { physical: 19, bcm: 10, role: 'gpio', buses: ['spi0'], label: 'MOSI' },
  { physical: 20, bcm: null, role: 'ground', buses: [], label: 'GND' },
  { physical: 21, bcm: 9, role: 'gpio', buses: ['spi0'], label: 'MISO' },
  { physical: 22, bcm: 25, role: 'gpio', buses: [], label: 'GPIO25' },
  { physical: 23, bcm: 11, role: 'gpio', buses: ['spi0'], label: 'SCLK' },
  { physical: 24, bcm: 8, role: 'gpio', buses: ['spi0'], label: 'CE0' },
  { physical: 25, bcm: null, role: 'ground', buses: [], label: 'GND' },
  { physical: 26, bcm: 7, role: 'gpio', buses: ['spi0'], label: 'CE1' },
  { physical: 27, bcm: 0, role: 'reserved', buses: ['i2c0'], label: 'ID_SD' },
  { physical: 28, bcm: 1, role: 'reserved', buses: ['i2c0'], label: 'ID_SC' },
  { physical: 29, bcm: 5, role: 'gpio', buses: [], label: 'GPIO5' },
  { physical: 30, bcm: null, role: 'ground', buses: [], label: 'GND' },
  { physical: 31, bcm: 6, role: 'gpio', buses: [], label: 'GPIO6' },
  { physical: 32, bcm: 12, role: 'gpio', buses: [], label: 'GPIO12' },
  { physical: 33, bcm: 13, role: 'gpio', buses: [], label: 'GPIO13' },
  { physical: 34, bcm: null, role: 'ground', buses: [], label: 'GND' },
  { physical: 35, bcm: 19, role: 'gpio', buses: [], label: 'GPIO19' },
  { physical: 36, bcm: 16, role: 'gpio', buses: [], label: 'GPIO16' },
  { physical: 37, bcm: 26, role: 'gpio', buses: [], label: 'GPIO26' },
  { physical: 38, bcm: 20, role: 'gpio', buses: [], label: 'GPIO20' },
  { physical: 39, bcm: null, role: 'ground', buses: [], label: 'GND' },
  { physical: 40, bcm: 21, role: 'gpio', buses: [], label: 'GPIO21' },
]

export const I2C_BUS_PHYSICAL_PINS = [3, 5]
export const UART_BUS_PHYSICAL_PINS = [8, 10]

const I2C_SOURCES = new Set(['ads1115', 'bh1750'])
const UART_SOURCES = new Set(['mhz19'])

/** @param {number} bcm */
export function pinByBcm(bcm) {
  return PI_HEADER_PINS.find((p) => p.bcm === bcm) || null
}

/** @param {number} physical */
export function pinByPhysical(physical) {
  return PI_HEADER_PINS.find((p) => p.physical === physical) || null
}

/** @param {string[]} roles @param {number|null|undefined} bcmPin */
export function physicalPinsForHookupRoles(roles, bcmPin = null) {
  /** @type {Set<number>} */
  const out = new Set()
  for (const role of roles || []) {
    if (role === 'gpio' && bcmPin != null) {
      const p = pinByBcm(bcmPin)
      if (p) out.add(p.physical)
      continue
    }
    if (role === 'i2c_sda') {
      out.add(3)
      continue
    }
    if (role === 'i2c_scl') {
      out.add(5)
      continue
    }
    if (role === 'uart_tx') {
      out.add(8)
      continue
    }
    if (role === 'uart_rx') {
      out.add(10)
      continue
    }
    for (const pin of PI_HEADER_PINS) {
      if (role === 'power3v3' && pin.role === 'power3v3') out.add(pin.physical)
      if (role === 'power5v' && pin.role === 'power5v') out.add(pin.physical)
      if (role === 'gnd' && pin.role === 'ground') out.add(pin.physical)
    }
  }
  return out
}

/** Odd pins column (left), even pins column (right) — 20 rows. */
export function headerGridRows() {
  const rows = []
  for (let row = 0; row < 20; row += 1) {
    const odd = PI_HEADER_PINS.find((p) => p.physical === row * 2 + 1)
    const even = PI_HEADER_PINS.find((p) => p.physical === row * 2 + 2)
    rows.push({ row, left: odd, right: even })
  }
  return rows
}

function channelFromActuator(a) {
  const hi = a?.hardware_identifier
  if (hi == null || hi === '') return null
  const m = String(hi).match(/(\d+)$/)
  if (!m) return null
  const n = parseInt(m[1], 10)
  return Number.isFinite(n) && n >= 0 ? n : null
}

function deviceIdForEntity(entity, wiring) {
  if (wiring?.device_id != null) return Number(wiring.device_id)
  if (entity?.device_id != null) return Number(entity.device_id)
  return null
}

function zoneIdForEntity(entity) {
  if (entity?.zone_id != null) return Number(entity.zone_id)
  return null
}

/**
 * @param {number} deviceId
 * @param {object[]} sensors
 * @param {object[]} actuators
 */
export function assignmentsForDevice(deviceId, sensors = [], actuators = []) {
  /** @type {Map<number, object[]>} */
  const byPhysical = new Map()
  /** @type {object[]} */
  const i2cAttachments = []
  /** @type {object[]} */
  const relayChannels = []
  /** @type {object[]} */
  const uartAttachments = []

  function addPinAssignment(physical, entry) {
    if (physical == null) return
    const list = byPhysical.get(physical) || []
    list.push(entry)
    byPhysical.set(physical, list)
  }

  for (const s of sensors) {
    const w = resolveWiring(s)
    const devId = deviceIdForEntity(s, w)
    if (devId !== deviceId) continue

    const base = {
      kind: 'sensor',
      id: s.id,
      name: s.name || s.sensor_type,
      zoneId: zoneIdForEntity(s),
      sensorType: s.sensor_type,
      source: w?.source || '',
    }

    if (w?.gpio_pin != null && I2C_SOURCES.has(w.source)) {
      i2cAttachments.push({
        ...base,
        bus: 'i2c1',
        i2cChannel: w.i2c_channel,
        label: `${String(w.source).toUpperCase()}${w.i2c_channel != null ? ` ch ${w.i2c_channel}` : ''}`,
      })
      continue
    }

    if (UART_SOURCES.has(w?.source) || w?.serial_port) {
      uartAttachments.push({
        ...base,
        serialPort: w?.serial_port || 'UART',
        label: w?.serial_port || String(w?.source || 'serial').toUpperCase(),
      })
      continue
    }

    if (w?.gpio_pin != null) {
      const pin = pinByBcm(Number(w.gpio_pin))
      addPinAssignment(pin?.physical, {
        ...base,
        bcm: Number(w.gpio_pin),
        label: `${base.name} · BCM ${w.gpio_pin}`,
      })
    }
  }

  for (const a of actuators) {
    const w = resolveWiring(a)
    const devId = deviceIdForEntity(a, w)
    if (devId !== deviceId) continue

    const ch = channelFromActuator(a)
    const base = {
      kind: 'actuator',
      id: a.id,
      name: a.name,
      zoneId: zoneIdForEntity(a),
      actuatorType: a.actuator_type,
      online: a.current_state_text === 'online',
    }

    if (ch != null) {
      relayChannels.push({
        ...base,
        channel: ch,
        label: `ch ${ch}`,
      })
      continue
    }

    if (w?.gpio_pin != null) {
      const pin = pinByBcm(Number(w.gpio_pin))
      addPinAssignment(pin?.physical, {
        ...base,
        bcm: Number(w.gpio_pin),
        label: `${base.name} · BCM ${w.gpio_pin}`,
      })
    }
  }

  relayChannels.sort((a, b) => a.channel - b.channel)
  i2cAttachments.sort((a, b) => String(a.label).localeCompare(String(b.label)))

  if (relayChannels.length || i2cAttachments.some((x) => I2C_SOURCES.has(x.source))) {
    const hatLabel = relayChannels.length
      ? `Relay HAT · ${relayChannels.length} channel(s)`
      : 'I2C bus'
    if (!i2cAttachments.some((x) => x.kind === 'hat')) {
      i2cAttachments.unshift({
        kind: 'hat',
        id: 'i2c-hat',
        name: hatLabel,
        label: hatLabel,
        zoneId: null,
      })
    }
  }

  return { byPhysical, i2cAttachments, relayChannels, uartAttachments }
}

/** Devices that have at least one wiring assignment. */
export function devicesWithWiring(devices, sensors, actuators) {
  const ids = new Set()
  for (const d of devices) {
    const { byPhysical, i2cAttachments, relayChannels, uartAttachments } =
      assignmentsForDevice(d.id, sensors, actuators)
    if (
      byPhysical.size > 0
      || i2cAttachments.length > 0
      || relayChannels.length > 0
      || uartAttachments.length > 0
    ) {
      ids.add(d.id)
    }
  }
  return devices.filter((d) => ids.has(d.id))
}

/** @param {PinDef} pin */
export function pinRoleClass(pin) {
  if (!pin) return 'bg-zinc-800'
  switch (pin.role) {
    case 'power3v3': return 'bg-red-900/60 border-red-800/80'
    case 'power5v': return 'bg-red-950/70 border-red-900/80'
    case 'ground': return 'bg-zinc-950 border-zinc-700'
    case 'reserved': return 'bg-zinc-800/80 border-zinc-600'
    default: return 'bg-zinc-900 border-zinc-700'
  }
}
