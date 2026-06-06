<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col sm:flex-row sm:items-start justify-between gap-3">
      <div>
        <h1 class="text-xl font-semibold text-white">Supplies</h1>
        <p class="text-zinc-500 text-sm mt-1 max-w-2xl">
          What you have on hand, what is running low, and where to log a mix. Farm-wide stock — not tied to one zone.
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

    <ZoneContextBanner
      v-if="zoneContextId"
      :zone-id="zoneContextId"
      :zone-name="zoneName(zoneContextId)"
      page-label="Supplies"
      back-to-zone-tab="water"
      :clear-route="{ path: '/operations/supplies' }"
    />

    <GuardianStarterChips :starters="suppliesStarters" />

    <div
      v-if="lowStockRows.length"
      class="rounded-xl border border-amber-800/80 bg-amber-950/40 px-4 py-3 space-y-2"
      data-test="supplies-low-stock-banner"
    >
      <p class="text-sm font-medium text-amber-200">
        {{ lowStockRows.length }} supply batch{{ lowStockRows.length === 1 ? '' : 'es' }} below the low-stock threshold
      </p>
      <ul class="text-sm text-amber-100/90 space-y-1">
        <li v-for="row in lowStockRows.slice(0, 5)" :key="row.batch.id">
          <strong>{{ row.inputName }}</strong>
          — {{ formatQty(row.remaining) }} left (threshold {{ formatQty(row.threshold) }})
        </li>
      </ul>
      <p v-if="lowStockRows.length > 5" class="text-xs text-amber-300/80">
        + {{ lowStockRows.length - 5 }} more in the list below
      </p>
      <router-link
        v-if="lowStockAlertLink"
        :to="lowStockAlertLink"
        class="inline-block text-xs text-amber-300 hover:text-amber-100 underline"
      >
        View low-stock alert →
      </router-link>
    </div>

    <div class="flex flex-wrap gap-2">
      <router-link
        :to="logMixLink"
        class="px-4 py-2 text-sm font-medium rounded-lg bg-green-900/50 text-green-400 border border-green-800 hover:bg-green-900/70 transition-colors"
        data-test="supplies-log-mix"
      >
        Log a mix
      </router-link>
      <router-link
        :to="recipesLink"
        class="px-4 py-2 text-sm font-medium rounded-lg bg-zinc-800 text-zinc-300 border border-zinc-700 hover:bg-zinc-700 transition-colors"
      >
        Mixing recipes ({{ recipes.length }})
      </router-link>
    </div>

    <div v-if="loading" class="text-zinc-400 text-sm">Loading supplies…</div>

    <EmptyStateHint
      v-else-if="!supplyRows.length"
      reason="no_data"
      message="No supply batches yet — add inputs and batches in the full editor, or start from a demo farm."
      action-label="Open full editor"
      :action-to="{ path: '/inventory', query: { tab: 'batches' } }"
    />

    <div v-else class="space-y-3">
      <p class="text-xs text-zinc-500 uppercase tracking-widest">
        On hand — {{ supplyRows.length }} batch{{ supplyRows.length === 1 ? '' : 'es' }}
      </p>
      <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
        <div
          v-for="row in supplyRows"
          :key="row.id"
          class="bg-zinc-900 border rounded-xl p-4 transition-colors"
          :class="row.lowStock ? 'border-amber-800/80' : 'border-zinc-800'"
          :data-test="`supply-row-${row.id}`"
        >
          <div class="flex items-start justify-between gap-2 mb-2">
            <div>
              <p class="text-white font-medium">{{ row.inputName }}</p>
              <p class="text-zinc-600 text-xs mt-0.5">{{ row.batchLabel }} · Farm-wide</p>
            </div>
            <span
              v-if="row.lowStock"
              class="text-[10px] px-1.5 py-0.5 rounded-full bg-amber-900/60 text-amber-200 font-semibold shrink-0"
            >Low</span>
            <span
              v-else
              class="text-[10px] px-1.5 py-0.5 rounded-full bg-zinc-800 text-zinc-500 capitalize shrink-0"
            >{{ formatStatus(row.status) }}</span>
          </div>

          <dl class="grid grid-cols-2 gap-2 text-xs mb-3">
            <div>
              <dt class="text-zinc-600">On hand</dt>
              <dd class="text-zinc-200 font-mono">{{ formatQty(row.remaining) }}</dd>
            </div>
            <div v-if="row.threshold != null">
              <dt class="text-zinc-600">Low at</dt>
              <dd class="text-zinc-400 font-mono">{{ formatQty(row.threshold) }}</dd>
            </div>
            <div v-if="row.storageLocation" class="col-span-2">
              <dt class="text-zinc-600">Storage</dt>
              <dd class="text-zinc-400">{{ row.storageLocation }}</dd>
            </div>
          </dl>

          <div class="flex flex-wrap gap-2 pt-2 border-t border-zinc-800">
            <router-link
              v-if="mixCount(row.id)"
              :to="logMixLink"
              class="text-xs text-green-500 hover:text-green-400"
            >
              {{ mixCount(row.id) }} mix{{ mixCount(row.id) > 1 ? 'es' : '' }} logged
            </router-link>
            <button
              type="button"
              class="text-xs text-zinc-500 hover:text-zinc-300"
              @click="openBatchEditor(row.id)"
            >
              Edit batch →
            </button>
          </div>
        </div>
      </div>
    </div>

    <footer class="pt-2 border-t border-zinc-800">
      <router-link
        :to="{ path: '/inventory', query: zoneContextId ? { tab: 'definitions', zone_id: String(zoneContextId) } : { tab: 'definitions' } }"
        class="text-xs text-zinc-400 hover:text-green-400"
        data-test="supplies-advanced-footer"
      >
        Full inventory editor (definitions, recipes, batches) →
      </router-link>
    </footer>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useFarmStore } from '../stores/farm.js'
