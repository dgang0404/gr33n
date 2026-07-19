import { describe, expect, it } from 'vitest'
import {
  animalGroupsForZone,
  aquaponicsLoopForZone,
  isFishTankZone,
  showPlantGrowUI,
} from '../lib/zoneDetailMode.js'

describe('zoneDetailMode', () => {
  it('hides plant grow UI for animal primary zones and fish tanks', () => {
    expect(showPlantGrowUI({
      groups: [{ id: 1, primary_zone_id: 55, active: true }],
      loops: [],
      zoneId: 55,
    })).toBe(false)
    expect(showPlantGrowUI({
      groups: [],
      loops: [{ id: 1, fish_tank_zone_id: 57, grow_bed_zone_id: 58, active: true }],
      zoneId: 57,
    })).toBe(false)
    expect(showPlantGrowUI({
      groups: [],
      loops: [{ id: 1, fish_tank_zone_id: 57, grow_bed_zone_id: 58, active: true }],
      zoneId: 58,
    })).toBe(true)
    expect(showPlantGrowUI({ groups: [], loops: [], zoneId: 1 })).toBe(true)
  })

  it('matches animal groups and aquaponics loops by zone id', () => {
    const groups = [{ id: 5, label: 'Laying flock', primary_zone_id: 55, active: true }]
    expect(animalGroupsForZone(groups, 55)).toHaveLength(1)
    expect(animalGroupsForZone(groups, 99)).toHaveLength(0)

    const loops = [{ id: 2, label: 'Tilapia loop', fish_tank_zone_id: 57, grow_bed_zone_id: 58, active: true }]
    expect(aquaponicsLoopForZone(loops, 57)?.label).toBe('Tilapia loop')
    expect(isFishTankZone(loops, 57)).toBe(true)
    expect(isFishTankZone(loops, 58)).toBe(false)
  })
})
