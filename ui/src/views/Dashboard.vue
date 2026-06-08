<template>
  <div class="space-y-6">

    <!-- Farm header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
      <div>
        <h2 class="text-xl font-bold text-white">{{ store.farm?.name ?? 'Loading...' }}
          <HelpTip position="bottom">
            <strong>How it all connects:</strong> Your farm has <em>zones</em> (grow areas), each with <em>sensors</em>
            (reading temp, humidity, EC) and <em>controls</em> (pumps, lights, fans). <em>Feeding plans</em> say when each zone
            gets water and nutrients. <em>Automations</em> react to readings. <em>Tasks</em> are your daily to-do list.
            Open <router-link to="/operator-guide" class="text-gr33n-400 underline">Guide</router-link> for a suggested click path.
          </HelpTip>
        </h2>
        <p class="text-sm text-gray-500">{{ store.zones.length }} zones · {{ store.sensors.length }} sensors · {{ store.devices.length }} devices</p>
      </div>
      <button @click="refreshAll" class="text-xs text-gr33n-400 hover:text-gr33n-300 transition-colors self-start sm:self-auto">
        &#x21bb; Refresh
      </button>
    </div>

    <!-- Phase 44 WS5 — first-run getting started checklist -->
    <GettingStartedChecklist
      v-if="showFirstRunChecklist"
      :items="firstRunItems"
      :farm-id="farmContext.farmId"
      :starters="firstRunStarters"
      @dismiss="firstRunDismissed = true"
    />

    <!-- Phase 41 WS1 — morning cockpit -->
    <FarmMorningStrip :chips="morningChips" />

    <GuardianStarterChips :starters="dashboardOpsStarters" />

    <!-- Quick actions -->
    <section class="flex flex-wrap gap-3">
      <router-link v-nav-hint="'/tasks'" to="/tasks?create=1"
        class="px-4 py-2 text-sm font-medium rounded-lg bg-green-900/50 text-green-400 border border-green-800 hover:bg-green-900/70 transition-colors">
        + New Task
      </router-link>
      <router-link v-nav-hint="'/feeding'" to="/feeding"
        class="px-4 py-2 text-sm font-medium rounded-lg bg-blue-900/50 text-blue-400 border border-blue-800 hover:bg-blue-900/70 transition-colors">
        Feed &amp; water
      </router-link>
      <router-link v-nav-hint="{ path: '/fertigation' }" :to="{ path: '/fertigation', query: { tab: 'mixing' } }"
        class="px-4 py-2 text-sm font-medium rounded-lg bg-zinc-800 text-zinc-400 border border-zinc-700 hover:bg-zinc-700 transition-colors">
        Log mix (advanced)
      </router-link>
      <router-link v-nav-hint="'/operator-guide'" to="/operator-guide"
        class="px-4 py-2 text-sm font-medium rounded-lg bg-zinc-800 text-gr33n-400 border border-zinc-600 hover:bg-zinc-700 transition-colors">
        Operator guide
      </router-link>
    </section>

    <!-- Today's Tasks + Alerts row -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">

      <!-- Today's Tasks -->
      <section class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
        <div class="flex items-center justify-between mb-3">
          <h3 class="text-xs font-semibold text-gray-500 uppercase tracking-widest">Today's Tasks</h3>
          <router-link v-nav-hint="'/tasks'" to="/tasks" class="text-xs text-gr33n-500 hover:text-gr33n-400">View all &rarr;</router-link>
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
        <EmptyStateHint
          v-else
          reason="no_data"
          message="No tasks due today."
          action-label="Open Tasks"
          action-to="/tasks"
        />
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
          <router-link v-nav-hint="'/alerts'" to="/alerts" class="text-xs text-gr33n-500 hover:text-gr33n-400">View all &rarr;</router-link>
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
        <EmptyStateHint
          v-else
          reason="automation_off"
          message="No recent alerts — thresholds and failed runs create them when rules are active."
          action-label="Automations"
          action-to="/automation"
        />
      </section>
    </div>

    <!-- What runs when + Recent feeds row -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">

      <!-- What runs when -->
      <section class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
        <div class="flex items-center justify-between mb-3">
          <h3 class="text-xs font-semibold text-gray-500 uppercase tracking-widest">What runs when</h3>
          <router-link v-nav-hint="'/schedules'" to="/schedules" class="text-xs text-gr33n-500 hover:text-gr33n-400">Farm-wide timing →</router-link>
        </div>
        <div v-if="activeSchedules.length" class="space-y-2">
          <div v-for="s in activeSchedules" :key="s.id"
            class="flex items-center justify-between gap-3 bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2">
            <div class="min-w-0">
              <p class="text-sm text-zinc-200 truncate">{{ s.name }}</p>
              <p class="text-[11px] text-zinc-500">{{ scheduleLabel(s) }}</p>
            </div>
            <span class="text-[10px] px-1.5 py-0.5 rounded bg-green-900/50 text-green-300 shrink-0">On</span>
          </div>
        </div>
        <EmptyStateHint
          v-else
          reason="automation_off"
          message="Nothing timed yet — feeding plans and lights need a daily time to run."
          action-label="Feed &amp; water"
          action-to="/feeding"
        />
      </section>

      <!-- Recent feeds -->
      <section class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
        <div class="flex items-center justify-between mb-3">
          <h3 class="text-xs font-semibold text-gray-500 uppercase tracking-widest">Recent feeds</h3>
          <router-link v-nav-hint="'/feeding'" :to="{ path: '/feeding' }" class="text-xs text-gr33n-500 hover:text-gr33n-400">Feed &amp; water →</router-link>
        </div>
        <div v-if="recentFertEvents.length" class="space-y-2">
          <div v-for="e in recentFertEvents" :key="e.id"
            class="flex items-center justify-between gap-3 bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2">
            <div class="flex items-center gap-2 min-w-0">
              <router-link v-if="e.zone_id" :to="{ path: `/zones/${e.zone_id}`, query: { tab: 'water' } }"
                class="text-sm text-zinc-200 hover:text-green-400 truncate">{{ zoneName(e.zone_id) }}</router-link>
              <span v-if="e.program_id" class="text-[11px] text-zinc-500">{{ programName(e.program_id) }}</span>
            </div>
            <div class="flex items-center gap-2 shrink-0">
              <span class="text-[11px] text-zinc-500 font-mono">{{ e.volume_applied_liters || 0 }}L</span>
              <span class="text-[11px] text-zinc-600">{{ formatShort(e.applied_at) }}</span>
            </div>
          </div>
        </div>
        <EmptyStateHint
          v-else
          reason="no_data"
          message="No feeds logged yet — they appear after programs run or you log a feed from a zone's Water tab."
          action-label="Feed &amp; water"
          :action-to="{ path: '/feeding' }"
        />
      </section>
    </div>

    <!-- Sensor tiles -->
    <section>
      <h3 class="text-xs font-semibold text-gray-500 uppercase tracking-widest mb-3">Live Sensors</h3>
      <div v-if="store.sensors.length" class="grid grid-cols-2 md:grid-cols-4 xl:grid-cols-7 gap-3">
        <SensorTile v-for="s in store.sensors" :key="s.id"
          :sensor="s" :reading="store.readings[s.id]" />
      </div>
      <EmptyStateHint
        v-else
        reason="no_telemetry"
        message="No sensors found for this farm — readings appear once hardware is registered and posting."
      />
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
        <div v-if="!store.zones.length" class="text-sm text-gray-600">
          No zones found.
          <p class="text-xs text-zinc-500 mt-1"><router-link v-nav-hint="'/zones'" class="text-gr33n-500 hover:underline" to="/zones">Create zones</router-link> first — sensors and actuators attach to them.</p>
        </div>
      </div>
    </section>

    <!-- Quick actuator panel -->
    <section>
      <h3 class="text-xs font-semibold text-gray-500 uppercase tracking-widest mb-3">All Actuators</h3>
      <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-3">
        <ActuatorCard v-for="d in store.devices" :key="d.id" :device="d" />
        <div v-if="!store.devices.length" class="text-sm text-gray-600">
          No devices found.
          <p class="text-xs text-zinc-500 mt-1">Register hardware under zones so actuators and sensors show up.</p>
        </div>
      </div>
    </section>

  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useFarmStore } from '../stores/farm'
