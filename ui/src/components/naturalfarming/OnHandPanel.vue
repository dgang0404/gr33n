<template>
  <div class="space-y-6 max-w-4xl p-4" data-test="nf-on-hand">
    <div>
      <h2 class="text-lg font-semibold text-white">On hand</h2>
      <p class="text-sm text-zinc-500 mt-1">
        Ready batches you can apply now. Unit costs and full inventory editing stay in Money — this tab
        is the operator view.
      </p>
    </div>

    <div class="flex flex-wrap gap-2">
      <router-link
        :to="{ path: '/natural-farming', query: { tab: 'batch' } }"
        class="px-3 py-1.5 text-xs font-medium rounded-lg bg-green-900/50 text-green-400 border border-green-800 hover:bg-green-900/70"
        data-test="nf-stock-make-batch"
      >
        Make a batch
      </router-link>
      <router-link
        v-nav-hint="'/money'"
        :to="moneySuppliesLink"
        class="px-3 py-1.5 text-xs font-medium rounded-lg border border-zinc-700 text-zinc-300 hover:text-white"
        data-test="nf-stock-money-supplies"
      >
        Restock / edit costs → Money
      </router-link>
    </div>

    <p v-if="loading" class="text-sm text-zinc-500">Loading batches…</p>
    <p v-else-if="loadError" class="text-sm text-red-400">{{ loadError }}</p>

    <template v-else>
      <div
        v-if="lowStockRows.length"
        class="rounded-xl border border-amber-800/80 bg-amber-950/40 px-4 py-3 space-y-2"
        data-test="nf-stock-low-stock-banner"
      >
        <p class="text-sm font-medium text-amber-200">
          {{ lowStockRows.length }} ready batch{{ lowStockRows.length === 1 ? '' : 'es' }} below threshold
        </p>
        <ul class="text-sm text-amber-100/90 space-y-1">
          <li v-for="row in lowStockRows.slice(0, 5)" :key="row.batch.id">
            <strong>{{ row.inputName }}</strong>
            — {{ formatStockQty(row.remaining) }} left (threshold {{ formatStockQty(row.threshold) }})
          </li>
        </ul>
        <p v-if="lowStockRows.length > 5" class="text-xs text-amber-300/80">
          + {{ lowStockRows.length - 5 }} more below
        </p>
        <router-link
          v-if="lowStockAlertLink"
          v-nav-hint="'/alerts'"
          :to="lowStockAlertLink"
          class="inline-block text-xs text-amber-300 hover:text-amber-100 underline"
          data-test="nf-stock-alert-link"
        >
          View low-stock alert →
        </router-link>
      </div>

      <p v-if="!readyRows.length" class="text-sm text-zinc-500">
        No ready batches yet — ferment one on the Make a batch tab, or open Money for the full editor.
      </p>

      <div v-else class="space-y-3">
        <p class="text-xs text-zinc-500 uppercase tracking-widest">
          Ready — {{ readyRows.length }} batch{{ readyRows.length === 1 ? '' : 'es' }}
        </p>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div
            v-for="row in readyRows"
            :key="row.id"
            class="bg-zinc-900 border rounded-xl p-4 transition-colors"
            :class="[
              row.lowStock ? 'border-amber-800/80' : 'border-zinc-800',
              highlightedBatchId === row.id ? 'ring-1 ring-green-700' : '',
            ]"
            :data-test="`nf-stock-row-${row.id}`"
          >
            <div class="flex items-start justify-between gap-2 mb-2">
              <div>
                <p class="text-white font-medium">{{ row.inputName }}</p>
                <p class="text-zinc-600 text-xs mt-0.5">{{ row.batchLabel }}</p>
              </div>
              <span
                v-if="row.lowStock"
                class="text-[10px] px-1.5 py-0.5 rounded-full bg-amber-900/60 text-amber-200 font-semibold shrink-0"
              >Low</span>
              <span
                v-else
                class="text-[10px] px-1.5 py-0.5 rounded-full bg-green-900/40 text-green-300 capitalize shrink-0"
              >{{ formatStatus(row.status) }}</span>
            </div>
            <dl class="grid grid-cols-2 gap-2 text-xs mb-3">
              <div>
                <dt class="text-zinc-600">On hand</dt>
                <dd class="text-zinc-200 font-mono">{{ formatStockQty(row.remaining) }}</dd>
              </div>
              <div v-if="row.threshold != null">
                <dt class="text-zinc-600">Low at</dt>
                <dd class="text-zinc-400 font-mono">{{ formatStockQty(row.threshold) }}</dd>
              </div>
              <div v-if="row.storageLocation" class="col-span-2">
                <dt class="text-zinc-600">Storage</dt>
                <dd class="text-zinc-400">{{ row.storageLocation }}</dd>
              </div>
              <div class="col-span-2">
                <dt class="text-zinc-600">Unit cost</dt>
                <dd class="text-zinc-400">{{ row.unitCostLabel || 'Set in Money → Supplies' }}</dd>
              </div>
            </dl>
            <div class="flex flex-wrap gap-2 pt-2 border-t border-zinc-800">
              <router-link
                :to="{ path: '/natural-farming', query: { tab: 'recipes' } }"
                class="text-xs text-green-500 hover:text-green-400"
              >
                Apply recipe →
              </router-link>
              <router-link
                v-nav-hint="'/money'"
                :to="moneySuppliesLink"
                class="text-xs text-zinc-500 hover:text-zinc-300"
              >
                Edit in Money
              </router-link>
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useFarmStore } from '../../stores/farm.js'
import { useFarmContextStore } from '../../stores/farmContext.js'
import { enumLabel, loadDomainEnums } from '../../lib/domainEnums.js'
import {
  filterLowStockAlerts,
  formatStockQty,
  lowStockFromReady,
  stockRows,
} from '../../lib/naturalFarmingStock.js'
import { moneyTabRoute } from '../../lib/workspaceRoutes.js'
import api from '../../api'

