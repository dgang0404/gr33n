<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col sm:flex-row sm:items-start justify-between gap-3">
      <div>
        <h1 class="text-xl font-semibold text-white">Targets &amp; schedules</h1>
        <p class="text-zinc-500 text-sm mt-1 max-w-2xl">
          Comfort bands per zone, what runs when, and automation toggles — without cron or JSON.
        </p>
      </div>
      <button
        type="button"
        class="text-xs text-zinc-400 hover:text-zinc-200 shrink-0"
        @click="refresh"
      >
        Refresh
      </button>
    </div>

    <nav class="flex flex-wrap gap-2 border-b border-zinc-800 pb-2" aria-label="Targets sections">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        type="button"
        class="text-xs px-3 py-1.5 rounded-t-lg border-b-2 -mb-[9px]"
        :class="activeTab === tab.id
          ? 'border-green-500 text-white'
          : 'border-transparent text-zinc-500 hover:text-zinc-300'"
        :data-test="`targets-tab-${tab.id}`"
        @click="selectTab(tab.id)"
      >
        {{ tab.label }}
      </button>
    </nav>

    <GuardianStarterChips :starters="activeStarters" />

    <ZoneContextBanner
      v-if="zoneContextId && activeTab === 'bands'"
      :zone-id="zoneContextId"
      :zone-name="zoneName(zoneContextId)"
      page-label="Comfort bands"
      back-to-zone-tab="air"
      :clear-route="{ path: '/comfort-targets' }"
    />

    <div v-if="!farmContext.farmId" class="text-zinc-400 text-sm">
      Select a farm to manage comfort targets.
    </div>

    <template v-else-if="activeTab === 'bands'">
      <div v-if="loading" class="text-zinc-400 text-sm">Loading comfort bands…</div>

      <EmptyStateHint
        v-else-if="!store.zones.length"
        reason="no_data"
        message="No zones yet — create zones first, then set comfort bands here or on each zone's Climate tab."
        action-label="My zones"
        action-to="/zones"
      />

      <EmptyStateHint
        v-else-if="zoneContextId && !filteredCards.length"
        reason="no_data"
        message="No comfort card for this zone filter."
        action-label="Show all zones"
        :action-to="{ path: '/comfort-targets' }"
      />

      <div v-else class="space-y-4">
        <article
          v-for="card in filteredCards"
          :key="card.zone.id"
          class="bg-zinc-900 border border-zinc-800 rounded-xl overflow-hidden"
          :data-test="`comfort-room-card-${card.zone.id}`"
        >
          <button
            type="button"
            class="w-full text-left p-4 hover:bg-zinc-900/80 transition-colors"
            @click="toggleExpanded(card.zone.id)"
          >
            <div class="flex items-start justify-between gap-3">
              <div>
                <p class="text-white font-medium">{{ card.zone.name }}</p>
                <p class="text-zinc-600 text-xs capitalize mt-0.5">{{ card.zone.zone_type || 'zone' }}</p>
                <p class="text-zinc-400 text-xs mt-2">{{ card.summaryLine }}</p>
              </div>
              <span
                class="text-[10px] px-2 py-0.5 rounded-full font-semibold shrink-0"
                :class="statusBadgeClass(card.status)"
                :data-test="`comfort-status-${card.zone.id}`"
              >
                {{ card.statusMeta.label }}
              </span>
            </div>
          </button>

          <div
            v-if="expandedZoneId === card.zone.id"
            class="border-t border-zinc-800 p-4 bg-zinc-950/40"
            :data-test="`comfort-editor-${card.zone.id}`"
          >
            <ComfortBandEditor
              need="air"
              :zone-id="card.zone.id"
              :farm-id="farmContext.farmId"
              :sensors="store.sensors"
              :setpoints="setpoints"
              :readings="store.readings"
              :sensor-types-filter="card.bands.map((b) => b.sensorType)"
              :empty-message="card.status === 'no_sensors'
                ? 'Add temperature or humidity sensors to this zone first.'
                : ''"
              @updated="onBandUpdated"
            />
            <router-link
              :to="{ path: `/zones/${card.zone.id}`, query: { tab: 'air' } }"
              class="inline-block text-[11px] text-green-600 hover:text-green-400 mt-3"
            >
              Open Climate tab →
            </router-link>
          </div>
        </article>
      </div>
    </template>

    <TargetsSchedulesPanel
      v-else-if="activeTab === 'schedules'"
      :zone-context-id="zoneContextId"
      @refresh="refresh"
    />

    <TargetsRulesPanel
      v-else-if="activeTab === 'rules'"
      :zone-context-id="zoneContextId"
      @refresh="refresh"
    />

    <footer class="border-t border-zinc-800 pt-4 flex flex-wrap items-center justify-between gap-3">
      <p class="text-zinc-600 text-xs">
        Cron strings, predicate JSON, and bulk CRUD live under Advanced power settings.
      </p>
      <router-link
        to="/setpoints"
        class="text-xs text-zinc-400 hover:text-green-400 border border-zinc-700 rounded-lg px-3 py-1.5"
        data-test="comfort-advanced-footer"
      >
        Advanced power settings →
      </router-link>
    </footer>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '../api'
