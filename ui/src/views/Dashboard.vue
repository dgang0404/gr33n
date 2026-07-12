<template>
  <div class="space-y-5">

    <FarmTodayHeader
      :farm-name="store.farm?.name ?? ''"
      :zones="store.zones"
      :get-status="zoneVisualStatus"
      :tasks-today-count="todayTasks.length"
      :unread-alerts="unreadAlerts"
      :overdue-task-count="overdueTaskCount"
      :tasks-link="tasksViewAllLink"
      :alerts-link="alertsViewAllLink"
      :site-weather="siteWeather"
      @refresh="refreshAll"
      @filter-attention="todayZoneFilter = 'attention'"
    />

    <!-- Phase 168 — empty farm: canvas CTA + Guardian setup (no IT checklist) -->
    <section
      v-if="showEmptyFarmStarters"
      class="rounded-xl border border-dashed border-zinc-700 bg-zinc-900/40 px-4 py-4 space-y-3"
      data-test="dashboard-empty-farm-starters"
    >
      <p class="text-sm text-zinc-300">
        {{ emptyFarmHint }}
      </p>
      <GuardianStarterChips :starters="emptyFarmStarters" data-test="dashboard-setup-starters" />
    </section>

    <!-- Phase 166 — site layer + visual farm canvas -->
    <FarmSiteStrip
      :site-weather="siteWeather"
      :zones="store.zones"
      :sensors="store.sensors"
      :readings="store.readings"
      :reservoirs="reservoirs"
      :programs="programs"
      :actuators="store.actuators"
      :schedules="schedules"
      :crop-cycles="cropCycles"
      :devices="store.devices"
      :queue-depth="queueDepth"
      :tasks="store.tasks"
      :alerts="alerts"
      :fertigation-events="fertigationEvents"
    />

    <FarmTodayAttentionStrip
      v-if="store.zones.length"
      :zones="store.zones"
      :sensors="store.sensors"
      :readings="store.readings"
      :actuators="store.actuators"
      :tasks="store.tasks"
      :alerts="alerts"
      :schedules="schedules"
      :programs="programs"
      :crop-cycles="cropCycles"
      :fertigation-events="fertigationEvents"
      @select-zone="openZoneQuickActions"
    />

    <FarmTodayZoneFilterBar
      v-if="store.zones.length"
      v-model="todayZoneFilter"
      :zones="store.zones"
      :get-status="zoneVisualStatus"
    />

    <FarmCanvas
      class="hidden md:block"
      :farm-id="farmContext.farmId"
      :zones="filteredZones"
      :total-zone-count="store.zones.length"
      :filter-label="activeFilterLabel"
      :sensors="store.sensors"
      :readings="store.readings"
      :actuators="store.actuators"
      :tasks="store.tasks"
      :alerts="alerts"
      :schedules="schedules"
      :programs="programs"
      :crop-cycles="cropCycles"
      :fertigation-events="fertigationEvents"
      :background-url="store.layoutBackgroundBlobUrl"
      @select-zone="openZoneQuickActions"
    />

    <FarmZoneStack
      :zones="filteredZones"
      :total-zone-count="store.zones.length"
      :filter-label="activeFilterLabel"
      :sensors="store.sensors"
      :readings="store.readings"
      :actuators="store.actuators"
      :tasks="store.tasks"
      :alerts="alerts"
      :schedules="schedules"
      :programs="programs"
      :crop-cycles="cropCycles"
      :fertigation-events="fertigationEvents"
      @select-zone="openZoneQuickActions"
    />

    <ZoneQuickActions
      :open="quickActionsOpen"
      :zone="quickZone"
      :status="quickStatus"
      :farm-id="farmContext.farmId"
      :programs="programs"
      :actuators="store.actuators"
      :tasks="store.tasks"
      :alerts="alerts"
      :sensors="store.sensors"
      @close="closeZoneQuickActions"
      @refresh="refreshAll"
    />

    <FarmTodayActionBar
      v-if="store.zones.length"
      :feed-water-link="feedWaterDailyLink"
      :new-task-link="newTaskLink"
      :schedules-link="comfortSchedulesLink"
      :low-stock-count="lowStockCount"
    />

    <FarmTodayAskGr33n
      v-if="store.zones.length"
      :starters="curatedTodayAskStarters"
    />

    <!-- Power-user details (collapsed by default) -->
    <details class="bg-zinc-900 border border-zinc-800 rounded-xl p-4 group" data-test="dashboard-details">
      <summary class="text-xs font-semibold text-gray-500 uppercase tracking-widest cursor-pointer list-none flex items-center justify-between">
        <span>All the details</span>
        <span class="text-zinc-600 group-open:rotate-180 transition-transform">▾</span>
      </summary>

      <div class="mt-4 space-y-6">
        <section
          v-if="detailsGuardianStarters.length"
          class="space-y-2"
          data-test="dashboard-details-guardian"
        >
          <h3 class="text-xs font-semibold text-gray-500 uppercase tracking-widest">Ask gr33n</h3>
          <GuardianStarterChips :starters="detailsGuardianStarters" />
        </section>

        <!-- Quick actions -->
        <section class="flex flex-wrap gap-3">
          <router-link v-nav-hint="'/zones'" :to="newTaskLink"
            class="px-4 py-2 text-sm font-medium rounded-lg bg-green-900/50 text-green-400 border border-green-800 hover:bg-green-900/70 transition-colors">
            + New Task
          </router-link>
          <router-link v-nav-hint="'/zones'" :to="feedWaterDailyLink"
            class="px-4 py-2 text-sm font-medium rounded-lg bg-blue-900/50 text-blue-400 border border-blue-800 hover:bg-blue-900/70 transition-colors">
            Feed &amp; water
          </router-link>
          <router-link v-nav-hint="'/zones'" :to="feedWaterNutrientsLink"
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
          <router-link v-nav-hint="'/zones'" :to="tasksViewAllLink" class="text-xs text-gr33n-500 hover:text-gr33n-400">View all &rarr;</router-link>
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
              <router-link
                v-if="t.zone_id"
                v-nav-hint="`/zones/${t.zone_id}`"
                :to="zoneTaskLink(t.zone_id)"
                class="text-[11px] text-zinc-500 hover:text-green-400"
              >{{ zoneName(t.zone_id) }}</router-link>
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
          action-label="Open zone Ops"
          :action-to="tasksViewAllLink"
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
          <router-link v-nav-hint="'/zones'" :to="alertsViewAllLink" class="text-xs text-gr33n-500 hover:text-gr33n-400">View all &rarr;</router-link>
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
          :action-to="comfortAutomationsLink"
        />
      </section>
    </div>

    <!-- What runs when + Recent feeds row -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">

      <!-- What runs when -->
      <section class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
        <div class="flex items-center justify-between mb-3">
          <h3 class="text-xs font-semibold text-gray-500 uppercase tracking-widest">What runs when</h3>
          <router-link v-nav-hint="'/comfort-targets'" to="/comfort-targets?tab=schedules" class="text-xs text-gr33n-500 hover:text-gr33n-400">Farm-wide timing →</router-link>
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
          :action-to="feedWaterDailyLink"
        />
      </section>

      <!-- Recent feeds -->
      <section class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
        <div class="flex items-center justify-between mb-3">
          <h3 class="text-xs font-semibold text-gray-500 uppercase tracking-widest">Recent feeds</h3>
          <router-link v-nav-hint="'/zones'" :to="feedWaterDailyLink" class="text-xs text-gr33n-500 hover:text-gr33n-400">Feed &amp; water →</router-link>
        </div>
        <div v-if="recentFertEvents.length" class="space-y-2">
          <div v-for="e in recentFertEvents" :key="e.id"
            class="flex items-center justify-between gap-3 bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2">
            <div class="flex items-center gap-2 min-w-0">
              <router-link v-if="e.zone_id" v-nav-hint="'/zones'" :to="{ path: `/zones/${e.zone_id}`, query: { tab: 'water' } }"
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
          :action-to="feedWaterDailyLink"
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

    <!-- Zone cards removed — FarmCanvas is the hero (Phase 166) -->

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
    </details>

  </div>
