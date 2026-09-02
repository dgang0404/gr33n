import { describe, it, expect } from 'vitest'
import { resolveScheduleTarget, resolveRuleTarget } from '../lib/automationTargetLink.js'

describe('automationTargetLink (Phase 217 WS2/WS8)', () => {
  describe('resolveScheduleTarget', () => {
    it('resolves to a zone via a linked fertigation program', () => {
      const schedule = { id: 10 }
      const result = resolveScheduleTarget(schedule, {
        programs: [{ schedule_id: 10, target_zone_id: 3 }],
      })
      expect(result).toEqual({ to: { path: '/zones/3' }, navHint: '/zones', label: 'zone' })
    })

    it('resolves to the reservoirs tab when the program has no target zone', () => {
      const schedule = { id: 11 }
      const result = resolveScheduleTarget(schedule, {
        programs: [{ schedule_id: 11, target_zone_id: null, reservoir_id: 7 }],
      })
      expect(result.label).toBe('reservoir')
      expect(result.to.path).toBe('/feed-water')
    })

    it('resolves to a zone via a linked lighting program', () => {
      const schedule = { id: 12 }
      const result = resolveScheduleTarget(schedule, {
        lightingPrograms: [{ schedule_on_id: 12, schedule_off_id: 13, zone_id: 4 }],
      })
      expect(result).toEqual({ to: { path: '/zones/4' }, navHint: '/zones', label: 'zone' })
    })

    it('returns null when nothing links to the schedule', () => {
      expect(resolveScheduleTarget({ id: 99 }, {})).toBeNull()
    })

    it('returns null for a missing schedule', () => {
      expect(resolveScheduleTarget(null, {})).toBeNull()
    })
  })

  describe('resolveRuleTarget', () => {
    it('resolves to a zone via a control_actuator action', () => {
      const rule = { id: 20 }
      const result = resolveRuleTarget(rule, {
        actions: [{ action_type: 'control_actuator', target_actuator_id: 5 }],
        actuators: [{ id: 5, zone_id: 2 }],
      })
      expect(result).toEqual({ to: { path: '/zones/2' }, navHint: '/zones', label: 'zone' })
    })

    it('returns null when the action has no matching actuator', () => {
      const rule = { id: 21 }
      const result = resolveRuleTarget(rule, {
        actions: [{ action_type: 'control_actuator', target_actuator_id: 999 }],
        actuators: [{ id: 5, zone_id: 2 }],
      })
      expect(result).toBeNull()
    })

    it('returns null when the actuator has no zone_id', () => {
      const rule = { id: 22 }
      const result = resolveRuleTarget(rule, {
        actions: [{ action_type: 'control_actuator', target_actuator_id: 5 }],
        actuators: [{ id: 5, zone_id: null }],
      })
      expect(result).toBeNull()
    })

    it('returns null for a notification-only rule (no control_actuator action)', () => {
      const rule = { id: 23 }
      const result = resolveRuleTarget(rule, {
        actions: [{ action_type: 'send_notification' }],
        actuators: [],
      })
      expect(result).toBeNull()
    })
  })
})
