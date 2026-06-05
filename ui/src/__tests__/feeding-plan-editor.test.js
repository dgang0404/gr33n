import { describe, it, expect } from 'vitest'
import { parseDailyFeedTime, buildDailyFeedCron } from '../lib/dailyFeedSchedule.js'
import {
  buildFeedingPlanDraft,
  buildProgramPatch,
  buildSchedulePatch,
  buildWizardSchedulePayload,
} from '../lib/feedingPlanEdit.js'

describe('Phase 47 WS3 — daily feed schedule helpers', () => {
  it('round-trips daily feed time without showing cron to operators', () => {
    expect(parseDailyFeedTime('0 8 * * *')).toBe('08:00')
    expect(parseDailyFeedTime('30 18 * * *')).toBe('18:30')
    expect(buildDailyFeedCron('08:00')).toBe('0 8 * * *')
    expect(buildDailyFeedCron('18:30')).toBe('30 18 * * *')
  })

  it('builds program and schedule patches from draft', () => {
    const program = {
      id: 10,
      name: 'Flower feed',
      is_active: true,
      reservoir_id: 1,
      target_zone_id: 3,
    }
    const schedule = {
      id: 20,
      name: 'Daily',
      schedule_type: 'cron',
      cron_expression: '0 6 * * *',
      timezone: 'America/New_York',
      is_active: true,
      preconditions: [],
    }
    const draft = buildFeedingPlanDraft({ activeProgram: program, schedule, ecTargets: [] })
    draft.volumeLiters = 0.5
    draft.dailyFeedTime = '07:30'
    draft.irrigationOnly = true
    draft.feedingPaused = true

    expect(buildProgramPatch(program, draft).total_volume_liters).toBe(0.5)
    expect(buildProgramPatch(program, draft).irrigation_only).toBe(true)

    const schedPatch = buildSchedulePatch(schedule, draft, buildDailyFeedCron(draft.dailyFeedTime))
    expect(schedPatch.cron_expression).toBe('30 7 * * *')
    expect(schedPatch.is_active).toBe(false)
  })

  it('creates wizard schedule without cron in the payload name', () => {
    const payload = buildWizardSchedulePayload({
      zoneName: 'Flower Room',
      farmTimezone: 'UTC',
      dailyFeedTime: '06:00',
    })
    expect(payload.name).toContain('Flower Room')
    expect(payload.cron_expression).toBe('0 6 * * *')
    expect(payload.is_active).toBe(true)
  })
})