</template>

<script setup>
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useFarmStore } from '../stores/farm'
import { useFarmContextStore } from '../stores/farmContext'
import SensorTile   from '../components/SensorTile.vue'
import ActuatorCard from '../components/ActuatorCard.vue'
import FarmTodayHeader from '../components/FarmTodayHeader.vue'
import FarmSiteStrip from '../components/FarmSiteStrip.vue'
import FarmTodayAttentionStrip from '../components/FarmTodayAttentionStrip.vue'
import FarmTodayZoneFilterBar from '../components/FarmTodayZoneFilterBar.vue'
import FarmCanvas from '../components/FarmCanvas.vue'
import FarmZoneStack from '../components/FarmZoneStack.vue'
import FarmTodayActionBar from '../components/FarmTodayActionBar.vue'
import FarmTodayAskGr33n from '../components/FarmTodayAskGr33n.vue'
import ZoneQuickActions from '../components/ZoneQuickActions.vue'
import EmptyStateHint from '../components/EmptyStateHint.vue'
import GuardianStarterChips from '../components/GuardianStarterChips.vue'
import { sumFarmPendingQueueDepth } from '../lib/farmQueueDepth.js'
import {
  buildDashboardOpsStarters,
  buildMorningWalkthroughStarters,
  buildSetupStarters,
  buildTodayAttentionStarters,
  buildWeatherStarters,
} from '../lib/guardianStarters.js'
import { fetchSiteWeather } from '../lib/siteWeather.js'
import { filterLowStockAlerts, listLowStockBatches } from '../lib/suppliesHub.js'
import { scheduleRunsLabel } from '../lib/cronHumanize.js'
import {
  alertsViewAllRoute,
  comfortRoute,
  feedWaterRoute,
  newTaskRoute,
  tasksViewAllRoute,
  zoneOpsRoute,
} from '../lib/dashboardWorkspaceLinks.js'
import { computeZoneVisualStatus } from '../lib/farmVisualStatus.js'
import {
  TODAY_ZONE_FILTERS,
  filterZonesForToday,
  readTodayZoneFilter,
  writeTodayZoneFilter,
} from '../lib/farmTodayZoneFilter.js'
import {
  buildCuratedTodayAskStarters,
  mergeTodayDetailsGuardianStarters,
  shouldOfferMorningCheckOnToday,
} from '../lib/farmTodayAskGr33n.js'

