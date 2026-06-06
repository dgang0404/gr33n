<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col sm:flex-row sm:items-start justify-between gap-3">
      <div>
        <h1 class="text-xl font-semibold text-white">Feeding (details)</h1>
        <p class="text-zinc-500 text-sm mt-1 max-w-2xl">
          Farm-wide feeding admin — programs, nutrient tanks, and strength targets. For daily “what runs next per room”, use Feed &amp; water.
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

    <div class="flex flex-wrap gap-2">
      <router-link
        :to="dailyFeedingLink"
        class="px-4 py-2 text-sm font-medium rounded-lg bg-blue-900/50 text-blue-400 border border-blue-800 hover:bg-blue-900/70 transition-colors"
      >
        Feed &amp; water (daily)
      </router-link>
      <router-link
        :to="logMixLink"
        class="px-4 py-2 text-sm font-medium rounded-lg bg-green-900/50 text-green-400 border border-green-800 hover:bg-green-900/70 transition-colors"
        data-test="feeding-admin-log-mix"
      >
        Log a mix
      </router-link>
    </div>

    <ZoneContextBanner
      v-if="zoneContextId"
      :zone-id="zoneContextId"
      :zone-name="zoneName(zoneContextId)"
      page-label="Feeding (details)"
      back-to-zone-tab="water"
      :clear-route="{ path: '/operations/feeding', query: { tab: activeTab } }"
    />

    <GuardianStarterChips :starters="feedingAdminStarters" />

    <div class="flex flex-wrap gap-1 bg-zinc-900 border border-zinc-800 rounded-lg p-1 w-fit">
      <button
        v-for="t in tabs"
        :key="t.id"
        type="button"
        class="px-4 py-2 text-sm rounded-md transition-colors font-medium"
        :class="activeTab === t.id ? 'bg-zinc-800 text-white' : 'text-zinc-400 hover:text-zinc-200'"
        @click="selectTab(t.id)"
      >
        {{ t.label }}
      </button>
    </div>

    <div v-if="loading" class="text-zinc-400 text-sm">Loading feeding admin…</div>

    <!-- Programs -->
    <template v-else-if="activeTab === 'programs'">
      <EmptyStateHint
        v-if="!programCards.length"
        reason="no_data"
        message="No feeding programs yet — add one in the full editor or start a plan from a room's Water tab."
        action-label="Open full editor"
        :action-to="technicalLink('programs')"
      />
      <div v-else class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
        <div
          v-for="card in programCards"
          :key="card.id"
          class="bg-zinc-900 border rounded-xl p-4"
          :class="card.isActive ? 'border-green-800/60' : 'border-zinc-800'"
          :data-test="`feeding-program-card-${card.id}`"
        >
          <div class="flex items-start justify-between gap-2 mb-2">
            <div>
              <p class="text-white font-medium">{{ card.zoneName }}</p>
              <p class="text-zinc-500 text-xs mt-0.5">{{ card.name }}</p>
            </div>
            <div class="flex flex-wrap gap-1 justify-end shrink-0">
              <span
                v-if="card.irrigationOnly"
                class="text-[10px] px-1.5 py-0.5 rounded-full bg-sky-900/50 text-sky-300 font-semibold"
              >Water only</span>
              <span
                v-if="card.isActive"
                class="text-[10px] px-1.5 py-0.5 rounded-full bg-green-900/50 text-green-300"
              >Active</span>
            </div>
          </div>
          <dl class="text-xs space-y-1">
            <div>
              <dt class="text-zinc-600">Next run</dt>
              <dd class="text-zinc-200">{{ card.nextRunLabel }}</dd>
            </div>
            <div v-if="card.volumeLiters != null">
              <dt class="text-zinc-600">Volume</dt>
              <dd class="text-zinc-300 font-mono">{{ card.volumeLiters }}L</dd>
            </div>
          </dl>
          <router-link
            v-if="card.zoneId"
            :to="{ path: `/zones/${card.zoneId}`, query: { tab: 'water' } }"
            class="inline-block mt-3 text-xs text-green-500 hover:text-green-400"
          >
            Open room Water tab →
          </router-link>
        </div>
      </div>
    </template>

    <!-- Reservoirs -->
    <template v-else-if="activeTab === 'reservoirs'">
      <EmptyStateHint
        v-if="!reservoirCards.length"
        reason="no_data"
        message="No nutrient tanks configured — add reservoirs in the full feeding editor."
        action-label="Open full editor"
        :action-to="technicalLink('reservoirs')"
      />
      <div v-else class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
        <div
          v-for="card in reservoirCards"
          :key="card.id"
          class="bg-zinc-900 border border-zinc-800 rounded-xl p-4 space-y-2"
          :data-test="`feeding-reservoir-card-${card.id}`"
        >
          <div class="flex items-center justify-between gap-2">
            <p class="text-white font-medium">{{ card.name }}</p>
            <span
              class="text-[10px] px-1.5 py-0.5 rounded-full font-semibold capitalize"
              :class="card.statusTone === 'ok' ? 'bg-green-900/50 text-green-300' : card.statusTone === 'warn' ? 'bg-amber-900/50 text-amber-200' : 'bg-zinc-800 text-zinc-400'"
            >{{ card.statusLabel }}</span>
          </div>
          <p class="text-zinc-600 text-xs">{{ card.zoneName }}</p>
          <div class="flex items-end gap-1">
            <span class="text-white text-lg font-mono">{{ card.currentLiters ?? 0 }}</span>
            <span class="text-zinc-500 text-sm mb-0.5">/ {{ card.capacityLiters ?? 0 }} L</span>
          </div>
          <div class="w-full bg-zinc-800 rounded-full h-2">
            <div
              class="h-2 rounded-full transition-all"
              :class="card.statusTone === 'warn' ? 'bg-amber-500' : 'bg-blue-500'"
              :style="{ width: `${card.fillPct}%` }"
            />
          </div>
          <p v-if="card.ec != null" class="text-zinc-500 text-xs">
            EC {{ card.ec }} mS/cm<span v-if="card.ph != null"> · pH {{ card.ph }}</span>
          </p>
        </div>
      </div>
    </template>

    <!-- EC targets -->
    <template v-else-if="activeTab === 'ec-targets'">
      <EmptyStateHint
        v-if="!ecTargetCards.length"
        reason="no_data"
        message="No nutrient strength targets yet — define EC ranges per crop stage in the full editor."
        action-label="Open full editor"
        :action-to="technicalLink('ec-targets')"
      />
      <div v-else class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
        <div
          v-for="card in ecTargetCards"
          :key="card.id"
          class="bg-zinc-900 border border-zinc-800 rounded-xl p-4"
          :data-test="`feeding-ec-card-${card.id}`"
        >
          <p class="text-white font-medium capitalize">{{ card.stageLabel }}</p>
          <p class="text-zinc-600 text-xs mt-0.5">{{ card.zoneName }}</p>
          <p class="text-sm text-zinc-200 font-mono mt-2">{{ card.ecRange }}</p>
          <p v-if="card.phRange" class="text-xs text-zinc-500 mt-1">{{ card.phRange }}</p>
          <p v-if="card.notes" class="text-xs text-zinc-600 mt-2 line-clamp-2">{{ card.notes }}</p>
        </div>
      </div>
    </template>

    <footer class="pt-2 border-t border-zinc-800">
      <router-link
        :to="technicalLink(activeTab === 'ec-targets' ? 'ec-targets' : activeTab)"
        class="text-xs text-zinc-400 hover:text-green-400"
        data-test="feeding-admin-technical-footer"
      >
        Full feeding editor (mixing log, crop cycles, bulk edit) →
      </router-link>
    </footer>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useFarmStore } from '../stores/farm.js'
