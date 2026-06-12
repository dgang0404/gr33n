<!--
  Phase 28 WS2 — Crop Cycle Compare.

  Multi-select crop cycles from the farm and render their summaries side
  by side. Backend caps at 5 cycles per call; UI mirrors that with a
  disabled-selection state. Best/worst columns per metric row get a
  subtle highlight so the operator can see at a glance "which cycle
  produced more grams per liter?".
-->
<template>
  <div class="p-6 max-w-7xl">
    <div class="flex items-center justify-between mb-4">
      <div class="flex items-center gap-2">
        <router-link v-nav-hint="'/zones'" :to="{ path: '/zones', query: { tab: 'plants' } }" class="text-xs text-zinc-400 hover:text-zinc-200">← Plants</router-link>
        <h1 class="text-xl font-semibold text-white ml-3">
          Compare crop cycles
          <HelpTip position="bottom">
            Pick up to 5 cycles from the current farm and view their summaries side by side.
            Best and worst columns per metric row are highlighted automatically — higher is better for yield, lower is better for cost-per-gram.
          </HelpTip>
        </h1>
      </div>
      <a
        v-if="selectedIds.length"
        :href="csvUrl"
        target="_blank"
        rel="noopener"
        class="text-xs font-medium px-3 py-1.5 rounded-lg bg-zinc-900 text-zinc-300 border border-zinc-700 hover:bg-zinc-800"
      >Download CSV</a>
    </div>

    <!-- Farm hint -->
    <div v-if="!farmId" class="rounded-xl border border-amber-900/60 bg-amber-950/40 px-4 py-3 text-sm text-amber-200 mb-4">
      Select a farm from the sidebar to start comparing crop cycles.
    </div>

    <template v-else>
      <!-- Picker -->
      <section data-test="picker" class="bg-zinc-800 border border-zinc-700 rounded-xl p-4 mb-5">
        <div class="flex items-center justify-between mb-3">
          <p class="text-zinc-300 text-sm">
            {{ selectedIds.length }} of {{ MAX_COMPARE }} selected
            <span v-if="loadingCycles" class="text-zinc-500 text-xs ml-2">loading cycles…</span>
          </p>
          <button
            v-if="selectedIds.length"
            type="button"
            @click="clearSelection"
            class="text-xs text-zinc-400 hover:text-zinc-200"
          >Clear</button>
        </div>

        <div v-if="!loadingCycles && !cycles.length" class="text-zinc-500 text-xs">
          No crop cycles yet on this farm.
        </div>

        <div v-else class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-2">
          <label
            v-for="c in cycles"
            :key="c.id"
            class="flex items-center gap-2 px-3 py-2 rounded-lg border text-xs cursor-pointer transition-colors"
            :class="isSelected(c.id)
              ? 'bg-green-900/40 border-green-700 text-white'
              : isDisabled(c.id)
                ? 'bg-zinc-900/60 border-zinc-800 text-zinc-600 cursor-not-allowed'
                : 'bg-zinc-900 border-zinc-700 text-zinc-300 hover:border-zinc-500'"
          >
            <input
              type="checkbox"
              :value="c.id"
              :checked="isSelected(c.id)"
              :disabled="!isSelected(c.id) && isDisabled(c.id)"
              @change="toggleSelect(c.id)"
              class="accent-emerald-500"
            />
            <span class="truncate">{{ c.name }}</span>
          </label>
        </div>
      </section>

      <!-- Empty state -->
      <div v-if="!selectedIds.length" class="text-zinc-500 text-sm bg-zinc-900 border border-zinc-800 rounded-xl p-6 text-center">
        Pick two or more crop cycles above to compare them side by side.
      </div>

      <!-- Loading + error -->
      <div v-else-if="loadingCompare" class="text-zinc-400 text-sm">Loading comparison…</div>
      <div v-else-if="loadError" class="text-red-400 text-sm">{{ loadError }}</div>

      <!-- Comparison table -->
      <section v-else-if="summaries.length" data-test="compare-table" class="bg-zinc-800 border border-zinc-700 rounded-xl overflow-x-auto">
        <table class="w-full text-sm">
          <thead class="bg-zinc-900 border-b border-zinc-700">
            <tr>
              <th class="text-left text-zinc-500 text-[11px] uppercase tracking-wide px-3 py-2">Metric</th>
              <th
                v-for="s in summaries"
                :key="s.cycle.id"
                class="text-left text-white text-sm font-medium px-3 py-2"
              >{{ s.cycle.name }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in compareRows" :key="row.key" class="border-b border-zinc-800 last:border-0">
              <th class="text-left text-zinc-400 text-xs font-normal px-3 py-2">
                {{ row.label }}
              </th>
              <td
                v-for="(val, idx) in row.values"
                :key="idx"
                class="px-3 py-2 font-mono text-sm"
                :class="cellClass(row, idx)"
              >{{ val.display }}</td>
            </tr>
          </tbody>
        </table>
      </section>
    </template>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import HelpTip from '../components/HelpTip.vue'
import { useFarmContextStore } from '../stores/farmContext'
import { useFarmStore } from '../stores/farm'

// Mirror the backend MaxCompareCycles. Drift between server + UI would
// only ever show as a 400, not a security issue, but keeping them in
// sync avoids the "Add cycle 6 → silent failure" surprise.
const MAX_COMPARE = 5

const route = useRoute()
const farmContext = useFarmContextStore()
const store = useFarmStore()

const farmId = computed(() => {
  const fromRoute = Number(route.params.fid)
  return fromRoute || farmContext.farmId
})
const cycles = ref([])
const loadingCycles = ref(false)
const selectedIds = ref([])
const summaries = ref([])
const loadingCompare = ref(false)
const loadError = ref('')

const csvUrl = computed(() => {
  if (!farmId.value || !selectedIds.value.length) return '#'
  const base = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'
  const tok = localStorage.getItem('gr33n_token') ?? ''
  const ids = selectedIds.value.join(',')
  return `${base}/farms/${farmId.value}/crop-cycles/compare.csv?ids=${ids}&token=${encodeURIComponent(tok)}`
})

function isSelected(id) {
  return selectedIds.value.includes(id)
}
function isDisabled(id) {
  return selectedIds.value.length >= MAX_COMPARE && !isSelected(id)
}
function toggleSelect(id) {
  if (isSelected(id)) {
    selectedIds.value = selectedIds.value.filter((x) => x !== id)
  } else if (selectedIds.value.length < MAX_COMPARE) {
    selectedIds.value = [...selectedIds.value, id]
  }
}
function clearSelection() {
  selectedIds.value = []
  summaries.value = []
}

async function loadCycles() {
  if (!farmId.value) return
  loadingCycles.value = true
  try {
    cycles.value = await store.loadCropCycles(farmId.value)
  } finally {
    loadingCycles.value = false
  }
}

async function loadCompare() {
  if (!farmId.value || !selectedIds.value.length) {
    summaries.value = []
    return
  }
  loadingCompare.value = true
  loadError.value = ''
  try {
    const data = await store.loadCropCycleCompare(farmId.value, selectedIds.value)
    summaries.value = Array.isArray(data?.cycles) ? data.cycles : []
  } catch (err) {
    loadError.value = err?.response?.data?.error || err?.message || 'Failed to load comparison'
    summaries.value = []
  } finally {
    loadingCompare.value = false
  }
}

// ── Compare table builder ─────────────────────────────────────────────
// Each row is one metric across all selected cycles. The `betterIsHigher`
// flag drives the best/worst highlight: yield wants higher, cost-per-gram
// wants lower. Rows whose value is undefined on every cycle are filtered
// out so the table stays compact.

function value(num) {
  if (num == null || Number.isNaN(Number(num))) return { display: '—', sortable: null }
  return { display: fmtNum(num), sortable: Number(num) }
}
function fmtNum(n) {
  const num = Number(n)
  if (Math.abs(num) >= 100) return num.toFixed(0)
  if (Math.abs(num) >= 10) return num.toFixed(1)
  return num.toFixed(2)
}

function singleCurrencyTotal(s) {
  // Many of our cost metrics only make sense per-currency. When a cycle
  // has costs in exactly one currency we use it; otherwise we surface
  // the multi-currency case as a `—` rather than summing dollars + euros.
  if (s?.cost?.totals?.length === 1) {
    return s.cost.totals[0]
  }
  return null
}

const compareRows = computed(() => {
  if (!summaries.value.length) return []
  const rows = [
    { key: 'duration', label: 'Duration (days)', better: 'higher',
      values: summaries.value.map((s) => value(s.duration_days)) },
    { key: 'events', label: 'Fertigation events', better: 'higher',
      values: summaries.value.map((s) => value(s.fertigation?.event_count)) },
    { key: 'liters', label: 'Liters delivered', better: 'higher',
      values: summaries.value.map((s) => value(s.fertigation?.total_liters)) },
    { key: 'ec_avg', label: 'Avg EC (mS/cm)', better: 'none',
      values: summaries.value.map((s) => value(s.fertigation?.avg_ec_mscm)) },
    { key: 'ec_min_max', label: 'EC min / max', better: 'none',
      values: summaries.value.map((s) => ({
        display: `${fmtNum(s.fertigation?.min_ec_mscm ?? 0)} / ${fmtNum(s.fertigation?.max_ec_mscm ?? 0)}`,
        sortable: null,
      })) },
    { key: 'ph_avg', label: 'Avg pH', better: 'none',
      values: summaries.value.map((s) => value(s.fertigation?.avg_ph)) },
    { key: 'yield', label: 'Yield (g)', better: 'higher',
      values: summaries.value.map((s) => value(s.yield?.grams)) },
    { key: 'g_per_l', label: 'g per liter', better: 'higher',
      values: summaries.value.map((s) => value(s.yield?.grams_per_liter)) },
    { key: 'g_per_d', label: 'g per day', better: 'higher',
      values: summaries.value.map((s) => value(s.yield?.grams_per_day)) },
    { key: 'cost_per_g', label: 'Cost per gram', better: 'lower',
      values: summaries.value.map((s) => value(s.yield?.cost_per_gram)) },
    { key: 'expenses', label: 'Total expenses', better: 'lower',
      values: summaries.value.map((s) => {
        const t = singleCurrencyTotal(s)
        return t ? { display: `${fmtNum(t.total_expenses)} ${t.currency}`, sortable: t.total_expenses } : { display: '—', sortable: null }
      }) },
  ]
  return rows.filter((row) => row.values.some((v) => v.sortable != null || (v.display && v.display !== '—')))
})

function cellClass(row, idx) {
  if (row.better === 'none') return 'text-white'
  const sortables = row.values.map((v) => v.sortable).filter((n) => n != null)
  if (sortables.length < 2) return 'text-white'
  const target = row.better === 'higher' ? Math.max(...sortables) : Math.min(...sortables)
  const opposite = row.better === 'higher' ? Math.min(...sortables) : Math.max(...sortables)
  const v = row.values[idx]?.sortable
  if (v == null) return 'text-zinc-500'
  if (v === target && target !== opposite) return 'text-emerald-400'
  if (v === opposite && target !== opposite) return 'text-amber-400'
  return 'text-white'
}

function applyIdsFromQuery() {
  const raw = route.query.ids
  if (!raw || typeof raw !== 'string') return
  const ids = raw
    .split(',')
    .map((s) => Number(s.trim()))
    .filter((n) => Number.isFinite(n) && n > 0)
    .slice(0, MAX_COMPARE)
  if (ids.length) selectedIds.value = ids
}

onMounted(async () => {
  applyIdsFromQuery()
  await loadCycles()
  if (selectedIds.value.length) await loadCompare()
})
watch(farmId, () => {
  clearSelection()
  applyIdsFromQuery()
  loadCycles()
})
watch(() => route.query.ids, () => {
  applyIdsFromQuery()
  if (selectedIds.value.length) loadCompare()
})
watch(selectedIds, loadCompare)
</script>
