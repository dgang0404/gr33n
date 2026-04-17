<template>
  <div class="space-y-6">

    <!-- Farm header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
      <div>
        <h2 class="text-xl font-bold text-white">{{ store.farm?.name ?? 'Loading...' }}
          <HelpTip position="bottom">
            <strong>How it all connects:</strong> Your farm has <em>zones</em> (grow areas), each with <em>sensors</em>
            (reading temp/humidity/EC) and <em>actuators</em> (pumps, lights). <em>Schedules</em> trigger actuators on a
            cron cadence or generate <em>tasks</em>. <em>Fertigation programs</em> tie a schedule + reservoir + recipe +
            EC target into an automated feeding plan. <em>Crop cycles</em> track a single grow run per zone.
          </HelpTip>
        </h2>
        <p class="text-sm text-gray-500">{{ store.zones.length }} zones · {{ store.sensors.length }} sensors · {{ store.devices.length }} devices</p>
      </div>
      <button @click="refreshAll" class="text-xs text-gr33n-400 hover:text-gr33n-300 transition-colors self-start sm:self-auto">
        &#x21bb; Refresh
      </button>
    </div>

    <!-- Quick actions -->
    <section class="flex flex-wrap gap-3">
      <router-link to="/tasks?create=1"
        class="px-4 py-2 text-sm font-medium rounded-lg bg-green-900/50 text-green-400 border border-green-800 hover:bg-green-900/70 transition-colors">
        + New Task
      </router-link>
      <router-link :to="{ path: '/fertigation', query: { tab: 'mixing' } }"
        class="px-4 py-2 text-sm font-medium rounded-lg bg-blue-900/50 text-blue-400 border border-blue-800 hover:bg-blue-900/70 transition-colors">
        + Log Mix
      </router-link>
      <router-link to="/schedules"
        class="px-4 py-2 text-sm font-medium rounded-lg bg-zinc-800 text-zinc-300 border border-zinc-700 hover:bg-zinc-700 transition-colors">
        Schedules
      </router-link>
    </section>

    <!-- Today's Tasks + Alerts row -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">

      <!-- Today's Tasks -->
      <section class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
        <div class="flex items-center justify-between mb-3">
          <h3 class="text-xs font-semibold text-gray-500 uppercase tracking-widest">Today's Tasks</h3>
          <router-link to="/tasks" class="text-xs text-gr33n-500 hover:text-gr33n-400">View all &rarr;</router-link>
        </div>
        <div v-if="todayTasks.length" class="space-y-2">
          <div v-for="t in todayTasks" :key="t.id"
            class="flex items-center justify-between gap-3 bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2">
            <div class="flex items-center gap-2 min-w-0">
              <span class="text-[10px] px-1.5 py-0.5 rounded capitalize shrink-0"
                :class="taskStatusClass(t.status)">
                {{ t.status?.replace(/_/g, ' ') }}
              </span>
              <span class="text-sm text-zinc-200 truncate">{{ t.title }}</span>
            </div>
            <div class="flex items-center gap-2 shrink-0">
              <router-link v-if="t.zone_id" :to="`/zones/${t.zone_id}`"
                class="text-[11px] text-zinc-500 hover:text-green-400">{{ zoneName(t.zone_id) }}</router-link>
              <span v-if="t.due_date" class="text-[11px]"
                :class="isOverdue(t.due_date) ? 'text-red-400' : 'text-zinc-600'">
                {{ formatDueDate(t.due_date) }}
              </span>
            </div>
          </div>
        </div>
        <p v-else class="text-sm text-zinc-600">No tasks due today.</p>
      </section>

      <!-- Alerts -->
      <section class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
        <div class="flex items-center justify-between mb-3">
          <h3 class="text-xs font-semibold text-gray-500 uppercase tracking-widest">
            Alerts
            <span v-if="unreadAlerts > 0"
              class="ml-2 inline-flex items-center justify-center text-[10px] px-1.5 py-0.5 rounded-full bg-red-600 text-white font-bold">
              {{ unreadAlerts }}
            </span>
          </h3>
          <router-link to="/alerts" class="text-xs text-gr33n-500 hover:text-gr33n-400">View all &rarr;</router-link>
        </div>
        <div v-if="recentAlerts.length" class="space-y-2">
          <div v-for="a in recentAlerts" :key="a.id"
            class="flex items-center justify-between gap-3 bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2">
            <div class="flex items-center gap-2 min-w-0">
              <span class="text-[10px] px-1.5 py-0.5 rounded capitalize shrink-0"
                :class="alertSeverityClass(a.severity)">
                {{ a.severity || 'info' }}
              </span>
              <span class="text-sm text-zinc-200 truncate">{{ a.title || a.message }}</span>
            </div>
            <span class="text-[11px] text-zinc-600 shrink-0">{{ formatShort(a.created_at) }}</span>
          </div>
        </div>
        <p v-else class="text-sm text-zinc-600">No recent alerts.</p>
      </section>
    </div>

    <!-- Active Schedules + Recent Fertigation row -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">

      <!-- Active Schedules -->
      <section class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
        <div class="flex items-center justify-between mb-3">
          <h3 class="text-xs font-semibold text-gray-500 uppercase tracking-widest">Active Schedules</h3>
          <router-link to="/schedules" class="text-xs text-gr33n-500 hover:text-gr33n-400">Manage &rarr;</router-link>
        </div>
        <div v-if="activeSchedules.length" class="space-y-2">
          <div v-for="s in activeSchedules" :key="s.id"
            class="flex items-center justify-between gap-3 bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2">
            <div class="min-w-0">
              <p class="text-sm text-zinc-200 truncate">{{ s.name }}</p>
              <p class="text-[11px] text-zinc-500 font-mono">{{ s.cron_expression }} · {{ s.timezone }}</p>
            </div>
            <span class="text-[10px] px-1.5 py-0.5 rounded bg-green-900/50 text-green-300 shrink-0">active</span>
          </div>
        </div>
        <p v-else class="text-sm text-zinc-600">No active schedules.</p>
      </section>

      <!-- Recent Fertigation Events -->
      <section class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
        <div class="flex items-center justify-between mb-3">
          <h3 class="text-xs font-semibold text-gray-500 uppercase tracking-widest">Recent Fertigation</h3>
          <router-link :to="{ path: '/fertigation', query: { tab: 'events' } }" class="text-xs text-gr33n-500 hover:text-gr33n-400">View all &rarr;</router-link>
        </div>
        <div v-if="recentFertEvents.length" class="space-y-2">
          <div v-for="e in recentFertEvents" :key="e.id"
            class="flex items-center justify-between gap-3 bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2">
            <div class="flex items-center gap-2 min-w-0">
              <router-link v-if="e.zone_id" :to="`/zones/${e.zone_id}`"
                class="text-sm text-zinc-200 hover:text-green-400 truncate">{{ zoneName(e.zone_id) }}</router-link>
              <span v-if="e.program_id" class="text-[11px] text-zinc-500">{{ programName(e.program_id) }}</span>
            </div>
            <div class="flex items-center gap-2 shrink-0">
              <span class="text-[11px] text-zinc-500 font-mono">{{ e.volume_applied_liters || 0 }}L</span>
              <span class="text-[11px] text-zinc-600">{{ formatShort(e.applied_at) }}</span>
            </div>
          </div>
        </div>
        <p v-else class="text-sm text-zinc-600">No fertigation events yet.</p>
      </section>
    </div>

    <!-- Sensor tiles -->
    <section>
      <h3 class="text-xs font-semibold text-gray-500 uppercase tracking-widest mb-3">Live Sensors</h3>
      <div v-if="store.sensors.length" class="grid grid-cols-2 md:grid-cols-4 xl:grid-cols-7 gap-3">
        <SensorTile v-for="s in store.sensors" :key="s.id"
          :sensor="s" :reading="store.readings[s.id]" />
      </div>
      <div v-else class="text-sm text-gray-600">No sensors found for this farm.</div>
    </section>

    <!-- Zone cards -->
    <section>
      <h3 class="text-xs font-semibold text-gray-500 uppercase tracking-widest mb-3">Zones</h3>
      <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
        <div v-for="zone in store.zones" :key="zone.id" class="card space-y-3">
          <div class="flex items-center justify-between">
            <span class="font-semibold text-white">{{ zone.name }}</span>
            <span class="text-xs text-gray-500">{{ zone.zone_type }}</span>
          </div>
          <div class="grid grid-cols-2 gap-2">
            <SensorTile v-for="s in store.sensorsByZone(zone.id)" :key="s.id"
              :sensor="s" :reading="store.readings[s.id]" />
          </div>
          <div class="space-y-2 pt-1 border-t border-gray-800">
            <ActuatorCard v-for="d in store.devicesByZone(zone.id)" :key="d.id" :device="d" />
          </div>
          <div v-if="!store.devicesByZone(zone.id).length && !store.sensorsByZone(zone.id).length"
            class="text-xs text-gray-600">No devices assigned to this zone yet.</div>
        </div>
        <div v-if="!store.zones.length" class="text-sm text-gray-600">No zones found.</div>
      </div>
    </section>

    <!-- Quick actuator panel -->
    <section>
      <h3 class="text-xs font-semibold text-gray-500 uppercase tracking-widest mb-3">All Actuators</h3>
      <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-3">
        <ActuatorCard v-for="d in store.devices" :key="d.id" :device="d" />
        <div v-if="!store.devices.length" class="text-sm text-gray-600">No devices found.</div>
      </div>
    </section>

  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useFarmStore } from '../stores/farm'
