<template>
  <div class="p-6">
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-xl font-semibold text-white">Automation Schedules</h1>
        <HelpTip position="bottom">
          Schedules use cron expressions to trigger actions automatically. A schedule can drive a fertigation program (auto-feed) or generate tasks (reminders).
          The automation worker checks active schedules and fires actuator events or creates tasks on the defined cadence.
        </HelpTip>
      <div class="flex items-center gap-3">
        <button class="px-3 py-1.5 text-xs rounded bg-gr33n-600 hover:bg-gr33n-500 text-white font-medium" @click="openCreate">+ New Schedule</button>
        <button class="text-xs text-zinc-400 hover:text-zinc-200" @click="refreshAll">Refresh</button>
      </div>
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
                <div class="flex flex-wrap gap-2 mt-1.5">
                  <router-link v-if="scheduleProgram(s.id)"
                    :to="{ path: '/fertigation', query: { tab: 'programs' } }"
                    class="text-[11px] px-1.5 py-0.5 rounded bg-green-900/40 text-green-400 border border-green-800/50 hover:bg-green-900/70">
                    Program: {{ scheduleProgram(s.id).name }}
                  </router-link>
                  <router-link v-if="scheduleTasks(s.id).length"
                    to="/tasks"
                    class="text-[11px] px-1.5 py-0.5 rounded bg-blue-900/40 text-blue-400 border border-blue-800/50 hover:bg-blue-900/70">
                    {{ scheduleTasks(s.id).length }} task{{ scheduleTasks(s.id).length > 1 ? 's' : '' }}
                  </router-link>
                </div>
              </div>
              <div class="flex items-center gap-2">
                <button @click="openEdit(s)" class="px-2 py-1 text-xs rounded border border-zinc-700 text-zinc-400 hover:text-zinc-200" title="Edit">&#9998;</button>
                <button @click="confirmDelete(s)" class="px-2 py-1 text-xs rounded border border-red-900/50 text-red-400 hover:text-red-300" title="Delete">&#128465;</button>
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
            <!-- Linked tasks -->
            <div v-if="scheduleTasks(s.id).length" class="mt-2 ml-0.5">
              <div class="space-y-1">
                <div v-for="t in scheduleTasks(s.id)" :key="t.id"
                  class="flex items-center gap-2 text-xs">
                  <span class="capitalize px-1.5 py-0.5 rounded text-[10px]"
                    :class="t.status === 'in_progress' ? 'bg-blue-900/50 text-blue-300' : t.status === 'completed' ? 'bg-green-900/50 text-green-300' : 'bg-zinc-800 text-zinc-400'">
                    {{ t.status?.replace(/_/g, ' ') }}
                  </span>
                  <span class="text-zinc-300">{{ t.title }}</span>
                  <span v-if="t.zone_id" class="text-zinc-600">· {{ taskZoneName(t.zone_id) }}</span>
                </div>
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

    <!-- Create/Edit Modal -->
    <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60" @click.self="showModal = false">
      <div class="bg-zinc-900 border border-zinc-700 rounded-xl p-6 w-full max-w-md space-y-4">
        <h3 class="text-white font-semibold">{{ editingSchedule ? 'Edit Schedule' : 'New Schedule' }}</h3>
        <div class="space-y-3">
          <div>
            <label class="text-xs text-zinc-400 block mb-1">Name</label>
            <input v-model="form.name" class="w-full bg-zinc-950 border border-zinc-700 rounded px-3 py-1.5 text-sm text-white" />
          </div>
          <div>
            <label class="text-xs text-zinc-400 block mb-1">Description</label>
            <input v-model="form.description" class="w-full bg-zinc-950 border border-zinc-700 rounded px-3 py-1.5 text-sm text-white" />
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="text-xs text-zinc-400 block mb-1">Type</label>
              <select v-model="form.schedule_type" class="w-full bg-zinc-950 border border-zinc-700 rounded px-3 py-1.5 text-sm text-white">
                <option value="cron">cron</option>
                <option value="interval">interval</option>
                <option value="one_time">one_time</option>
              </select>
            </div>
            <div>
              <label class="text-xs text-zinc-400 block mb-1">Timezone</label>
              <input v-model="form.timezone" class="w-full bg-zinc-950 border border-zinc-700 rounded px-3 py-1.5 text-sm text-white" placeholder="UTC" />
            </div>
          </div>
          <div>
            <label class="text-xs text-zinc-400 block mb-1">Cron Expression</label>
            <input v-model="form.cron_expression" class="w-full bg-zinc-950 border border-zinc-700 rounded px-3 py-1.5 text-sm text-white font-mono" placeholder="0 6 * * *" />
          </div>
          <div class="flex items-center gap-2">
            <input type="checkbox" v-model="form.is_active" id="sched-active" class="rounded border-zinc-600" />
            <label for="sched-active" class="text-xs text-zinc-300">Active</label>
          </div>
        </div>
        <div v-if="formError" class="text-red-400 text-xs">{{ formError }}</div>
        <div class="flex justify-end gap-3 pt-2">
          <button @click="showModal = false" class="px-3 py-1.5 text-xs rounded border border-zinc-700 text-zinc-400 hover:text-zinc-200">Cancel</button>
          <button @click="saveSchedule" :disabled="saving" class="px-3 py-1.5 text-xs rounded bg-gr33n-600 hover:bg-gr33n-500 text-white font-medium disabled:opacity-50">
            {{ saving ? 'Saving…' : (editingSchedule ? 'Update' : 'Create') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Delete Confirmation -->
    <div v-if="deleteTarget" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60" @click.self="deleteTarget = null">
      <div class="bg-zinc-900 border border-zinc-700 rounded-xl p-6 w-full max-w-sm space-y-4">
        <h3 class="text-white font-semibold">Delete Schedule</h3>
        <p class="text-sm text-zinc-300">Delete <span class="text-white font-medium">{{ deleteTarget.name }}</span>? This will also remove linked automation runs and executable actions.</p>
        <div class="flex justify-end gap-3 pt-2">
          <button @click="deleteTarget = null" class="px-3 py-1.5 text-xs rounded border border-zinc-700 text-zinc-400 hover:text-zinc-200">Cancel</button>
          <button @click="doDelete" :disabled="saving" class="px-3 py-1.5 text-xs rounded bg-red-600 hover:bg-red-500 text-white font-medium disabled:opacity-50">
            {{ saving ? 'Deleting…' : 'Delete' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useFarmStore } from '../stores/farm'
import { useFarmContextStore } from '../stores/farmContext'
import HelpTip from '../components/HelpTip.vue'
import api from '../api'

const store = useFarmStore()
const farmContext = useFarmContextStore()
const schedules = ref([])
const runs = ref([])
const programs = ref([])
const tasks = ref([])
const worker = ref({ running: false, simulation_mode: false })
const loading = ref(false)
const expandedSchedule = ref(null)
const scheduleEvents = ref([])
const eventsLoading = ref(false)

const showModal = ref(false)
const editingSchedule = ref(null)
const saving = ref(false)
const formError = ref('')
const deleteTarget = ref(null)
const form = ref(emptyForm())

function emptyForm() {
  return { name: '', description: '', schedule_type: 'cron', cron_expression: '', timezone: 'UTC', is_active: true }
}

function openCreate() {
  editingSchedule.value = null
  form.value = emptyForm()
  formError.value = ''
  showModal.value = true
}

function openEdit(s) {
  editingSchedule.value = s
  form.value = {
    name: s.name,
    description: s.description || '',
    schedule_type: s.schedule_type,
    cron_expression: s.cron_expression,
    timezone: s.timezone,
    is_active: s.is_active,
  }
  formError.value = ''
  showModal.value = true
}

async function saveSchedule() {
  formError.value = ''
  if (!form.value.name || !form.value.cron_expression) {
    formError.value = 'Name and cron expression are required.'
    return
  }
  saving.value = true
  try {
    const payload = { ...form.value, description: form.value.description || null }
    if (editingSchedule.value) {
      const updated = await store.updateSchedule(editingSchedule.value.id, payload)
      const idx = schedules.value.findIndex(s => s.id === editingSchedule.value.id)
      if (idx >= 0) schedules.value[idx] = updated
    } else {
      const created = await store.createSchedule(farmContext.farmId, payload)
      schedules.value = [...schedules.value, created]
    }
    showModal.value = false
  } catch (e) {
    formError.value = e?.response?.data?.error || 'Failed to save schedule.'
  } finally {
    saving.value = false
  }
}

function confirmDelete(s) {
  deleteTarget.value = s
}

async function doDelete() {
  saving.value = true
  try {
    await store.deleteSchedule(deleteTarget.value.id)
    schedules.value = schedules.value.filter(s => s.id !== deleteTarget.value.id)
    deleteTarget.value = null
  } catch (e) {
    formError.value = e?.response?.data?.error || 'Failed to delete schedule.'
  } finally {
    saving.value = false
  }
}

function scheduleProgram(scheduleId) {
  return programs.value.find(p => p.schedule_id === scheduleId && p.is_active)
}
function scheduleTasks(scheduleId) {
  return tasks.value.filter(t => t.schedule_id === scheduleId)
}
function taskZoneName(zoneId) {
  return store.zones.find(z => z.id === zoneId)?.name || ''
}

async function refreshAll() {
  const fid = farmContext.farmId
  if (!store.zones.length && fid) await store.loadAll(fid)
  loading.value = true
  try {
    const [s, r, w, p] = await Promise.all([
      store.loadSchedules(fid),
      store.loadAutomationRuns(fid),
      api.get('/automation/worker/health'),
      store.loadFertigationPrograms(fid),
    ])
    schedules.value = s
    runs.value = r
    worker.value = w.data || { running: false, simulation_mode: false }
    programs.value = p
    await store.loadTasks(fid)
    tasks.value = store.tasks
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
  if (!t) return '\u2014'
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