import { useFarmContextStore } from '../stores/farmContext.js'
import ZoneContextBanner from '../components/ZoneContextBanner.vue'
import EmptyStateHint from '../components/EmptyStateHint.vue'
import GuardianStarterChips from '../components/GuardianStarterChips.vue'
import { parseZoneIdQuery, filterProgramsForZone } from '../lib/zoneContext.js'
import { buildFeedingAdminStarters } from '../lib/guardianStarters.js'
import {
  buildProgramAdminCards,
  buildReservoirAdminCards,
  buildEcTargetAdminCards,
  filterReservoirsForZone,
  filterEcTargetsForZone,
} from '../lib/feedingAdminHub.js'

const route = useRoute()
const router = useRouter()
const store = useFarmStore()
const farmContext = useFarmContextStore()

const loading = ref(false)
const programs = ref([])
const schedules = ref([])
const reservoirs = ref([])
const ecTargets = ref([])
const cropCycles = ref([])

const tabs = [
  { id: 'programs', label: 'Programs' },
  { id: 'reservoirs', label: 'Nutrient tanks' },
  { id: 'ec-targets', label: 'Strength targets' },
]

const activeTab = ref('programs')

const zoneContextId = computed(() => parseZoneIdQuery(route.query.zone_id))

const filteredPrograms = computed(() => {
  if (zoneContextId.value == null) return programs.value
  return filterProgramsForZone(programs.value, zoneContextId.value, cropCycles.value)
})

