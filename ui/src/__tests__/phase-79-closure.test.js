import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { setActivePinia, createPinia } from 'pinia'

vi.mock('vue-router', () => ({
  useRoute: () => ({ query: {}, path: '/zones/2' }),
  useRouter: () => ({ push: vi.fn(), replace: vi.fn() }),
}))

vi.mock('../api', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    patch: vi.fn(),
    interceptors: { request: { use: vi.fn() }, response: { use: vi.fn() } },
  },
}))

import Tasks from '../views/Tasks.vue'
import { useFarmStore } from '../stores/farm'
import { operatorConcept, COMFORT_WORKSPACE_CONCEPTS } from '../lib/operatorConcepts.js'
import { WORKSPACES } from '../lib/workspaces.js'

describe('Phase 79 — tasks fix, concepts, inventory', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('Tasks embedded with lockZoneId does not throw filterZone TDZ', async () => {
    const store = useFarmStore()
    store.zones = [{ id: 2, name: 'Flower Room' }]
    store.tasks = [{ id: 1, title: 'Check EC', zone_id: 2, status: 'todo' }]
    store.loadTasks = vi.fn().mockResolvedValue([])
    store.loadSchedules = vi.fn().mockResolvedValue([])
    store.loadNfBatches = vi.fn().mockResolvedValue([])
    store.loadNfInputs = vi.fn().mockResolvedValue([])
    store.loadFarmTaskConsumptions = vi.fn().mockResolvedValue([])

    const wrapper = mount(Tasks, {
      props: { embedded: true, lockZoneId: 2 },
      global: {
        stubs: {
          RouterLink: true,
          HelpTip: true,
          ZoneContextBanner: true,
          EmptyStateHint: true,
          TaskCompleteSheet: true,
        },
      },
    })

    await flushPromises()
    expect(wrapper.text()).toContain('Check EC')
  })

  it('operator concepts map to distinct DB tables', () => {
    expect(operatorConcept('task')?.dbTable).toContain('tasks')
    expect(operatorConcept('rule')?.dbTable).toContain('automation_rules')
    expect(operatorConcept('schedule')?.dbTable).toContain('schedules')
    expect(operatorConcept('comfort_band')?.dbTable).toContain('zone_setpoints')
    expect(operatorConcept('alert')?.dbTable).toContain('alerts')
    expect(COMFORT_WORKSPACE_CONCEPTS.length).toBeGreaterThanOrEqual(6)
  })

  it('money workspace includes inventory and recipes tab', () => {
    const tabs = WORKSPACES.money.tabs.map((t) => t.id)
    expect(tabs).toContain('inventory')
    expect(WORKSPACES.money.absorbs['/inventory']).toEqual({ tab: 'inventory' })
  })
})