import { useFarmStore } from '../stores/farm.js'
import { useFarmContextStore } from '../stores/farmContext.js'
import ZoneContextBanner from '../components/ZoneContextBanner.vue'
import EmptyStateHint from '../components/EmptyStateHint.vue'
import GuardianStarterChips from '../components/GuardianStarterChips.vue'
import ComfortBandEditor from '../components/ComfortBandEditor.vue'
import TargetsSchedulesPanel from '../components/TargetsSchedulesPanel.vue'
import TargetsRulesPanel from '../components/TargetsRulesPanel.vue'
import { parseZoneIdQuery, filterRulesForZone, filterSchedulesForZone } from '../lib/zoneContext.js'
import {
  buildComfortHubStarters,
  buildSchedulesFarmerStarters,
  buildRulesFarmerStarters,
} from '../lib/guardianStarters.js'
import {
  buildFarmComfortCards,
  filterComfortCardsByZone,
} from '../lib/farmComfortHub.js'

const route = useRoute()
const router = useRouter()
const store = useFarmStore()
const farmContext = useFarmContextStore()

const loading = ref(false)
const setpoints = ref([])
const rules = ref([])
const schedules = ref([])
const programs = ref([])
const cropCycles = ref([])
const expandedZoneId = ref(null)
const activeTab = ref('bands')

const tabs = [
  { id: 'bands', label: 'Comfort bands' },
  { id: 'schedules', label: 'What runs when' },
  { id: 'rules', label: 'Automation' },
]

const zoneContextId = computed(() => parseZoneIdQuery(route.query.zone_id))

const comfortCards = computed(() =>
  buildFarmComfortCards({
    zones: store.zones,
    sensors: store.sensors,
    setpoints: setpoints.value,
    readings: store.readings,
  }),
)

const filteredCards = computed(() =>
  filterComfortCardsByZone(comfortCards.value, zoneContextId.value),
)

const zoneScopedRules = computed(() => {
  if (!zoneContextId.value) return rules.value
  return filterRulesForZone(
    rules.value,
    zoneContextId.value,
    zoneName(zoneContextId.value),
    store.sensors,
  )
})

const zoneScopedSchedules = computed(() => {
  if (!zoneContextId.value) return schedules.value
  return filterSchedulesForZone(
    schedules.value,
    zoneContextId.value,
    zoneName(zoneContextId.value),
    programs.value,
    [],
    store.tasks,
  )
})

const activeStarters = computed(() => {
  const base = {
    zones: store.zones,
    zoneContextId: zoneContextId.value,
    zoneName: zoneContextId.value ? zoneName(zoneContextId.value) : '',
  }
  if (activeTab.value === 'schedules') {
    return buildSchedulesFarmerStarters({ ...base, schedules: zoneScopedSchedules.value })
  }
  if (activeTab.value === 'rules') {
    return buildRulesFarmerStarters({ ...base, rules: zoneScopedRules.value })
  }
  return buildComfortHubStarters({
    ...base,
    cards: filteredCards.value,
    rules: zoneScopedRules.value,
    programs: programs.value,
    schedules: zoneScopedSchedules.value,
    alerts: store.alerts,
    activeCycles: cropCycles.value,
    surface: zoneContextId.value ? 'comfort_hub_zone' : 'comfort_hub',
  })
})

watch(
  () => route.query.tab,
  (tab) => {
    if (tab === 'schedules' || tab === 'rules') activeTab.value = tab
    else activeTab.value = 'bands'
  },
  { immediate: true },
)

function selectTab(tabId) {
  activeTab.value = tabId
  const q = { ...route.query }
  if (tabId === 'bands') delete q.tab
  else q.tab = tabId
  router.replace({ query: q })
}

watch(zoneContextId, (id) => {
  if (id != null) expandedZoneId.value = id
})

function zoneName(zoneId) {
  return store.zones.find((z) => z.id === zoneId)?.name || `Zone ${zoneId}`
}

function statusBadgeClass(status) {
  if (status === 'ok') return 'bg-green-900/50 text-green-300'
  if (status === 'out_of_range') return 'bg-red-900/50 text-red-300'
  if (status === 'missing') return 'bg-amber-900/60 text-amber-200'
  return 'bg-zinc-800 text-zinc-500'
}

function toggleExpanded(zoneId) {
  expandedZoneId.value = expandedZoneId.value === zoneId ? null : zoneId
}

async function loadSetpoints() {
  const fid = farmContext.farmId
  if (!fid) return
  try {
    const params = zoneContextId.value ? { zone_id: zoneContextId.value } : {}
    const res = await api.get(`/farms/${fid}/setpoints`, { params })
    setpoints.value = res.data ?? []
  } catch {
    setpoints.value = []
  }
  return setpoints.value
}

async function onBandUpdated() {
  await loadSetpoints()
}

async function refresh() {
  const fid = farmContext.farmId
  if (!fid) return
  loading.value = true
  try {
    if (!store.zones.length) await store.loadAll(fid)
    else await store.refreshReadings(fid)
    const [sp, rs, sc, pr, cycles] = await Promise.all([
      loadSetpoints(),
      store.loadAutomationRules(fid).catch(() => []),
      store.loadSchedules(fid).catch(() => []),
      store.loadFertigationPrograms(fid).catch(() => []),
      api.get(`/farms/${fid}/crop-cycles`).then((r) => r.data ?? []).catch(() => []),
    ])
    void sp
    rules.value = rs
    schedules.value = sc
    programs.value = pr
    cropCycles.value = Array.isArray(cycles) ? cycles.filter((c) => c.is_active) : []
    await store.loadAlerts(fid).catch(() => {})
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  if (zoneContextId.value) expandedZoneId.value = zoneContextId.value
  await refresh()
})
</script>
