const LEGACY_STORAGE_KEY = 'gr33n_task_write_queue_v1'
const STORAGE_KEY = 'gr33n_offline_write_queue_v2'

function safeParse(raw) {
  try {
    const parsed = JSON.parse(raw)
    return Array.isArray(parsed) ? parsed : []
  } catch {
    return []
  }
}

/** Loads unified offline queue (task + cost writes); migrates legacy task-only key once. */
export function loadOfflineQueue() {
  if (typeof localStorage === 'undefined') return []
  let raw = localStorage.getItem(STORAGE_KEY)
  if (!raw) {
    const legacy = localStorage.getItem(LEGACY_STORAGE_KEY)
    if (legacy) {
      raw = legacy
      localStorage.setItem(STORAGE_KEY, legacy)
      localStorage.removeItem(LEGACY_STORAGE_KEY)
    }
  }
  return safeParse(raw || '[]')
}

export function saveOfflineQueue(items) {
  if (typeof localStorage === 'undefined') return
  localStorage.setItem(STORAGE_KEY, JSON.stringify(items))
}

/** @deprecated use loadOfflineQueue */
export function loadTaskQueue() {
  return loadOfflineQueue()
}

/** @deprecated use saveOfflineQueue */
export function saveTaskQueue(items) {
  saveOfflineQueue(items)
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

/**
 * @param {string} farmId
 * @param {object} payload API body for POST /farms/:id/costs
 * @param {object} [opts]
 * @param {string} [opts.idempotencyKey] defaults to random UUID
 * @param {string} [opts.receiptDataUrl] optional data URL (receipt queued until sync)
 * @param {string} [opts.receiptFileName] original filename for upload
 */
export function makeCreateCostItem(farmId, payload, opts = {}) {
  const now = new Date().toISOString()
  const clientCostId = makeQueueId('local-cost')
  const idempotencyKey =
    opts.idempotencyKey ||
    (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function'
      ? crypto.randomUUID()
      : makeQueueId('idem'))
  return {
    id: makeQueueId('cost-create'),
    farmId,
    type: 'create_cost',
    payload: { ...payload },
    clientCostId,
    idempotencyKey,
    receiptDataUrl: opts.receiptDataUrl || '',
    receiptFileName: opts.receiptFileName || '',
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

const MAX_RECEIPT_BYTES = 5 * 1024 * 1024

export function fileToDataUrl(file) {
  return new Promise((resolve, reject) => {
    if (!file) {
      resolve('')
      return
    }
    if (file.size > MAX_RECEIPT_BYTES) {
      reject(new Error('Receipt must be 5 MB or smaller'))
      return
    }
    const reader = new FileReader()
    reader.onload = () => resolve(String(reader.result || ''))
    reader.onerror = () => reject(reader.error || new Error('read failed'))
    reader.readAsDataURL(file)
  })
}

/** @param {string} dataUrl */
export function dataUrlToFile(dataUrl, filename = 'receipt') {
  const parts = dataUrl.split(',')
  if (parts.length < 2) throw new Error('invalid data url')
  const head = parts[0]
  const b64 = parts.slice(1).join(',')
  const mimeMatch = head.match(/data:(.*?);/)
  const mime = mimeMatch ? mimeMatch[1] : 'application/octet-stream'
  const binary = atob(b64)
  const bytes = new Uint8Array(binary.length)
  for (let i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i)
  return new File([bytes], filename, { type: mime })
}
