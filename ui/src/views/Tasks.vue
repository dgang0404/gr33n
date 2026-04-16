<template>
  <div class="p-6">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-xl font-semibold text-white">Tasks</h1>
        <p
          v-if="syncStatusText"
          class="mt-1 text-[11px]"
          :class="syncStatusClass"
        >
          {{ syncStatusText }}
        </p>
      </div>
      <div class="flex items-center gap-3">
        <span
          v-if="!isOnline"
          class="text-[11px] px-2 py-1 rounded border border-amber-700 bg-amber-900/40 text-amber-300"
        >
          Offline mode
        </span>
        <span
          v-if="pendingWrites > 0"
          class="text-[11px] px-2 py-1 rounded border border-blue-700 bg-blue-900/40 text-blue-300"
        >
          {{ pendingWrites }} queued write{{ pendingWrites === 1 ? '' : 's' }}
        </span>
        <button
          type="button"
          @click="showForm = !showForm"
          class="text-xs font-medium px-3 py-1.5 rounded-lg bg-green-900/50 text-green-400 border border-green-800 hover:bg-green-900/70"
        >
          + New task
        </button>
        <button
          type="button"
          v-if="pendingWrites > 0 || queueHasStale"
          :disabled="syncing"
          @click="syncNow"
          class="text-xs font-medium px-3 py-1.5 rounded-lg bg-zinc-800 text-zinc-200 border border-zinc-700 hover:bg-zinc-700 disabled:opacity-40"
        >
          {{ syncing ? 'Syncing…' : 'Sync now' }}
        </button>
        <button
          type="button"
          v-if="queueItems.length > 0"
          @click="showQueueDetails = true"
          class="text-xs font-medium px-3 py-1.5 rounded-lg bg-zinc-900 text-zinc-200 border border-zinc-700 hover:bg-zinc-800"
        >
          Queue details
        </button>
        <span class="text-xs text-zinc-500">{{ tasks.length }} tasks</span>
      </div>
    </div>

    <p v-if="queueHasStale" class="mb-3 text-xs text-amber-300">
      Some queued updates need review due to server-side conflicts. Open the task card and retry after checking latest data.
    </p>

    <div
      v-if="showForm"
      class="mb-6 bg-zinc-900 border border-zinc-800 rounded-xl p-4 space-y-3 max-w-xl"
    >
      <h2 class="text-sm font-medium text-white">Create task</h2>
      <div>
        <label class="block text-xs text-zinc-500 mb-1">Title</label>
        <input v-model="form.title" type="text" required
          class="w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white" />
      </div>
      <div>
        <label class="block text-xs text-zinc-500 mb-1">Description</label>
        <textarea v-model="form.description" rows="2"
          class="w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white" />
      </div>
      <div class="grid grid-cols-2 gap-3">
        <div>
          <label class="block text-xs text-zinc-500 mb-1">Zone</label>
          <select v-model="form.zone_id"
            class="w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white">
            <option value="">—</option>
            <option v-for="z in store.zones" :key="z.id" :value="String(z.id)">{{ z.name }}</option>
          </select>
        </div>
        <div>
          <label class="block text-xs text-zinc-500 mb-1">Priority</label>
          <select v-model.number="form.priority"
            class="w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white">
            <option :value="0">Low</option>
            <option :value="1">Normal</option>
            <option :value="2">High</option>
            <option :value="3">Urgent</option>
          </select>
        </div>
      </div>
      <div class="grid grid-cols-2 gap-3">
        <div>
          <label class="block text-xs text-zinc-500 mb-1">Due date</label>
          <input v-model="form.due_date" type="date"
            class="w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white" />
        </div>
        <div>
          <label class="block text-xs text-zinc-500 mb-1">Type</label>
          <input v-model="form.task_type" type="text" placeholder="e.g. inspection"
            class="w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white" />
        </div>
      </div>
      <p v-if="formError" class="text-xs text-red-400">{{ formError }}</p>
      <div class="flex gap-2">
        <button
          type="button"
          :disabled="submitting || !form.title.trim()"
          @click="submitTask"
          class="text-xs font-medium px-3 py-1.5 rounded-lg bg-green-700 text-white disabled:opacity-40"
        >
          Create
        </button>
        <button type="button" @click="showForm = false"
          class="text-xs text-zinc-500 hover:text-zinc-300">Cancel</button>
      </div>
    </div>

    <div v-if="loading" class="text-zinc-400 text-sm">Loading tasks…</div>

    <div v-else class="grid grid-cols-1 md:grid-cols-3 gap-4 items-start">
      <div
        v-for="col in COLUMNS"
        :key="col.id"
        class="bg-zinc-900 border border-zinc-800 rounded-xl p-4 flex flex-col gap-3"
      >
        <div class="flex items-center justify-between pb-1 border-b border-zinc-800">
          <span class="flex items-center gap-2 font-medium text-white">
            <span>{{ col.icon }}</span>{{ col.label }}
          </span>
          <span class="text-xs bg-zinc-800 text-zinc-400 px-2 py-0.5 rounded-full">
            {{ colTasks(col).length }}
          </span>
        </div>

        <p v-if="!colTasks(col).length" class="text-zinc-700 text-sm py-4 text-center">
          No tasks
        </p>

        <div
          v-for="task in colTasks(col)"
          :key="task.id"
          class="bg-zinc-950 border border-zinc-800 rounded-lg p-3 hover:border-zinc-700 transition-colors"
        >
          <div class="flex items-start justify-between gap-2 mb-1.5">
            <p class="text-white text-sm font-medium leading-snug">{{ task.title }}</p>
            <span :class="priorityBadge(task.priority)"
              class="shrink-0 text-xs px-1.5 py-0.5 rounded">
              {{ PRIORITY_LABELS[task.priority] ?? 'normal' }}
            </span>
          </div>
          <p
            v-if="task._offline?.queued"
            class="text-[11px] mb-1"
            :class="task._offline?.stale ? 'text-amber-300' : 'text-blue-300'"
          >
            {{ task._offline?.stale ? `Sync conflict: ${task._offline?.conflict || 'needs retry'}` : 'Queued for sync' }}
          </p>
          <div v-if="task._offline?.queueItemId" class="mb-2 flex gap-2">
            <button
              type="button"
              class="text-[11px] px-2 py-1 rounded bg-zinc-800 border border-zinc-700 text-zinc-200 hover:bg-zinc-700"
              @click="retryQueueItem(task._offline.queueItemId)"
            >
              Retry
            </button>
            <button
              type="button"
              class="text-[11px] px-2 py-1 rounded bg-zinc-900 border border-zinc-700 text-zinc-400 hover:text-zinc-200"
              @click="discardQueueItem(task._offline.queueItemId)"
            >
              Discard
            </button>
          </div>
          <p v-if="task.description" class="text-zinc-500 text-xs line-clamp-2 mb-2">
            {{ task.description }}
          </p>
          <div class="flex items-center justify-between text-xs text-zinc-600 mb-2">
            <span>{{ zoneName(task.zone_id) }}</span>
            <span v-if="task.due_date">Due {{ task.due_date }}</span>
          </div>
          <button
            v-if="col.next"
            @click="advance(task, col.next)"
            class="text-xs text-green-600 hover:text-green-400 transition-colors"
          >
            → {{ col.nextLabel }}
          </button>
        </div>
      </div>
    </div>

    <div
      v-if="showQueueDetails"
      class="fixed inset-0 z-50 bg-black/70 p-4 flex items-center justify-center"
      @click.self="showQueueDetails = false"
    >
      <div class="w-full max-w-2xl bg-zinc-900 border border-zinc-700 rounded-xl p-4">
        <div class="flex items-center justify-between mb-3">
          <h2 class="text-sm font-semibold text-white">Queued task writes</h2>
          <button
            type="button"
            class="text-xs text-zinc-400 hover:text-zinc-200"
            @click="showQueueDetails = false"
          >
            Close
          </button>
        </div>

        <p v-if="queueItems.length === 0" class="text-xs text-zinc-500">
          No queued task writes.
        </p>
        <div v-else class="space-y-2 max-h-[60vh] overflow-auto pr-1">
          <div
            v-for="item in queueItems"
            :key="item.id"
            class="bg-zinc-950 border border-zinc-800 rounded-lg p-3"
          >
            <div class="flex items-center justify-between mb-1">
              <span class="text-xs font-medium text-zinc-100">{{ queueItemLabel(item) }}</span>
              <span
                class="text-[11px] px-2 py-0.5 rounded"
                :class="queueItemStateClass(item.state)"
              >
                {{ item.state }}
              </span>
            </div>
            <p class="text-[11px] text-zinc-400 mb-1">
              attempts: {{ item.attempts }} · updated: {{ formatQueueTime(item.updatedAt) }}
            </p>
            <p v-if="item.lastError" class="text-[11px] text-amber-300 mb-2">
              {{ item.lastError }}
            </p>
            <div class="flex gap-2">
              <button
                type="button"
                class="text-[11px] px-2 py-1 rounded bg-zinc-800 border border-zinc-700 text-zinc-200 hover:bg-zinc-700"
                @click="retryQueueItem(item.id)"
              >
                Retry
              </button>
              <button
                type="button"
                class="text-[11px] px-2 py-1 rounded bg-zinc-900 border border-zinc-700 text-zinc-400 hover:text-zinc-200"
                @click="discardQueueItem(item.id)"
              >
                Discard
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useFarmStore } from '../stores/farm'
import { useFarmContextStore } from '../stores/farmContext'

