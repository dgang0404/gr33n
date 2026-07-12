/**
 * Phase 142 — Virtual Pi field validation status for operator banners.
 */

import { assignmentsForDevice } from './piPinMap.js'
import { wiringDriftStatus } from './piConfigDrift.js'

/**
 * @param {number} deviceId
 * @param {object[]} sensors
 * @param {object[]} actuators
 */
export function wiringCountForDevice(deviceId, sensors, actuators) {
  if (!deviceId) return 0
  const { byPhysical, i2cAttachments, relayChannels, uartAttachments } =
    assignmentsForDevice(deviceId, sensors, actuators)
  const gpioPins = byPhysical.size
  const buses =
    i2cAttachments.length + relayChannels.length + uartAttachments.length
  return gpioPins + buses
}

/**
 * @typedef {object} ValidationChecklistItem
 * @property {string} id
 * @property {string} label
 * @property {boolean} ok
 */

/**
 * @param {object} input
 * @param {object|null|undefined} input.device
 * @param {object[]} input.sensors
 * @param {object[]} input.actuators
 * @param {string} input.expectedConfigSha
 * @param {boolean} [input.configDownloaded]
 * @returns {{
 *   status: 'needs_wiring'|'stale'|'ready_dry_run'|'ready_live',
 *   title: string,
 *   hint: string,
 *   checklist: ValidationChecklistItem[],
 * }}
 */
export function computeVirtualPiValidation({
  device,
  sensors,
  actuators,
  expectedConfigSha,
  configDownloaded = false,
}) {
  const deviceId = device?.id
  const wiringCount = wiringCountForDevice(deviceId, sensors, actuators)
  const hasWiring = wiringCount > 0
  const hasConfig = Boolean(String(expectedConfigSha || '').trim())
  const drift = wiringDriftStatus(device, expectedConfigSha)

  const checklist = [
    {
      id: 'wiring',
      label: hasWiring
        ? `Wiring assigned (${wiringCount} pin/channel${wiringCount === 1 ? '' : 's'})`
        : 'Wiring assigned — add sensors or actuators on zone pages',
      ok: hasWiring,
    },
    {
      id: 'config',
      label: hasConfig ? 'Platform config.yaml ready to export' : 'Platform config — loading or unavailable',
      ok: hasConfig,
    },
    {
      id: 'drift',
      label:
        drift === 'synced'
          ? 'Pi wiring hash matches platform'
          : drift === 'stale'
            ? 'Pi config hash stale — download or Notify Pi to reload'
            : 'Pi drift unknown — OK for laptop dry run after download',
      ok: drift !== 'stale',
    },
  ]

  if (configDownloaded) {
    checklist.push({
      id: 'downloaded',
      label: 'Config downloaded this session',
      ok: true,
    })
  }

  if (!hasWiring) {
    return {
      status: 'needs_wiring',
      title: 'Needs wiring before field validation',
      hint: 'Wire at least one sensor or actuator to this device, then download config.yaml.',
      checklist,
    }
  }

  if (drift === 'stale') {
    return {
      status: 'stale',
      title: 'Stale Pi config — fix before dry run',
      hint: 'Download config.yaml to the Pi or use Notify Pi to reload, then re-check drift.',
      checklist,
    }
  }

  if (drift === 'synced') {
    return {
      status: 'ready_live',
      title: 'Ready for live field validation',
      hint: 'Pi matches platform. Continue with live relay testing on the field.',
      checklist,
    }
  }

  return {
    status: 'ready_dry_run',
    title: 'Ready for LED simulation dry run',
    hint: 'Download config from Virtual Pi, merge into your Pi client config, then run the LED simulation demo.',
    checklist,
  }
}

/** @param {'needs_wiring'|'stale'|'ready_dry_run'|'ready_live'} status */
export function validationBannerClass(status) {
  switch (status) {
    case 'ready_live':
    case 'ready_dry_run':
      return 'border-green-800/50 bg-green-950/20 text-green-200'
    case 'stale':
      return 'border-amber-800/50 bg-amber-950/30 text-amber-200'
    default:
      return 'border-zinc-700 bg-zinc-900/60 text-zinc-300'
  }
}
