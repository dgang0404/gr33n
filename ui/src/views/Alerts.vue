<template>
  <div class="p-6 max-w-5xl">
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 mb-6">
      <h1 class="text-2xl font-bold text-green-400">Alerts</h1>
      <div class="flex items-center gap-3">
        <select v-model="severityFilter"
          class="bg-zinc-800 border border-zinc-700 text-gray-300 text-xs rounded-lg px-3 py-1.5 focus:outline-none focus:ring-1 focus:ring-gr33n-600">
          <option value="">All severities</option>
          <option value="critical">Critical</option>
          <option value="high">High</option>
          <option value="medium">Medium</option>
          <option value="low">Low</option>
        </select>
        <button @click="refresh"
          class="text-xs text-zinc-400 hover:text-white border border-zinc-700 rounded-lg px-3 py-1.5 transition-colors">
          Refresh
        </button>
      </div>
    </div>

    <div v-if="loading" class="text-zinc-500 text-sm">Loading alerts...</div>
    <div v-else-if="filtered.length === 0" class="text-zinc-500 text-sm bg-zinc-800 border border-zinc-700 rounded-xl p-8 text-center">
      No alerts{{ severityFilter ? ` with severity "${severityFilter}"` : '' }}.
    </div>

    <div v-else class="space-y-2">
      <div v-for="a in filtered" :key="a.id"
        class="bg-zinc-800 border border-zinc-700 rounded-xl p-4 flex flex-col sm:flex-row items-start gap-3 sm:gap-4"
        :class="{ 'opacity-60': a.is_acknowledged }">
        <span :class="severityBadge(a.severity)" class="mt-0.5 text-xs font-bold px-2 py-0.5 rounded uppercase shrink-0">
          {{ a.severity?.gr33ncore_notification_priority_enum || a.severity || 'medium' }}
        </span>
        <div class="flex-1 min-w-0">
          <p class="text-white text-sm font-medium truncate">{{ a.subject_rendered || 'Alert' }}</p>
          <p class="text-zinc-400 text-xs mt-0.5">{{ a.message_text_rendered }}</p>
          <p class="text-zinc-600 text-xs mt-1">{{ formatTime(a.created_at) }}</p>
          <div v-if="linkedTasks(a.id).length" class="mt-1 flex flex-wrap gap-1">
            <router-link
              v-for="t in linkedTasks(a.id)"
              :key="t.id"
              to="/tasks"
              class="text-[11px] px-2 py-0.5 rounded bg-green-900/40 border border-green-800 text-green-300 hover:bg-green-900/60"
            >
              → Task #{{ t.id }}
            </router-link>
          </div>
        </div>
        <div class="flex items-center gap-2 shrink-0 self-end sm:self-auto">
          <button @click="openCreateTask(a)"
            class="text-xs text-blue-300 hover:text-blue-200 border border-blue-800 rounded px-2 py-1 transition-colors">
            Create task
          </button>
          <span v-if="a.is_read" class="text-zinc-600 text-xs">Read</span>
          <button v-else @click="markRead(a.id)"
            class="text-xs text-zinc-400 hover:text-white border border-zinc-700 rounded px-2 py-1 transition-colors">
            Mark read
          </button>
          <span v-if="a.is_acknowledged" class="text-green-600 text-xs font-medium">ACK</span>
          <button v-else @click="acknowledge(a.id)"
            class="text-xs text-green-500 hover:text-green-300 border border-green-800 rounded px-2 py-1 transition-colors">
            Acknowledge
          </button>
        </div>
      </div>
    </div>

    <div v-if="!loading && filtered.length >= 50" class="mt-4 text-center">
      <button @click="loadMore"
        class="text-xs text-zinc-400 hover:text-white border border-zinc-700 rounded-lg px-4 py-2 transition-colors">
        Load more
      </button>
    </div>

    <!-- Create-task modal -->
    <div
      v-if="createTaskAlert"
      class="fixed inset-0 z-50 bg-black/70 p-4 flex items-center justify-center"
      @click.self="cancelCreateTask"
    >
      <div class="bg-zinc-900 border border-zinc-700 rounded-xl p-5 w-full max-w-md space-y-3">
        <div class="flex items-center justify-between">
          <h2 class="text-sm font-semibold text-white">Create task from alert #{{ createTaskAlert.id }}</h2>
          <button class="text-xs text-zinc-500 hover:text-zinc-200" @click="cancelCreateTask">Close</button>
        </div>
        <p class="text-[11px] text-zinc-500">
          Prefilled from the alert. Anything you leave as-is is derived server-side.
        </p>
        <div>
          <label class="block text-xs text-zinc-500 mb-1">Title</label>
          <input v-model="taskForm.title" type="text"
            class="w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white" />
        </div>
        <div>
          <label class="block text-xs text-zinc-500 mb-1">Description</label>
          <textarea v-model="taskForm.description" rows="3"
            class="w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white" />
        </div>
        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="block text-xs text-zinc-500 mb-1">Priority</label>
            <select v-model.number="taskForm.priority"
              class="w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white">
              <option :value="0">Low</option>
              <option :value="1">Normal</option>
              <option :value="2">High</option>
              <option :value="3">Urgent</option>
            </select>
          </div>
          <div>
            <label class="block text-xs text-zinc-500 mb-1">Due date</label>
            <input v-model="taskForm.due_date" type="date"
              class="w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white" />
          </div>
        </div>
        <p v-if="createTaskError" class="text-xs text-red-400">{{ createTaskError }}</p>
        <div class="flex justify-end gap-2 pt-1">
          <button @click="cancelCreateTask" class="text-xs text-zinc-400 hover:text-zinc-200">Cancel</button>
          <button
            :disabled="creatingTask"
            @click="submitCreateTask"
            class="text-xs font-medium px-3 py-1.5 rounded-lg bg-green-700 text-white disabled:opacity-40"
          >
            {{ creatingTask ? 'Creating…' : 'Create task' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useFarmStore } from '../stores/farm'
import { useFarmContextStore } from '../stores/farmContext'

const farmStore = useFarmStore()
const farmContext = useFarmContextStore()
const loading = ref(false)
const severityFilter = ref('')
const offset = ref(0)

const filtered = computed(() => {
  if (!severityFilter.value) return farmStore.alerts
  return farmStore.alerts.filter(a => {
    const sev = a.severity?.gr33ncore_notification_priority_enum || a.severity || ''
    return sev === severityFilter.value
  })
})

function severityBadge(sev) {
  const s = sev?.gr33ncore_notification_priority_enum || sev || 'medium'
  return {
    critical: 'bg-red-900 text-red-300 border border-red-700',
    high:     'bg-orange-900 text-orange-300 border border-orange-700',
    medium:   'bg-yellow-900 text-yellow-300 border border-yellow-700',
    low:      'bg-zinc-700 text-zinc-300 border border-zinc-600',
  }[s] || 'bg-zinc-700 text-zinc-300'
}

function severityToPriority(sev) {
  const s = sev?.gr33ncore_notification_priority_enum || sev || 'medium'
  return { critical: 3, high: 2, medium: 1, low: 0 }[s] ?? 1
}

function formatTime(ts) {
  if (!ts) return ''
  const d = new Date(ts)
  return d.toLocaleString()
}

function linkedTasks(alertId) {
  return farmStore.tasks.filter((t) => Number(t.source_alert_id) === Number(alertId))
}

async function refresh() {
  if (!farmContext.farmId) return
  loading.value = true
  offset.value = 0
  try {
    await farmStore.loadAlerts(farmContext.farmId, { limit: 50, offset: 0 })
    await farmStore.countUnreadAlerts(farmContext.farmId)
    // Pull tasks too so we can render the "→ Task #N" badge without an extra round trip.
    await farmStore.loadTasks(farmContext.farmId)
  } finally {
    loading.value = false
  }
}

async function loadMore() {
  offset.value += 50
  const more = await farmStore.loadAlerts(farmContext.farmId, { limit: 50, offset: offset.value })
  if (more.length === 0) offset.value -= 50
}

async function markRead(id) {
  await farmStore.markAlertRead(id)
  await farmStore.countUnreadAlerts(farmContext.farmId)
}

async function acknowledge(id) {
  await farmStore.markAlertAcknowledged(id)
  await farmStore.countUnreadAlerts(farmContext.farmId)
}

// Create-task modal state
const createTaskAlert = ref(null)
const creatingTask = ref(false)
const createTaskError = ref('')
const taskForm = ref({ title: '', description: '', priority: 1, due_date: '' })

function openCreateTask(alert) {
  createTaskError.value = ''
  createTaskAlert.value = alert
  taskForm.value = {
    title: alert.subject_rendered || `Follow up on alert #${alert.id}`,
    description: alert.message_text_rendered || '',
    priority: severityToPriority(alert.severity),
    due_date: '',
  }
}

function cancelCreateTask() {
  createTaskAlert.value = null
  createTaskError.value = ''
}

async function submitCreateTask() {
  if (!createTaskAlert.value) return
  creatingTask.value = true
  createTaskError.value = ''
  try {
    const payload = {
      title: taskForm.value.title.trim() || undefined,
      description: taskForm.value.description.trim() ? taskForm.value.description.trim() : null,
      priority: taskForm.value.priority,
    }
    if (taskForm.value.due_date) payload.due_date = taskForm.value.due_date
    await farmStore.createTaskFromAlert(createTaskAlert.value.id, payload)
    cancelCreateTask()
  } catch (e) {
    createTaskError.value = e.response?.data?.error || e.message || 'Failed to create task'
  } finally {
    creatingTask.value = false
  }
}

onMounted(refresh)
watch(() => farmContext.farmId, refresh)
</script>
