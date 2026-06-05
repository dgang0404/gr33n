import { describe, it, expect } from 'vitest'
import {
  buildWaterGrowStory,
  formatLastFeedLine,
  formatEdgeQueueLine,
  formatQueueHeadLabel,
  lastZoneFeedEvent,
} from '../lib/zoneWaterGrowStory.js'

describe('Phase 40 WS5 — zone water grow story', () => {
  it('picks last feed event for zone', () => {
    const events = [
      { id: 1, zone_id: 3, applied_at: '2026-06-01T08:00:00Z', volume_applied_liters: 0.5, program_id: 10 },
      { id: 2, zone_id: 3, applied_at: '2026-06-04T08:00:00Z', volume_applied_liters: 0.9, ec_before_mscm: 1.8, ec_after_mscm: 2.1, program_id: 10 },
      { id: 3, zone_id: 99, applied_at: '2026-06-05T08:00:00Z', volume_applied_liters: 1 },
    ]
    const last = lastZoneFeedEvent(events, 3)
    expect(last.id).toBe(2)
    expect(formatLastFeedLine(last, 'Flower FFJ')).toContain('0.9L')
    expect(formatLastFeedLine(last, 'Flower FFJ')).toContain('Flower FFJ')
  })

  it('formats queue head command types for farmers', () => {
    expect(formatQueueHeadLabel({ command_type: 'mix_batch' })).toBe('Mix batch')
    expect(formatQueueHeadLabel({ command_type: 'pulse', payload: { duration_seconds: 30 } }))
      .toBe('Pump pulse (30s)')
    expect(formatEdgeQueueLine(2, { command_type: 'mix_batch' })).toBe('2 queued · next: Mix batch')
  })

  it('builds last/next/edge story lines', () => {
    const story = buildWaterGrowStory({
      zoneId: 3,
      events: [
        {
          id: 2,
          zone_id: 3,
          applied_at: '2026-06-04T08:00:00Z',
          volume_applied_liters: 0.9,
          ec_before_mscm: 1.8,
          ec_after_mscm: 2.1,
          program_id: 10,
        },
      ],
      programs: [{ id: 10, name: 'Flower FFJ Program' }],
      schedules: [
        {
          id: 20,
          name: 'Water Early Flower Daily',
          schedule_type: 'irrigation',
          cron_expression: '0 8 * * *',
          is_active: true,
        },
      ],
      activeProgram: { id: 10, name: 'Flower FFJ Program', schedule_id: 20 },
      waterStatus: { queue_depth: 1 },
      queueHead: { command_type: 'pulse', payload: { duration_seconds: 45 } },
    })

    expect(story.lastFeed.line).toContain('0.9L')
    expect(story.nextFeed.line).toContain('Flower FFJ Program')
    expect(story.nextFeed.line).toContain('8 AM')
    expect(story.edge.line).toContain('Pump pulse')
  })
})
