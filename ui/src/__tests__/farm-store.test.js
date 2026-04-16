import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'

vi.mock('../api', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    patch: vi.fn(),
    delete: vi.fn(),
    interceptors: { request: { use: vi.fn() }, response: { use: vi.fn() } },
  },
}))

import { useFarmStore } from '../stores/farm'
import api from '../api'

describe('farm store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    localStorage.clear()
  })

  it('loadAll() fetches farm, zones, sensors, devices, actuators', async () => {
    api.get.mockImplementation((url) => {
      if (url.endsWith('/farms/1'))           return Promise.resolve({ data: { id: 1, name: 'Test Farm' } })
      if (url.endsWith('/zones'))             return Promise.resolve({ data: [{ id: 1 }] })
      if (url.endsWith('/sensors'))           return Promise.resolve({ data: [{ id: 1 }, { id: 2 }] })
      if (url.endsWith('/devices'))           return Promise.resolve({ data: [{ id: 1, status: 'online' }] })
      if (url.endsWith('/actuators'))         return Promise.resolve({ data: [] })
      return Promise.resolve({ data: [] })
    })

    const farm = useFarmStore()
    await farm.loadAll(1)

    expect(api.get).toHaveBeenCalledWith('/farms/1')
    expect(api.get).toHaveBeenCalledWith('/farms/1/zones')
    expect(api.get).toHaveBeenCalledWith('/farms/1/sensors')
    expect(api.get).toHaveBeenCalledWith('/farms/1/devices')
    expect(api.get).toHaveBeenCalledWith('/farms/1/actuators')
    expect(farm.farm.name).toBe('Test Farm')
    expect(farm.zones).toHaveLength(1)
    expect(farm.sensors).toHaveLength(2)
    expect(farm.devices).toHaveLength(1)
    expect(farm.loading).toBe(false)
  })

  it('loadAll() sets error on failure', async () => {
    api.get.mockRejectedValue(new Error('network down'))
    const farm = useFarmStore()

    await farm.loadAll(1)

    expect(farm.error).toBe('network down')
    expect(farm.loading).toBe(false)
  })

  it('activeDevices getter filters online devices', async () => {
    const farm = useFarmStore()
    farm.devices = [
      { id: 1, status: 'online' },
      { id: 2, status: 'offline' },
      { id: 3, status: 'online' },
    ]

    expect(farm.activeDevices).toHaveLength(2)
  })

  it('queues task create when network is unavailable', async () => {
    const farm = useFarmStore()
    api.post.mockRejectedValueOnce(new Error('network down'))

    const created = await farm.createTask(1, { title: 'Offline task', priority: 1 })

    expect(created.id).toContain('local-task-')
    expect(farm.taskWriteQueue).toHaveLength(1)
    expect(farm.taskWriteQueue[0].type).toBe('create_task')
    expect(farm.taskQueuePendingCount(1)).toBe(1)
  })

  it('queues cost create when network is unavailable', async () => {
    const farm = useFarmStore()
    api.post.mockRejectedValueOnce(new Error('network down'))

    const created = await farm.createCost(1, {
      transaction_date: '2026-04-16',
      category: 'miscellaneous',
      amount: 12.5,
      currency: 'USD',
      is_income: false,
    })

    expect(String(created.id)).toContain('local-cost-')
    expect(created._offline?.queued).toBe(true)
    expect(farm.taskWriteQueue).toHaveLength(1)
    expect(farm.taskWriteQueue[0].type).toBe('create_cost')
    expect(farm.taskWriteQueue[0].idempotencyKey).toBeTruthy()
    expect(farm.taskWriteQueue[0].payload.amount).toBe(12.5)
  })

  it('can retry and discard queued items', async () => {
    const farm = useFarmStore()
    api.post.mockRejectedValueOnce(new Error('network down'))
    const created = await farm.createTask(1, { title: 'Offline task', priority: 1 })
    const queueId = farm.taskWriteQueue[0].id

    farm.taskWriteQueue[0].state = 'failed'
    farm.taskWriteQueue[0].lastError = 'validation failed'
    const retried = farm.retryTaskQueueItem(queueId)
    expect(retried).toBe(true)
    expect(farm.taskWriteQueue[0].state).toBe('pending')
    expect(farm.taskWriteQueue[0].lastError).toBe('')

    const discarded = farm.discardTaskQueueItem(queueId)
    expect(discarded).toBe(true)
    expect(farm.taskWriteQueue).toHaveLength(0)
    expect(farm.tasks.find((t) => t.id === created.id)).toBeUndefined()
  })
})
