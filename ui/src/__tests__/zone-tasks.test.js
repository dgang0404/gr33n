import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { setActivePinia, createPinia } from 'pinia'

vi.mock('../api', () => ({
  default: {
    patch: vi.fn(),
    put: vi.fn(),
    interceptors: { request: { use: vi.fn() }, response: { use: vi.fn() } },
  },
}))

import ZoneTasksPanel from '../components/ZoneTasksPanel.vue'
import { useFarmStore } from '../stores/farm.js'
import { useFarmContextStore } from '../stores/farmContext.js'
import {
  zoneTasksDueToday,
  snoozeDueDateToTomorrow,
  formatTaskDueLabel,
} from '../lib/zoneTasks.js'

describe('Phase 40 WS6 — zone tasks', () => {
  const today = new Date().toISOString().slice(0, 10)

  it('filters due today tasks for zone', () => {
    const tasks = [
      { id: 1, zone_id: 3, status: 'todo', due_date: today, title: 'Defoliate' },
      { id: 2, zone_id: 3, status: 'todo', due_date: '2099-01-01', title: 'Future' },
      { id: 3, zone_id: 99, status: 'todo', due_date: today, title: 'Other zone' },
      { id: 4, zone_id: 3, status: 'completed', due_date: today, title: 'Done' },
    ]
    const due = zoneTasksDueToday(tasks, 3)
    expect(due).toHaveLength(1)
    expect(due[0].id).toBe(1)
  })

  it('snoozes due date to tomorrow', () => {
    expect(snoozeDueDateToTomorrow(today)).not.toBe(today)
    expect(formatTaskDueLabel(today)).toBe('Due today')
  })

  it('opens complete sheet from zone panel Done button', async () => {
    setActivePinia(createPinia())
    const store = useFarmStore()
    const farmContext = useFarmContextStore()
    farmContext.farmId = 1
    store.loadNfBatches = vi.fn().mockResolvedValue([])
    store.loadNfInputs = vi.fn().mockResolvedValue([])

    const wrapper = mount(ZoneTasksPanel, {
      props: {
        zoneId: 3,
        tasks: [
          { id: 5, zone_id: 3, status: 'todo', due_date: today, title: 'Refill reservoir' },
        ],
      },
      global: { stubs: { RouterLink: true } },
    })

    await wrapper.find('[data-test="zone-task-complete-5"]').trigger('click')
    await flushPromises()

    expect(store.loadNfBatches).toHaveBeenCalled()
    expect(wrapper.find('[data-test="task-complete-sheet"]').exists()).toBe(true)
  })
})