const store = useFarmStore()
const farmContext = useFarmContextStore()

const schedules = ref([])
const setpoints = ref([])
const fertigationEvents = ref([])
const alerts = ref([])
const unreadAlerts = ref(0)
const programs = ref([])
const queueDepth = ref(0)
const nfBatches = ref([])
const nfInputs = ref([])
const costTransactions = ref([])
const cropCycles = ref([])
const siteWeather = ref(null)
const reservoirs = ref([])
const quickActionsOpen = ref(false)
const quickZone = ref(null)
const quickStatus = ref(null)

const lowStockCount = computed(() =>
  listLowStockBatches(nfBatches.value, nfInputs.value).length,
)

const lowStockAlerts = computed(() => filterLowStockAlerts(alerts.value))

const showEmptyFarmStarters = computed(() =>
  store.zones.length === 0 || store.devices.length === 0,
)

const emptyFarmHint = computed(() => {
  if (!store.zones.length) return 'Add your first grow area to see your farm here.'
  if (!store.devices.length) return 'Connect your Pi or edge device when you are ready — zones work without hardware too.'
  return ''
})

const emptyFarmStarters = computed(() => {
  if (!showEmptyFarmStarters.value) return []
  return buildSetupStarters({
    surface: 'first_run_dashboard',
    farmId: farmContext.farmId,
    zoneCount: store.zones.length,
    zones: store.zones,
    unreadAlerts: alerts.value.filter((a) => !a.is_read),
    deviceOffline: store.devices.length > 0 && store.devices.some((d) => d.status !== 'online'),
  })
})

function zoneVisualStatus(zone) {
  return computeZoneVisualStatus({
    zone,
    sensors: store.sensors,
    readings: store.readings,
    actuators: store.actuators,
    tasks: store.tasks,
    alerts: alerts.value,
    schedules: schedules.value,
    programs: programs.value,
    cropCycles: cropCycles.value,
    fertigationEvents: fertigationEvents.value,
  })
}

// Phase 173 — large-farm zone filter (chip bar only shows for ≥9 zones)
const todayZoneFilter = ref(readTodayZoneFilter())
watch(todayZoneFilter, (id) => writeTodayZoneFilter(id))

const filteredZones = computed(() =>
  filterZonesForToday(store.zones, todayZoneFilter.value, zoneVisualStatus),
)