const store = useFarmStore()
const farmContext = useFarmContextStore()
const tasks = computed(() => store.tasks)
const pendingWrites = computed(() => store.taskQueuePendingCount(farmContext.farmId))
const queueHasStale = computed(() => tasks.value.some((t) => t._offline?.stale))
const queueItems = computed(() => {
  const fid = farmContext.farmId
  return store.taskWriteQueue.filter((i) => i.farmId === fid)
})
const syncStatusText = computed(() => {
  const status = store.taskSyncStatus
  if (!status?.lastAttemptAt) return ''
  const when = new Date(status.lastAttemptAt).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  if (status.lastResult === 'running') return `Last sync attempt ${when}: running`
  if (status.lastResult === 'partial_error') return `Last sync attempt ${when}: ${status.lastMessage || 'needs review'}`
  if (status.lastResult === 'ok') return `Last sync attempt ${when}: ${status.lastMessage || 'ok'}`
  return `Last sync attempt ${when}`
})
const syncStatusClass = computed(() => {
  const result = store.taskSyncStatus?.lastResult
  if (result === 'partial_error') return 'text-amber-300'
  if (result === 'ok') return 'text-emerald-300'
  return 'text-zinc-400'
})
const loading = ref(false)
const showForm = ref(false)
const showQueueDetails = ref(false)
const submitting = ref(false)
const syncing = ref(false)
const isOnline = ref(typeof navigator === 'undefined' ? true : navigator.onLine)
const formError = ref('')
const form = ref({
  title: '',
  description: '',
  zone_id: '',
  task_type: '',
  priority: 1,
  due_date: '',
})