import { useFarmContextStore } from '../stores/farmContext'
import SensorTile   from '../components/SensorTile.vue'
import ActuatorCard from '../components/ActuatorCard.vue'
import HelpTip from '../components/HelpTip.vue'
import FarmMorningStrip from '../components/FarmMorningStrip.vue'
import EmptyStateHint from '../components/EmptyStateHint.vue'
import GettingStartedChecklist from '../components/GettingStartedChecklist.vue'
import GuardianStarterChips from '../components/GuardianStarterChips.vue'
import { computeFarmMorningSnapshot } from '../lib/farmGrowSummary.js'
import { sumFarmPendingQueueDepth } from '../lib/farmQueueDepth.js'
import {
  computeFirstRunChecklist,
  shouldShowFirstRunChecklist,
} from '../lib/firstRunChecklist.js'
import { buildDashboardOpsStarters, buildSetupStarters } from '../lib/guardianStarters.js'
import { filterLowStockAlerts, listLowStockBatches } from '../lib/suppliesHub.js'
import { computeMonthSummary } from '../lib/moneyHub.js'
import { scheduleRunsLabel } from '../lib/cronHumanize.js'

const store = useFarmStore()
const farmContext = useFarmContextStore()

const schedules = ref([])
const setpoints = ref([])
const fertigationEvents = ref([])
const alerts = ref([])
const firstRunDismissed = ref(false)
const unreadAlerts = ref(0)
const programs = ref([])
const queueDepth = ref(0)
const nfBatches = ref([])
const nfInputs = ref([])
const costTransactions = ref([])