const activeFilterLabel = computed(() => {
  if (todayZoneFilter.value === 'all') return ''
  return TODAY_ZONE_FILTERS.find((f) => f.id === todayZoneFilter.value)?.label ?? ''
})

const attentionStarters = computed(() => buildTodayAttentionStarters({
  zones: store.zones,
  getStatus: zoneVisualStatus,
  farmId: farmContext.farmId,
  farmName: store.farm?.name || '',
}))

const morningWalkthroughStarters = computed(() => buildMorningWalkthroughStarters({
  surface: 'dashboard',
  farmName: store.farm?.name || '',
}))

const weatherStarters = computed(() => buildWeatherStarters({
  surface: 'dashboard',
  farmName: store.farm?.name || '',
}))

const dashboardOpsStarters = computed(() => buildDashboardOpsStarters({
  lowStockCount: lowStockCount.value,
  lowStockAlerts: lowStockAlerts.value,
}))

const curatedTodayAskStarters = computed(() => buildCuratedTodayAskStarters({
  morningStarters: morningWalkthroughStarters.value,
  showMorningCheck: shouldOfferMorningCheckOnToday(),
  farmName: store.farm?.name || '',
}))

const detailsGuardianStarters = computed(() => mergeTodayDetailsGuardianStarters(
  attentionStarters.value,
  morningWalkthroughStarters.value,
  weatherStarters.value,
  dashboardOpsStarters.value,
))

const feedWaterDailyLink = computed(() => feedWaterRoute(store.zones))
const feedWaterNutrientsLink = computed(() => feedWaterRoute(store.zones))
const comfortAutomationsLink = computed(() => comfortRoute('automations'))
const comfortSchedulesLink = computed(() => comfortRoute('schedules'))
const newTaskLink = computed(() => newTaskRoute(store.tasks, store.zones))
const tasksViewAllLink = computed(() => tasksViewAllRoute(store.tasks, store.zones))
const alertsViewAllLink = computed(() => alertsViewAllRoute(alerts.value, store.zones, store.sensors))

function zoneTaskLink(zoneId) {
  return zoneOpsRoute(zoneId, 'tasks')
}

function openZoneQuickActions(zone, status) {
  quickZone.value = zone
  quickStatus.value = status || computeZoneVisualStatus({
    zone,
    sensors: store.sensors,
    readings: store.readings,
    actuators: store.actuators,
    tasks: store.tasks,
    alerts: alerts.value,
    schedules: schedules.value,
    programs: programs.value,
    cropCycles: cropCycles.value,
    fertigationEvents: fertigationEvents.value,
  })
  quickActionsOpen.value = true
}

function closeZoneQuickActions() {
  quickActionsOpen.value = false
  quickZone.value = null
  quickStatus.value = null
}

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

const overdueTaskCount = computed(() =>
  todayTasks.value.filter((t) => isOverdue(t.due_date)).length,
)

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
  const [sch, sp, ev, al, unread, prog, batches, inputs, costs, cycles] = await Promise.all([
    store.loadSchedules(fid),
    store.loadSetpoints(fid),
    store.loadFertigationEvents(fid),
    store.loadAlerts(fid),
    store.countUnreadAlerts(fid),
    store.loadFertigationPrograms(fid),
    store.loadNfBatches(fid),
    store.loadNfInputs(fid),
    store.loadCosts(fid, { limit: 100, offset: 0 }),
    store.loadCropCycles(fid),
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
  cropCycles.value = cycles
  await store.loadTasks(fid)
  queueDepth.value = await sumFarmPendingQueueDepth(store.devices)
  await store.refreshReadings(fid)
  try {
    siteWeather.value = await fetchSiteWeather(fid)
  } catch {
    siteWeather.value = null
  }
  try {
    reservoirs.value = await store.loadReservoirs(fid)
  } catch {
    reservoirs.value = []
  }
  if (fid) {
    await store.loadLayoutBackground(fid)
  }
}


onMounted(() => refreshAll())

function syncDocumentTitle() {
  const name = store.farm?.name
  document.title = name ? `Today · ${name}` : 'Today'
}

watch(() => store.farm?.name, syncDocumentTitle, { immediate: true })

onUnmounted(() => {
  document.title = 'gr33n'
})
</script>