onMounted(async () => {
  const fid = farmContext.farmId
  if (!store.zones.length && fid) await store.loadAll(fid)
  loading.value = true
  try {
    await store.loadTasks(fid)
    await syncNow()
  } finally { loading.value = false }
  window.addEventListener('online', onConnectionChange)
  window.addEventListener('offline', onConnectionChange)
})

onUnmounted(() => {
  window.removeEventListener('online', onConnectionChange)
  window.removeEventListener('offline', onConnectionChange)
})

async function submitTask() {
  formError.value = ''
  const fid = farmContext.farmId
  if (!fid) {
    formError.value = 'No farm selected'
    return
  }
  const title = form.value.title.trim()
  if (!title) return
  submitting.value = true
  try {
    const payload = {
      title,
      priority: form.value.priority,
    }
    const d = form.value.description.trim()
    if (d) payload.description = d
    if (form.value.zone_id) payload.zone_id = Number(form.value.zone_id)
    const tt = form.value.task_type.trim()
    if (tt) payload.task_type = tt
    if (form.value.due_date) payload.due_date = form.value.due_date
    await store.createTask(fid, payload)
    await store.loadTasks(fid)
    await syncNow()
    showForm.value = false
    form.value = {
      title: '',
      description: '',
      zone_id: '',
      task_type: '',
      priority: 1,
      due_date: '',
    }
  } catch (e) {
    formError.value = e.response?.data?.error || e.message || 'Failed to create task'
  } finally {
    submitting.value = false
  }
}