const lowStockCount = computed(() =>
  listLowStockBatches(nfBatches.value, nfInputs.value).length,
)

const lowStockAlerts = computed(() => filterLowStockAlerts(alerts.value))

const firstRunItems = computed(() => computeFirstRunChecklist({
  zones: store.zones,
  devices: store.devices,
  setpoints: setpoints.value,
  schedules: schedules.value,
  farmId: farmContext.farmId,
}))

const showFirstRunChecklist = computed(() => {
  if (firstRunDismissed.value) return false
  return shouldShowFirstRunChecklist(farmContext.farmId, firstRunItems.value)
})

const firstRunStarters = computed(() => {
  if (!showFirstRunChecklist.value) return []
  return buildSetupStarters({
    surface: 'first_run_dashboard',
    farmId: farmContext.farmId,
    zoneCount: store.zones.length,
    zones: store.zones,
    unreadAlerts: alerts.value.filter((a) => !a.is_read),
    deviceOffline: store.devices.length > 0 && store.devices.some((d) => d.status !== 'online'),
  })
})

const dashboardOpsStarters = computed(() => buildDashboardOpsStarters({
  lowStockCount: lowStockCount.value,
  lowStockAlerts: lowStockAlerts.value,
}))

const monthExpenses = computed(() => computeMonthSummary(costTransactions.value).expenses)

const morningChips = computed(() =>
  computeFarmMorningSnapshot({
    tasks: store.tasks,
    alerts: alerts.value,
    schedules: schedules.value,
    devices: store.devices,
    zones: store.zones,
    programs: programs.value,
    queueDepth: queueDepth.value,
    lowStockCount: lowStockCount.value,
    monthExpenses: monthExpenses.value,
  }).chips,
)

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

function scheduleLabel(schedule) {
  const when = scheduleRunsLabel(schedule)
  const tz = schedule.timezone && schedule.timezone !== 'UTC' ? schedule.timezone : null
  return tz ? `${when} · ${tz}` : when
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
  const [sch, sp, ev, al, unread, prog, batches, inputs, costs] = await Promise.all([
    store.loadSchedules(fid),
    store.loadSetpoints(fid),
    store.loadFertigationEvents(fid),
    store.loadAlerts(fid),
    store.countUnreadAlerts(fid),
    store.loadFertigationPrograms(fid),
    store.loadNfBatches(fid),
    store.loadNfInputs(fid),
    store.loadCosts(fid, { limit: 100, offset: 0 }),
  ])
  schedules.value = sch
  setpoints.value = sp
  fertigationEvents.value = ev
  alerts.value = al
  unreadAlerts.value = unread
  programs.value = prog
  nfBatches.value = batches
  nfInputs.value = inputs
  costTransactions.value = costs
  await store.loadTasks(fid)
  queueDepth.value = await sumFarmPendingQueueDepth(store.devices)
}

watch(
  () => farmContext.farmId,
  () => {
    firstRunDismissed.value = false
  },
)

onMounted(() => refreshAll())
</script>
