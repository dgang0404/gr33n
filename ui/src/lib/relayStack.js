/**
 * Phase 120 — derive relay HAT stack layout from live channel assignments.
 * Matches PiSetupGuide dipTable (stack 0 @ 0x27, DIP = 3-bit stack level).
 */

/** @param {number} channel */
export function channelToStackLevel(channel) {
  return Math.floor(channel / 8)
}

/** @param {number} level 0–7 */
export function stackLevelToI2cAddress(level) {
  const n = 0x27 - level
  return `0x${n.toString(16).padStart(2, '0')}`
}

/** @param {number} level */
export function stackLevelToDipBits(level) {
  return {
    id0: (level & 1) === 1,
    id1: (level & 2) === 2,
    id2: (level & 4) === 4,
  }
}

/** @param {{ id0: boolean, id1: boolean, id2: boolean }} dip */
export function formatDipSwitch(dip) {
  return `ID0 ${dip.id0 ? 'ON' : 'OFF'} · ID1 ${dip.id1 ? 'ON' : 'OFF'} · ID2 ${dip.id2 ? 'ON' : 'OFF'}`
}

/**
 * @param {object[]} relayChannels — from assignmentsForDevice().relayChannels
 */
export function buildRelayStacks(relayChannels) {
  if (!relayChannels?.length) return []
  const maxLevel = Math.max(...relayChannels.map((r) => channelToStackLevel(r.channel)))
  /** @type {object[]} */
  const stacks = []
  for (let level = 0; level <= maxLevel; level += 1) {
    const dip = stackLevelToDipBits(level)
    const byChannel = new Map(
      relayChannels
        .filter((r) => channelToStackLevel(r.channel) === level)
        .map((r) => [r.channel, r]),
    )
    const slots = Array.from({ length: 8 }, (_, relayIdx) => {
      const ch = level * 8 + relayIdx
      return {
        channel: ch,
        relay: relayIdx + 1,
        assigned: byChannel.get(ch) || null,
      }
    })
    stacks.push({
      level,
      i2c: stackLevelToI2cAddress(level),
      dip,
      dipLabel: formatDipSwitch(dip),
      channelRange: `ch ${level * 8} – ${level * 8 + 7}`,
      slots,
    })
  }
  return stacks
}
