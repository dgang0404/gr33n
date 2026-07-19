/**
 * Phase 183 — when a zone is an animal primary paddock or aquaponics fish tank,
 * hide crop "start a grow" chrome (same hardware tabs still apply).
 */

export function animalGroupsForZone(groups, zoneId) {
  const zid = Number(zoneId)
  return (groups || []).filter(
    (g) => g.active !== false && Number(g.primary_zone_id) === zid,
  )
}

export function aquaponicsLoopForZone(loops, zoneId) {
  const zid = Number(zoneId)
  return (loops || []).find(
    (l) => l.active !== false
      && (Number(l.fish_tank_zone_id) === zid || Number(l.grow_bed_zone_id) === zid),
  ) || null
}

export function isFishTankZone(loops, zoneId) {
  const loop = aquaponicsLoopForZone(loops, zoneId)
  return loop != null && Number(loop.fish_tank_zone_id) === Number(zoneId)
}

export function showPlantGrowUI({ groups, loops, zoneId }) {
  if (animalGroupsForZone(groups, zoneId).length > 0) return false
  if (isFishTankZone(loops, zoneId)) return false
  return true
}
