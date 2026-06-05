<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col sm:flex-row sm:items-start justify-between gap-3">
      <div>
        <h1 class="text-xl font-semibold text-white">Feed &amp; water</h1>
        <p class="text-zinc-500 text-sm mt-1 max-w-2xl">
          One card per room — next feed, last run, and plan status. Open a room to edit the plan on the Water tab.
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

    <GuardianStarterChips :starters="feedingStarters" />

    <ZoneContextBanner
      v-if="zoneContextId"
      :zone-id="zoneContextId"
      :zone-name="zoneName(zoneContextId)"
      page-label="Feed &amp; water"
      back-to-zone-tab="water"
      :clear-route="{ path: '/feeding' }"
    />

    <div v-if="loading" class="text-zinc-400 text-sm">Loading feeding plans…</div>

    <EmptyStateHint
      v-else-if="!store.zones.length"
      reason="no_data"
      message="No rooms yet — create zones first, then start a feeding plan on each room's Water tab."
      action-label="My rooms"
      action-to="/zones"
    />

    <EmptyStateHint
      v-else-if="zoneContextId && !filteredCards.length"
      reason="no_data"
      message="No feeding card for this room filter."
      action-label="Show all rooms"
      :action-to="{ path: '/feeding' }"
    />

    <div v-else class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
      <router-link
        v-for="card in filteredCards"
        :key="card.zone.id"
        :to="zoneWaterLink(card.zone.id)"
        class="bg-zinc-900 border border-zinc-800 rounded-xl p-4 hover:border-green-800/80 transition-colors block"
        :data-test="`feeding-room-card-${card.zone.id}`"
      >
        <div class="flex items-start justify-between gap-2 mb-2">
          <div>
            <p class="text-white font-medium">{{ card.zone.name }}</p>
            <p class="text-zinc-600 text-xs capitalize mt-0.5">{{ card.zone.zone_type || 'zone' }}</p>
          </div>
          <div class="flex flex-wrap gap-1 justify-end">
            <span
              v-if="card.plan.irrigationOnly"
              class="text-[10px] px-1.5 py-0.5 rounded-full bg-sky-900/50 text-sky-300 font-semibold"
            >Water only</span>
            <span
              v-if="card.attention?.level === 'warn'"
              class="text-[10px] px-1.5 py-0.5 rounded-full bg-amber-900/60 text-amber-200 font-semibold"
              :data-test="`feeding-attention-${card.zone.id}`"
            >{{ card.attention.label }}</span>
            <span
              v-else-if="card.attention?.level === 'muted'"
              class="text-[10px] px-1.5 py-0.5 rounded-full bg-zinc-800 text-zinc-500"
            >{{ card.attention.label }}</span>
          </div>
        </div>

        <p class="text-sm text-zinc-200 mb-2">{{ card.plan.statusLine }}</p>

        <dl class="grid grid-cols-2 gap-2 text-xs">
          <div>
            <dt class="text-zinc-600">Last feed</dt>
            <dd class="text-zinc-300 mt-0.5 line-clamp-2">{{ card.plan.lastEventSummary }}</dd>
          </div>
          <div>
            <dt class="text-zinc-600">Reservoir</dt>
            <dd class="text-zinc-300 mt-0.5">{{ card.plan.reservoirLabel }}</dd>
          </div>
        </dl>

        <p class="text-[11px] text-green-600 mt-3">Open Water tab →</p>
      </router-link>
    </div>

    <footer class="border-t border-zinc-800 pt-4 flex flex-wrap items-center justify-between gap-3">
      <p class="text-zinc-600 text-xs">
        Programs, reservoirs, EC targets, mixing log, and recipes live under Advanced feeding.
      </p>
      <router-link
        :to="advancedFeedingLink"
        class="text-xs text-zinc-400 hover:text-green-400 border border-zinc-700 rounded-lg px-3 py-1.5"
        data-test="feeding-advanced-footer"
      >
        Advanced feeding (technical) →
      </router-link>
    </footer>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { useFarmStore } from '../stores/farm.js'
import { useFarmContextStore } from '../stores/farmContext.js'
import ZoneContextBanner from '../components/ZoneContextBanner.vue'
import EmptyStateHint from '../components/EmptyStateHint.vue'
import GuardianStarterChips from '../components/GuardianStarterChips.vue'
import { parseZoneIdQuery } from '../lib/zoneContext.js'
import { buildFeedingHubStarters } from '../lib/guardianStarters.js'
import {
  buildFarmFeedingCards,
  filterFeedingCardsByZone,
} from '../lib/farmFeedingHub.js'

const route = useRoute()
const store = useFarmStore()
const farmContext = useFarmContextStore()

const loading = ref(false)
const programs = ref([])
const schedules = ref([])
const events = ref([])
const ecTargets = ref([])
const reservoirs = ref([])

const zoneContextId = computed(() => parseZoneIdQuery(route.query.zone_id))

const feedingCards = computed(() =>
  buildFarmFeedingCards({
    zones: store.zones,
    programs: programs.value,
    schedules: schedules.value,
    events: events.value,
    ecTargets: ecTargets.value,
    reservoirs: reservoirs.value,
  }),
)

const filteredCards = computed(() =>
  filterFeedingCardsByZone(feedingCards.value, zoneContextId.value),
)

const feedingStarters = computed(() =>
  buildFeedingHubStarters({
    zones: store.zones,
    zoneContextId: zoneContextId.value,
    zoneName: zoneContextId.value ? zoneName(zoneContextId.value) : '',
  }),
)

const advancedFeedingLink = computed(() => {
  if (zoneContextId.value) {
    return { path: '/fertigation', query: { tab: 'programs', zone_id: String(zoneContextId.value) } }
  }
  return { path: '/fertigation', query: { tab: 'programs' } }
})

function zoneName(zoneId) {
  return store.zones.find((z) => z.id === zoneId)?.name || `Zone ${zoneId}`
}

function zoneWaterLink(zoneId) {
  return { path: `/zones/${zoneId}`, query: { tab: 'water' } }
}

async function refresh() {
  const fid = farmContext.farmId
  if (!fid) return
  loading.value = true
  try {
    if (!store.zones.length) await store.loadAll(fid)
    const [p, s, ev, ec, res] = await Promise.all([
      store.loadFertigationPrograms(fid),
      store.loadSchedules(fid),
      store.loadFertigationEvents(fid),
      store.loadEcTargets(fid),
      store.loadReservoirs(fid),
    ])
    programs.value = p
    schedules.value = s
    events.value = ev
    ecTargets.value = ec
    reservoirs.value = res
  } finally {
    loading.value = false
  }
}

onMounted(refresh)
</script>
