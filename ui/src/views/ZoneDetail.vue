<template>
  <div class="p-6 space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <router-link to="/zones" class="text-xs text-zinc-500 hover:text-zinc-300">&larr; Back to zones</router-link>
        <h1 class="text-xl font-semibold text-white mt-1">{{ zone?.name || 'Zone' }}</h1>
        <p class="text-zinc-500 text-sm">{{ zone?.description || 'No description' }}</p>
      </div>
      <span :class="zoneBadge(zone?.zone_type)" class="text-xs font-medium px-2 py-1 rounded-full capitalize">
        {{ zone?.zone_type || 'unknown' }}
      </span>
    </div>

    <div v-if="!zone" class="text-zinc-500 text-sm">Zone not found.</div>

    <template v-else>
      <!-- KPI row -->
      <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
          <p class="text-zinc-400 text-xs mb-1">Sensors</p>
          <p class="text-white text-2xl font-semibold">{{ sensors.length }}</p>
        </div>
        <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
          <p class="text-zinc-400 text-xs mb-1">Actuators</p>
          <p class="text-white text-2xl font-semibold">{{ actuators.length }}</p>
        </div>
        <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
          <p class="text-zinc-400 text-xs mb-1">Active Program</p>
          <p class="text-white text-sm font-medium truncate">{{ activeProgram?.name || 'None' }}</p>
        </div>
        <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
          <p class="text-zinc-400 text-xs mb-1">Last Fertigation</p>
          <p class="text-white text-sm font-medium truncate">
            {{ latestEvent ? `${formatTime(latestEvent.applied_at)} · ${latestEvent.volume_applied_liters || '0'}L` : 'None' }}
          </p>
        </div>
      </div>

      <!-- Live Sensor Readings -->
      <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
        <h2 class="text-sm font-semibold text-white mb-3">Live Readings</h2>
        <p v-if="!sensors.length" class="text-zinc-500 text-sm">No sensors assigned to this zone.</p>
        <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-3">
          <SensorTile
            v-for="s in sensors" :key="s.id"
            :sensor="s"
            :reading="store.readings[s.id]"
          />
        </div>
      </div>

      <!-- Actuator Controls (Quick Actions) -->
      <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
        <h2 class="text-sm font-semibold text-white mb-3">Controls</h2>
        <p v-if="!actuators.length" class="text-zinc-500 text-sm">No actuators assigned to this zone.</p>
        <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
          <div
            v-for="a in actuators" :key="a.id"
            class="bg-zinc-950 border rounded-lg p-3 flex items-center justify-between gap-3 transition-colors"
            :class="a.current_state_text === 'online' ? 'border-green-800/70' : 'border-zinc-800'"
          >
            <div class="flex items-center gap-2 min-w-0">
              <span class="text-lg shrink-0">{{ actuatorIcon(a.actuator_type) }}</span>
              <div class="min-w-0">
                <p class="text-white text-sm font-medium truncate">{{ a.name }}</p>
                <p class="text-zinc-500 text-xs capitalize">{{ a.actuator_type }}</p>
              </div>
            </div>
            <button
              @click="toggleActuator(a)"
              :disabled="toggling[a.id]"
              class="relative shrink-0 w-11 h-6 rounded-full transition-colors focus:outline-none focus:ring-2 focus:ring-green-600 disabled:opacity-40"
              :class="a.current_state_text === 'online' ? 'bg-green-600' : 'bg-zinc-700'"
              :title="a.current_state_text === 'online' ? 'Turn off' : 'Turn on'"
            >
              <span
                class="absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full shadow transition-transform duration-200"
                :class="a.current_state_text === 'online' ? 'translate-x-5' : 'translate-x-0'"
              />
            </button>
          </div>
        </div>
      </div>

      <!-- Recent Actuator Events -->
      <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
        <div class="flex items-center justify-between mb-3">
          <h2 class="text-sm font-semibold text-white">Recent Actuator Events</h2>
          <button @click="loadEvents" class="text-xs text-zinc-500 hover:text-zinc-300">Refresh</button>
        </div>
        <p v-if="eventsLoading" class="text-zinc-500 text-sm">Loading events…</p>
        <p v-else-if="!actuatorEvents.length" class="text-zinc-500 text-sm">No recent events.</p>
        <div v-else class="space-y-2 max-h-64 overflow-y-auto">
          <div v-for="ev in actuatorEvents" :key="ev.event_time + '-' + ev.actuator_id"
            class="bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 flex items-center justify-between gap-2"
          >
            <div class="min-w-0">
              <p class="text-zinc-200 text-xs font-medium truncate">
                {{ actuatorName(ev.actuator_id) }}
                <span class="text-zinc-500">→ {{ ev.command_sent || 'unknown' }}</span>
              </p>
              <p class="text-zinc-600 text-xs">{{ ev.source }} · {{ formatTime(ev.event_time) }}</p>
            </div>
            <span class="shrink-0 text-xs px-2 py-0.5 rounded"
              :class="eventStatusClass(ev.execution_status)">
              {{ shortStatus(ev.execution_status) }}
            </span>
          </div>
        </div>
      </div>

      <!-- Fertigation Summary -->
      <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
        <h2 class="text-sm font-semibold text-white mb-3">Fertigation Summary</h2>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
          <div class="bg-zinc-950 border border-zinc-800 rounded-lg p-3">
            <p class="text-zinc-400 text-xs mb-1">Active Program</p>
            <p class="text-zinc-200 text-sm">{{ activeProgram?.name || 'None' }}</p>
            <p v-if="activeProgram" class="text-zinc-600 text-xs mt-1">
              Vol: {{ activeProgram.total_volume_liters || '—' }}L
            </p>
          </div>
          <div class="bg-zinc-950 border border-zinc-800 rounded-lg p-3">
            <p class="text-zinc-400 text-xs mb-1">Recent Events ({{ zoneEvents.length }})</p>
            <div v-if="!zoneEvents.length" class="text-zinc-500 text-sm">No events</div>
            <div v-else class="space-y-1 max-h-32 overflow-y-auto">
              <p v-for="e in zoneEvents.slice(0, 5)" :key="e.id"
                class="text-zinc-300 text-xs">
                {{ formatTime(e.applied_at) }} · {{ e.volume_applied_liters || '0' }}L
                <span v-if="e.ec_after_mscm" class="text-zinc-500">· EC {{ e.ec_after_mscm }}</span>
              </p>
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { useFarmStore } from '../stores/farm'
import SensorTile from '../components/SensorTile.vue'