import { useFarmContextStore } from '../stores/farmContext'
import SensorTile   from '../components/SensorTile.vue'
import ActuatorCard from '../components/ActuatorCard.vue'
import HelpTip from '../components/HelpTip.vue'

const store = useFarmStore()
const farmContext = useFarmContextStore()

const schedules = ref([])
const fertigationEvents = ref([])
const alerts = ref([])
const unreadAlerts = ref(0)
const programs = ref([])

const todayTasks = computed(() => {
  const today = new Date().toISOString().slice(0, 10)
  return store.tasks
    .filter(t => {
      if (t.status === 'completed' || t.status === 'cancelled') return false
      if (!t.due_date) return false
      const dd = String(t.due_date).slice(0, 10)
      return dd <= today
    })
    .slice(0, 8)
})

const activeSchedules = computed(() =>
  schedules.value.filter(s => s.is_active).slice(0, 6)
)

const recentFertEvents = computed(() =>
  [...fertigationEvents.value]
    .sort((a, b) => new Date(b.applied_at) - new Date(a.applied_at))
    .slice(0, 5)
)

const recentAlerts = computed(() =>
  [...alerts.value]
    .sort((a, b) => new Date(b.created_at) - new Date(a.created_at))
    .slice(0, 4)
)

function zoneName(id) {
  return store.zones.find(z => z.id === id)?.name ?? `Zone ${id}`
}

