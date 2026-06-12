import { describe, it, expect } from 'vitest'
import {
  sensorPlantNeed,
  actuatorPlantNeed,
  supportsPulseCommand,
  PLANT_NEEDS,
} from '../lib/plantNeeds.js'

describe('plantNeeds', () => {
  it('classifies water sensors', () => {
    expect(sensorPlantNeed('ec')).toBe(PLANT_NEEDS.water)
    expect(sensorPlantNeed('soil_moisture')).toBe(PLANT_NEEDS.water)
  })

  it('classifies light sensors', () => {
    expect(sensorPlantNeed('lux')).toBe(PLANT_NEEDS.light)
    expect(sensorPlantNeed('par_umol')).toBe(PLANT_NEEDS.light)
  })

  it('classifies climate sensors including temp_f', () => {
    expect(sensorPlantNeed('air_temp')).toBe(PLANT_NEEDS.air)
    expect(sensorPlantNeed('humidity')).toBe(PLANT_NEEDS.air)
    expect(sensorPlantNeed('temp_f')).toBe(PLANT_NEEDS.air)
  })

  it('classifies actuators', () => {
    expect(actuatorPlantNeed('pump')).toBe(PLANT_NEEDS.water)
    expect(actuatorPlantNeed('grow_light')).toBe(PLANT_NEEDS.light)
    expect(actuatorPlantNeed('exhaust_fan')).toBe(PLANT_NEEDS.air)
  })

  it('pulse support for pumps and relays', () => {
    expect(supportsPulseCommand('pump')).toBe(true)
    expect(supportsPulseCommand('relay')).toBe(true)
    expect(supportsPulseCommand('grow_light')).toBe(false)
  })
})