const route = useRoute()
const store = useFarmStore()
const toggling = ref({})

const programs = ref([])
const events = ref([])
const actuatorEvents = ref([])
const eventsLoading = ref(false)

const farmId = computed(() => store.farm?.id || 1)
const zoneId = computed(() => Number(route.params.id))
const zone = computed(() => store.zones.find(z => z.id === zoneId.value))
const sensors = computed(() => store.sensorsByZone(zoneId.value))
const actuators = computed(() => store.actuatorsByZone(zoneId.value))
const activeProgram = computed(() =>
  programs.value.find(p => p.target_zone_id === zoneId.value && p.is_active)
)
const zoneEvents = computed(() =>
  events.value
    .filter(e => e.zone_id === zoneId.value)
    .sort((a, b) => new Date(b.applied_at) - new Date(a.applied_at))
)
const latestEvent = computed(() => zoneEvents.value[0] || null)

async function loadEvents() {
  eventsLoading.value = true
  try {
    const all = []
    for (const a of actuators.value) {
      const evts = await store.loadActuatorEvents(a.id, { limit: 10 })
      all.push(...evts)
    }
    all.sort((a, b) => new Date(b.event_time) - new Date(a.event_time))
    actuatorEvents.value = all.slice(0, 30)
  } catch { actuatorEvents.value = [] }
  finally { eventsLoading.value = false }
}

async function toggleActuator(a) {
  toggling.value[a.id] = true
  try {
    await store.toggleActuator(a.id, a.current_state_text || 'offline')
  } finally {
    toggling.value[a.id] = false
  }
}

onMounted(async () => {
  if (!store.zones.length) await store.loadAll(1)
  const fid = farmId.value
  const [p, e] = await Promise.all([
    store.loadFertigationPrograms(fid),
    store.loadFertigationEvents(fid),
  ])
  programs.value = p
  events.value = e
  await loadEvents()
})

function actuatorName(id) {
  return store.actuators.find(a => a.id === id)?.name || `Actuator ${id}`
}

const ACTUATOR_ICONS = {
  pump: '🔧', fan: '🌀', light: '💡', valve: '🚰',
  heater: '🔥', cooler: '❄️', humidifier: '💨', co2: '🫧',
  relay: '⚡', controller: '🖥', default: '⚙️'
}
function actuatorIcon(type) {
  if (!type) return ACTUATOR_ICONS.default
  const k = type.toLowerCase()
  for (const [n, i] of Object.entries(ACTUATOR_ICONS)) { if (k.includes(n)) return i }
  return ACTUATOR_ICONS.default
}

const BADGE = {
  indoor: 'bg-indigo-900/60 text-indigo-300',
  outdoor: 'bg-emerald-900/60 text-emerald-300',
  greenhouse: 'bg-green-900/60 text-green-300',
}
function zoneBadge(type) {
  if (!type) return 'bg-zinc-800 text-zinc-400'
  const k = String(type).toLowerCase()
  for (const [name, cls] of Object.entries(BADGE)) {
    if (k.includes(name)) return cls
  }
  return 'bg-zinc-800 text-zinc-400'
}

function formatTime(ts) {
  if (!ts) return '—'
  const d = new Date(ts)
  const mins = Math.floor((Date.now() - d) / 60000)
  if (mins < 1) return 'just now'
  if (mins < 60) return `${mins}m ago`
  const hrs = Math.floor(mins / 60)
  if (hrs < 24) return `${hrs}h ago`
  return d.toLocaleDateString()
}

const EVENT_STATUS = {
  execution_completed_success_on_device: 'bg-green-900/50 text-green-300',
  pending_confirmation_from_feedback: 'bg-yellow-900/50 text-yellow-300',
}
function eventStatusClass(s) {
  if (!s) return 'bg-zinc-800 text-zinc-400'
  return EVENT_STATUS[s] ?? 'bg-zinc-800 text-zinc-400'
}
function shortStatus(s) {
  if (!s) return '—'
  if (s.includes('success')) return 'OK'
  if (s.includes('pending')) return 'pending'
  if (s.includes('fail')) return 'failed'
  return s.split('_').slice(0, 2).join(' ')
}
</script>
