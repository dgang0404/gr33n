const STORAGE_KEY = 'gr33n_task_write_queue_v1'

function safeParse(raw) {
  try {
    const parsed = JSON.parse(raw)
    return Array.isArray(parsed) ? parsed : []
  } catch {
    return []
  }
}

export function loadTaskQueue() {
  if (typeof localStorage === 'undefined') return []
  return safeParse(localStorage.getItem(STORAGE_KEY) || '[]')
}

export function saveTaskQueue(items) {
  if (typeof localStorage === 'undefined') return
  localStorage.setItem(STORAGE_KEY, JSON.stringify(items))
}

export function makeQueueId(prefix = 'q') {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return `${prefix}-${crypto.randomUUID()}`
  }
  return `${prefix}-${Date.now()}-${Math.floor(Math.random() * 1e6)}`
}

export function makeCreateTaskItem(farmId, payload) {
  const now = new Date().toISOString()
  const clientTaskId = makeQueueId('local-task')
  return {
    id: makeQueueId('task-create'),
    farmId,
    type: 'create_task',
    payload,
    clientTaskId,
    attempts: 0,
    state: 'pending',
    lastError: '',
    createdAt: now,
    updatedAt: now,
  }
}

export function makeUpdateTaskStatusItem(farmId, taskRef, nextStatus) {
  const now = new Date().toISOString()
  return {
    id: makeQueueId('task-status'),
    farmId,
    type: 'update_task_status',
    payload: {
      taskId: typeof taskRef === 'number' ? taskRef : null,
      clientTaskId: typeof taskRef === 'string' ? taskRef : null,
      status: nextStatus,
    },
    attempts: 0,
    state: 'pending',
    lastError: '',
    createdAt: now,
    updatedAt: now,
  }
}

export function isRetryableTaskQueueError(err) {
  if (!err) return false
  if (!err.response) return true
  const code = Number(err.response.status || 0)
  return code >= 500 || code === 429
}

export function pendingCount(queue, farmId) {
  return queue.filter((i) => i.farmId === farmId && i.state !== 'synced').length
}