const filteredReservoirs = computed(() =>
  filterReservoirsForZone(reservoirs.value, zoneContextId.value),
)

const filteredEcTargets = computed(() =>
  filterEcTargetsForZone(ecTargets.value, zoneContextId.value),
)

const programCards = computed(() =>
  buildProgramAdminCards(filteredPrograms.value, store.zones, schedules.value),
)

const feedingAdminStarters = computed(() => buildFeedingAdminStarters({
  zones: store.zones,
  zoneContextId: zoneContextId.value,
  programs: programs.value,
}))

const reservoirCards = computed(() =>
  buildReservoirAdminCards(filteredReservoirs.value, store.zones),
)

const ecTargetCards = computed(() =>
  buildEcTargetAdminCards(filteredEcTargets.value, store.zones),
)

const dailyFeedingLink = computed(() => {
  const q = zoneContextId.value ? { zone_id: String(zoneContextId.value) } : {}
  return { path: '/feeding', query: q }
})

const logMixLink = computed(() => technicalLink('mixing'))

function technicalLink(tab) {
  const q = { tab }
  if (zoneContextId.value) q.zone_id = String(zoneContextId.value)
  return { path: '/fertigation', query: q }
}

function zoneName(zoneId) {
  return store.zones.find((z) => z.id === zoneId)?.name || `Zone ${zoneId}`
}

function tabFromQuery(query) {
  const raw = query.tab
  const s = Array.isArray(raw) ? raw[0] : raw
  if (s === 'mixing') return null
  if (s && tabs.some((t) => t.id === s)) return s
  return 'programs'
}

function selectTab(id) {
  activeTab.value = id
  const q = { ...route.query, tab: id }
  router.replace({ path: '/operations/feeding', query: q }).catch(() => {})
}

watch(
  () => route.fullPath,
  () => {
    if (route.name !== 'operations-feeding') return
    const tab = tabFromQuery(route.query)
    if (tab === null) {
      router.replace({ path: '/fertigation', query: route.query }).catch(() => {})
      return
    }
    activeTab.value = tab
  },
  { immediate: true },
)

async function refresh() {
  const fid = farmContext.farmId
  if (!fid) return
  loading.value = true
  try {
    if (!store.zones.length) await store.loadAll(fid)
    const [p, s, r, ec, cycles] = await Promise.all([
      store.loadFertigationPrograms(fid),
      store.loadSchedules(fid),
      store.loadReservoirs(fid),
      store.loadEcTargets(fid),
      store.loadCropCycles(fid),
    ])
    programs.value = p
    schedules.value = s
    reservoirs.value = r
    ecTargets.value = ec
    cropCycles.value = cycles
  } finally {
    loading.value = false
  }
}

onMounted(refresh)
</script>
