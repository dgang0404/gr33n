/**
 * Phase 58 WS4 / OC-58 — task consumptions & operator runtime closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'
import { validateConsumptionQty } from '../lib/taskConsumption.js'
import {
  refillTaskFromLowStock,
  buildCheckSensorTaskPayload,
  buildReviewFeedingPlanPayload,
  detectMissedFeedSchedule,
} from '../lib/taskTemplates.js'
import { countZoneOverdueTasks } from '../lib/zoneTasks.js'

const repoDocs = join(process.cwd(), '..', 'docs')
const repoRoot = join(process.cwd(), '..')

describe('Phase 58 WS4 / OC-58 — task consumptions closure', () => {
  it('documents consumptions and plan shipped', () => {
    const plan = readFileSync(join(repoDocs, 'plans/phase_58_task_consumptions_runtime.plan.md'), 'utf8')
    const tour = readFileSync(join(repoDocs, 'operator-tour.md'), 'utf8')
    expect(plan).toContain('**Shipped.**')
    expect(tour).toContain('Task consumptions')
    expect(existsSync(join(repoRoot, 'ui/src/components/TaskCompleteSheet.vue'))).toBe(true)
  })

  it('validateConsumptionQty blocks over-draw', () => {
    expect(validateConsumptionQty(5, { current_quantity_remaining: 3 })).toContain('Only 3 on hand')
    expect(validateConsumptionQty(2, { current_quantity_remaining: 5 })).toBe('')
  })

  it('task templates pre-fill refill and sensor check payloads', () => {
    const refill = refillTaskFromLowStock({
      inputName: 'CalMag',
      remaining: 1,
      threshold: 5,
      batch: { id: 9 },
    })
    expect(refill.title).toContain('Refill CalMag')
    expect(refill.template_id).toBe('refill')

    const check = buildCheckSensorTaskPayload(
      { id: 2, triggering_event_source_type: 'sensor', triggering_event_source_id: 7 },
      [{ id: 7, name: 'Flower temp' }],
    )
    expect(check.title).toContain('Flower temp')

    const review = buildReviewFeedingPlanPayload({ id: 3, name: '9am feed' })
    expect(review.title).toContain('9am feed')
  })

  it('detectMissedFeedSchedule finds overdue active schedules', () => {
    const past = new Date(Date.now() - 60 * 60 * 1000).toISOString()
    const sched = detectMissedFeedSchedule([
      { is_active: true, next_expected_trigger_time: past, last_triggered_time: null, name: 'Morning' },
    ])
    expect(sched?.name).toBe('Morning')
  })

  it('UI wires complete sheet, store actions, and farm consumptions API', () => {
    const tasks = readFileSync(join(process.cwd(), 'src/views/Tasks.vue'), 'utf8')
    const store = readFileSync(join(process.cwd(), 'src/stores/farm.js'), 'utf8')
    const routes = readFileSync(join(repoRoot, 'cmd/api/routes.go'), 'utf8')
    expect(tasks).toContain('TaskCompleteSheet')
    expect(tasks).toContain('recordTaskConsumption')
    expect(store).toContain('loadFarmTaskConsumptions')
    expect(routes).toContain('GET /farms/{id}/task-consumptions')
    expect(countZoneOverdueTasks([{ zone_id: 1, status: 'todo', due_date: '2000-01-01' }], 1)).toBe(1)
  })
})