const route = useRoute()
const store = useFarmStore()
const farmContext = useFarmContextStore()
const { farmId } = storeToRefs(farmContext)

const loading = ref(true)
const loadError = ref('')
const inputs = ref([])
const batches = ref([])
const alerts = ref([])
const domainEnums = ref(null)

const moneySuppliesLink = moneyTabRoute('supplies')

const readyRows = computed(() => stockRows(batches.value, inputs.value))
const lowStockRows = computed(() => lowStockFromReady(batches.value, inputs.value))
const lowStockAlerts = computed(() => filterLowStockAlerts(alerts.value))

const lowStockAlertLink = computed(() => {
  const first = lowStockAlerts.value[0]
  if (!first) return null
  return { path: '/alerts', query: { highlight: String(first.id) } }
})

const highlightedBatchId = computed(() => {
  const raw = route.query.batch_id
  const id = Number(Array.isArray(raw) ? raw[0] : raw)
  return Number.isFinite(id) ? id : null
})

function formatStatus(status) {
  return enumLabel('batch_statuses', status, domainEnums.value) || String(status || '').replace(/_/g, ' ')
}

async function loadAll() {
  if (!farmId.value) return
  try {
    const [i, b, a, enums] = await Promise.all([
      store.loadNfInputs(farmId.value),
      store.loadNfBatches(farmId.value),
      store.loadAlerts(farmId.value),
      loadDomainEnums(api),
    ])
    inputs.value = i
    batches.value = b
    alerts.value = a
    domainEnums.value = enums
  } catch (e) {
    loadError.value = e?.message || 'Failed to load batches'
  }
}

onMounted(async () => {
  try {
    await loadAll()
  } finally {
    loading.value = false
  }
})

watch(farmId, () => {
  loading.value = true
  loadAll().finally(() => { loading.value = false })
})
</script>