import { useFarmContextStore } from '../stores/farmContext.js'
import ZoneContextBanner from '../components/ZoneContextBanner.vue'
import EmptyStateHint from '../components/EmptyStateHint.vue'
import GuardianStarterChips from '../components/GuardianStarterChips.vue'
import { parseZoneIdQuery } from '../lib/zoneContext.js'
import { buildSuppliesHubStarters } from '../lib/guardianStarters.js'
import {
  buildSupplyRows,
  filterLowStockAlerts,
  listLowStockBatches,
} from '../lib/suppliesHub.js'

const route = useRoute()
const router = useRouter()
const store = useFarmStore()
const farmContext = useFarmContextStore()

const loading = ref(false)
const inputs = ref([])
const batches = ref([])
const recipes = ref([])
const alerts = ref([])
const programs = ref([])
const mixingComponentsByBatch = ref({})

const zoneContextId = computed(() => parseZoneIdQuery(route.query.zone_id))

const lowStockRows = computed(() => listLowStockBatches(batches.value, inputs.value))

const supplyRows = computed(() => buildSupplyRows(batches.value, inputs.value))

const lowStockAlerts = computed(() => filterLowStockAlerts(alerts.value))

const lowStockAlertLink = computed(() => {
  const first = lowStockAlerts.value[0]
  if (!first) return null
  return { path: '/alerts', query: { highlight: String(first.id) } }
})

const suppliesStarters = computed(() => buildSuppliesHubStarters({
  lowStockRows: lowStockRows.value,
  lowStockAlerts: lowStockAlerts.value,
  recipes: recipes.value,
  zones: store.zones,
  zoneContextId: zoneContextId.value,
  programs: programs.value,
  surface: zoneContextId.value ? 'supplies_hub_zone' : 'supplies_hub',
}))

const logMixLink = computed(() => {
  const q = { tab: 'mixing' }
  if (zoneContextId.value) q.zone_id = String(zoneContextId.value)
  return { path: '/operations/feeding', query: q }
})

const recipesLink = computed(() => ({
  path: '/inventory',
  query: { tab: 'recipes' },
}))

function zoneName(zoneId) {
  return store.zones.find((z) => z.id === zoneId)?.name || `Zone ${zoneId}`
}

function mixCount(batchId) {
  return mixingComponentsByBatch.value[batchId] || 0
}

function formatQty(value) {
  if (value == null || value === '') return '—'
  const n = Number(value)
  return Number.isFinite(n) ? String(n) : String(value)
}

function formatStatus(status) {
  return status ? String(status).replace(/_/g, ' ') : '—'
}

function openBatchEditor(batchId) {
  router.push({ path: '/inventory', query: { tab: 'batches', batch_id: String(batchId) } })
}

async function loadMixCounts(fid) {
  const counts = {}
  try {
    const mixEvents = await store.loadMixingEvents(fid)
    for (const me of mixEvents) {
      try {
        const comps = await store.loadMixingEventComponents(fid, me.id)
        for (const c of comps) {
          if (c.input_batch_id) counts[c.input_batch_id] = (counts[c.input_batch_id] || 0) + 1
        }
      } catch { /* skip */ }
    }
  } catch { /* skip */ }
  mixingComponentsByBatch.value = counts
}

async function refresh() {
  const fid = farmContext.farmId
  if (!fid) return
  loading.value = true
  try {
    if (!store.zones.length) await store.loadAll(fid)
    const [i, b, r, a, p] = await Promise.all([
      store.loadNfInputs(fid),
      store.loadNfBatches(fid),
      store.loadRecipes(fid),
      store.loadAlerts(fid, { limit: 100 }),
      store.loadFertigationPrograms(fid),
    ])
    inputs.value = i
    batches.value = b
    recipes.value = r
    alerts.value = a
    programs.value = p
    await loadMixCounts(fid)
  } finally {
    loading.value = false
  }
}

onMounted(refresh)
</script>