const COLUMNS = [
  { id: 'scheduled', label: 'Scheduled', icon: '📋',
    statuses: ['todo', 'on_hold', 'blocked_requires_input', 'pending_review'],
    next: 'in_progress', nextLabel: 'Start' },
  { id: 'running', label: 'Running', icon: '⚡',
    statuses: ['in_progress'],
    next: 'completed', nextLabel: 'Mark Done' },
  { id: 'done', label: 'Done', icon: '✅',
    statuses: ['completed', 'cancelled'],
    next: null, nextLabel: null },
]
function colTasks(col) { return tasks.value.filter(t => col.statuses.includes(t.status)) }
async function advance(task, nextStatus) {
  await store.updateTaskStatus(task.id, nextStatus)
  await syncNow()
}
function zoneName(id) {
  if (!id) return ''
  return store.zones.find(z => z.id === id)?.name ?? `Zone ${id}`
}
const PRIORITY_LABELS = { 0: 'low', 1: 'normal', 2: 'high', 3: 'urgent' }
const PRIORITY_BADGE = {
  0: 'bg-zinc-800 text-zinc-400', 1: 'bg-blue-900/50 text-blue-400',
  2: 'bg-yellow-900/50 text-yellow-400', 3: 'bg-red-900/50 text-red-400',
}
function priorityBadge(p) { return PRIORITY_BADGE[p] ?? PRIORITY_BADGE[1] }
function queueItemLabel(item) {
  if (item.type === 'create_task') {
    return `Create task: ${item.payload?.title || '(untitled)'}`
  }
  if (item.type === 'update_task_status') {
    return `Update status: ${item.payload?.status || 'unknown'}`
  }
  return item.type || 'queued write'
}
function queueItemStateClass(state) {
  if (state === 'failed') return 'bg-amber-900/40 border border-amber-700 text-amber-300'
  if (state === 'pending') return 'bg-blue-900/40 border border-blue-700 text-blue-300'
  return 'bg-zinc-800 text-zinc-300'
}
function formatQueueTime(value) {
  if (!value) return 'n/a'
  try {
    return new Date(value).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  } catch {
    return value
  }
}

async function syncNow() {
  const fid = farmContext.farmId
  if (!fid || syncing.value) return
  syncing.value = true
  try {
    await store.flushTaskWriteQueue({ farmId: fid })
    await store.loadTasks(fid)
  } finally {
    syncing.value = false
  }
}

function onConnectionChange() {
  isOnline.value = navigator.onLine
  if (isOnline.value) {
    void syncNow()
  }
}

async function retryQueueItem(queueItemId) {
  if (!queueItemId) return
  store.retryTaskQueueItem(queueItemId)
  await syncNow()
}

async function discardQueueItem(queueItemId) {
  if (!queueItemId) return
  store.discardTaskQueueItem(queueItemId)
  const fid = farmContext.farmId
  if (fid) {
    await store.loadTasks(fid)
  }
}
</script>
