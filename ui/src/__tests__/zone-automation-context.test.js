import { describe, it, expect, vi } from 'vitest'
import {
  zoneAutomationForNeed,
  ruleAppliesToNeed,
  isGreenhouseRule,
  greenhouseRuleBadges,
} from '../lib/zoneAutomationContext.js'
import { PLANT_NEEDS } from '../lib/plantNeeds.js'

describe('Phase 40 WS3 — zone automation context', () => {
  it('detects greenhouse rules', () => {
    expect(isGreenhouseRule({ name: 'GH — High lux deploy shade' })).toBe(true)
    expect(isGreenhouseRule({ name: 'AUTO Light ON' })).toBe(false)
  })

  it('filters light cron rules for zone', () => {
    const rule = {
      id: 1,
      name: 'AUTO Light ON 12/12 Flower',
      is_active: true,
      trigger_configuration: {
        cron: '0 6 * * *',
        target_zone: 'Flower Room',
        action: 'actuator_on',
      },
      conditions_jsonb: [],
    }
    expect(ruleAppliesToNeed(rule, PLANT_NEEDS.light, {
      zoneId: 3,
      zoneName: 'Flower Room',
      sensors: [],
    })).toBe(true)
  })

  it('lists zone irrigation schedule on water need', () => {
    const result = zoneAutomationForNeed(PLANT_NEEDS.water, {
      zoneId: 3,
      zoneName: 'Flower Room',
      sensors: [],
      rules: [],
      schedules: [
        {
          id: 20,
          name: 'Water Early Flower Daily',
          schedule_type: 'irrigation',
          cron_expression: '0 8 * * *',
          is_active: true,
          description: 'Zone: Flower Room.',
        },
      ],
      activeProgram: { schedule_id: 20, name: 'Flower FFJ Program' },
      lightingPrograms: [],
    })
    expect(result.schedules).toHaveLength(1)
    expect(result.schedules[0].runsLabel).toContain('8 AM')
    expect(result.schedules[0].linkedName).toBe('Flower FFJ Program')
  })

  it('adds lux interlock badge when GH shade rule lacks sensor', () => {
    const badges = greenhouseRuleBadges(
      { name: 'GH — High lux deploy shade' },
      [{ sensor_type: 'humidity' }],
    )
    expect(badges.some((b) => b.id === 'no-lux')).toBe(true)
  })
})