function programName(id) {
  return programs.value.find(p => p.id === id)?.name ?? ''
}

function isOverdue(due) {
  const today = new Date().toISOString().slice(0, 10)
  return String(due).slice(0, 10) < today
}

function formatDueDate(d) {
  if (!d) return ''
  const s = String(d).slice(0, 10)
  const today = new Date().toISOString().slice(0, 10)
  if (s === today) return 'today'
  return s
}

function formatShort(ts) {
  if (!ts) return '\u2014'
  return new Date(ts).toLocaleString(undefined, { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
}

function taskStatusClass(status) {
  if (status === 'in_progress') return 'bg-blue-900/50 text-blue-300'
  if (status === 'completed') return 'bg-green-900/50 text-green-300'
  if (status === 'on_hold' || status === 'blocked_requires_input') return 'bg-yellow-900/50 text-yellow-300'
  return 'bg-zinc-800 text-zinc-400'
}

function alertSeverityClass(sev) {
  if (sev === 'critical' || sev === 'high') return 'bg-red-900/50 text-red-300'
  if (sev === 'warning' || sev === 'medium') return 'bg-yellow-900/50 text-yellow-300'
  return 'bg-zinc-800 text-zinc-400'
}

async function refreshAll() {
  const fid = farmContext.farmId
  if (!fid) return
  await store.loadAll(fid)
  const [sch, ev, al, unread, prog] = await Promise.all([
    store.loadSchedules(fid),
    store.loadFertigationEvents(fid),
    store.loadAlerts(fid),
    store.countUnreadAlerts(fid),
    store.loadFertigationPrograms(fid),
  ])
  schedules.value = sch
  fertigationEvents.value = ev
  alerts.value = al
  unreadAlerts.value = unread
  programs.value = prog
  await store.loadTasks(fid)
}

onMounted(() => refreshAll())
</script>
