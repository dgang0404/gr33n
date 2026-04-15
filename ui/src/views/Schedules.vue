<template>
  <div class="p-6">
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-xl font-semibold text-white">Automation Schedules</h1>
      <button class="text-xs text-zinc-400 hover:text-zinc-200" @click="refreshAll">Refresh</button>
    </div>

    <div v-if="loading" class="text-zinc-400 text-sm">Loading schedules…</div>

    <div v-else class="space-y-6">
      <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
        <h2 class="text-white text-sm font-semibold mb-3">Schedules</h2>
        <p v-if="!schedules.length" class="text-zinc-500 text-sm">No schedules found.</p>
        <div v-else class="space-y-3">
          <div v-for="s in schedules" :key="s.id" class="bg-zinc-950 border border-zinc-800 rounded-lg p-3">
            <div class="flex items-start justify-between gap-3">
              <div class="flex-1 min-w-0">
                <p class="text-sm text-zinc-200 font-medium">{{ s.name }}</p>
                <p class="text-xs text-zinc-500 mt-1">{{ s.schedule_type }} · {{ s.cron_expression }} · {{ s.timezone }}</p>
                <p class="text-xs text-zinc-600 mt-1">Last trigger: {{ s.last_triggered_time || 'never' }}</p>
              </div>
              <div class="flex items-center gap-2">
                <button
                  @click="toggleEvents(s.id)"
                  class="px-2 py-1 text-xs rounded border border-zinc-700 text-zinc-400 hover:text-zinc-200"
                >
                  {{ expandedSchedule === s.id ? 'Hide Events' : 'Events' }}
                </button>
                <button
                  @click="toggleSchedule(s)"
                  class="px-2 py-1 text-xs rounded border"
                  :class="s.is_active ? 'border-green-700 text-green-400' : 'border-zinc-700 text-zinc-400'"
                >
                  {{ s.is_active ? 'Active' : 'Inactive' }}
                </button>
              </div>
            </div>
            <!-- Actuator event history for this schedule -->
            <div v-if="expandedSchedule === s.id" class="mt-3 border-t border-zinc-800 pt-3">
              <p v-if="eventsLoading" class="text-zinc-500 text-xs">Loading actuator events…</p>
              <p v-else-if="!scheduleEvents.length" class="text-zinc-600 text-xs">No actuator events triggered by this schedule.</p>
              <div v-else class="space-y-1.5 max-h-48 overflow-y-auto">
                <div v-for="(ev, i) in scheduleEvents" :key="i"
                  class="flex items-center gap-3 text-xs bg-zinc-900 rounded px-2 py-1.5">
                  <span class="text-zinc-500 shrink-0 w-36">{{ formatTime(ev.event_time) }}</span>
                  <span class="text-zinc-300">{{ ev.command_sent || '—' }}</span>
                  <span class="px-1.5 py-0.5 rounded text-[10px]"
                    :class="execStatusClass(ev.execution_status)">
                    {{ ev.execution_status?.gr33ncore_actuator_execution_status_enum || 'pending' }}
                  </span>
                  <span class="text-zinc-600 ml-auto">actuator #{{ ev.actuator_id }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
        <div class="flex items-center justify-between mb-3">
          <h2 class="text-white text-sm font-semibold">Automation Runs</h2>
          <div class="text-xs" :class="worker.running ? 'text-green-400' : 'text-zinc-500'">
            Worker: {{ worker.running ? 'running' : 'stopped' }}
          </div>
        </div>
        <p class="text-zinc-600 text-xs mb-3">
          Simulation mode: {{ worker.simulation_mode ? 'enabled' : 'disabled' }}
        </p>
        <p v-if="!runs.length" class="text-zinc-500 text-sm">No automation runs yet.</p>
        <div v-else class="space-y-2">
          <div v-for="r in runs" :key="r.id" class="bg-zinc-950 border border-zinc-800 rounded-lg p-3">
            <div class="flex items-center justify-between gap-2">
              <div class="flex items-center gap-2">
                <span class="text-xs px-2 py-0.5 rounded"
                  :class="statusClass(r.status)">{{ r.status }}</span>
                <span v-if="r.schedule_id" class="text-[10px] text-zinc-600">
                  schedule #{{ r.schedule_id }}
                </span>
              </div>
              <span class="text-xs text-zinc-600">{{ formatTime(r.executed_at) }}</span>
            </div>
            <p v-if="r.message" class="text-zinc-300 text-xs mt-2">{{ r.message }}</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useFarmStore } from '../stores/farm'
import api from '../api'

const store = useFarmStore()
const schedules = ref([])
const runs = ref([])
const worker = ref({ running: false, simulation_mode: false })
const loading = ref(false)
const expandedSchedule = ref(null)
const scheduleEvents = ref([])
const eventsLoading = ref(false)

async function refreshAll() {
  if (!store.zones.length) await store.loadAll()
  loading.value = true
  try {
    const [s, r, w] = await Promise.all([
      store.loadSchedules(),
      store.loadAutomationRuns(),
      api.get('/automation/worker/health'),
    ])
    schedules.value = s
    runs.value = r
    worker.value = w.data || { running: false, simulation_mode: false }
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  try { await refreshAll() }
  finally { loading.value = false }
})

async function toggleSchedule(schedule) {
  const updated = await store.updateScheduleActive(schedule.id, !schedule.is_active)
  const idx = schedules.value.findIndex(s => s.id === schedule.id)
  if (idx >= 0) schedules.value[idx] = updated
}

async function toggleEvents(scheduleId) {
  if (expandedSchedule.value === scheduleId) {
    expandedSchedule.value = null
    scheduleEvents.value = []
    return
  }
  expandedSchedule.value = scheduleId
  eventsLoading.value = true
  try {
    scheduleEvents.value = await store.loadActuatorEventsBySchedule(scheduleId)
  } finally {
    eventsLoading.value = false
  }
}

function formatTime(t) {
  if (!t) return '—'
  return new Date(t).toLocaleString(undefined, {
    month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit', second: '2-digit'
  })
}

const STATUS_MAP = {
  success: 'bg-green-900/50 text-green-300',
  partial_success: 'bg-yellow-900/50 text-yellow-300',
  failed: 'bg-red-900/50 text-red-300',
  skipped: 'bg-zinc-800 text-zinc-300',
}
function statusClass(status) { return STATUS_MAP[status] ?? STATUS_MAP.skipped }

function execStatusClass(es) {
  const v = es?.gr33ncore_actuator_execution_status_enum || ''
  if (v.includes('success')) return 'bg-green-900/50 text-green-300'
  if (v.includes('error') || v.includes('failed')) return 'bg-red-900/50 text-red-300'
  return 'bg-zinc-800 text-zinc-400'
}
</script>
