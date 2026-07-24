<!--
  Phase 28 WS2 — Crop Cycle Summary.

  Renders the response of GET /crop-cycles/{id}/summary as four metric
  cards (fertigation, cost, yield, stages) plus a header strip. Includes
  a CSV download (opens /crop-cycles/{id}/summary.csv with the JWT
  attached via the same axios instance behind a query string).
-->
<template>
  <div class="p-6 max-w-5xl">
    <div class="flex items-center justify-between mb-4">
      <div class="flex items-center gap-2">
        <router-link v-nav-hint="'/fertigation'" to="/fertigation" class="text-xs text-zinc-400 hover:text-zinc-200">← Fertigation</router-link>
        <h1 class="text-xl font-semibold text-white ml-3">
          Crop cycle summary
          <HelpTip position="bottom">
            Per-cycle rollup: fertigation history, cost totals tagged to the cycle, yield metrics and stage info.
            Most rows roll up rows where <code>crop_cycle_id</code> matches this cycle. Numbers may differ from the
            zone-wide totals because zone-tagged costs that don't link to a cycle are excluded here.
          </HelpTip>
        </h1>
      </div>
      <div class="flex items-center gap-3">
        <AskGuardianButton
          v-if="cycleId"
          variant="primary"
          size="sm"
          :prefilled-message="'Summarize this cycle and compare to typical flower targets'"
          :context-ref="{ type: 'crop_cycle', id: cycleId }"
        />
        <button
          v-if="cycleId"
          type="button"
          class="text-xs font-medium px-3 py-1.5 rounded-lg bg-zinc-900 text-zinc-300 border border-zinc-700 hover:bg-zinc-800"
          @click="downloadCsv"
        >Download CSV</button>
        <router-link
          v-if="summary && summary.cycle && compareRoute"
          v-nav-hint="'/fertigation'"
          :to="compareRoute"
          class="text-xs font-medium px-3 py-1.5 rounded-lg bg-green-900/50 text-green-400 border border-green-800 hover:bg-green-900/70"
        >Compare ↔</router-link>
      </div>
    </div>

    <div v-if="loading" class="text-zinc-400 text-sm">Loading summary…</div>
    <div v-else-if="loadError" class="text-red-400 text-sm">{{ loadError }}</div>

    <template v-else-if="summary">
      <!-- Header strip -->
      <section data-test="summary-header" class="bg-zinc-800 border border-zinc-700 rounded-xl p-5 mb-5">
        <div class="flex flex-wrap items-center gap-x-6 gap-y-2">
          <div>
            <p class="text-zinc-500 text-[11px] uppercase tracking-wide">Cycle</p>
            <p class="text-white text-base font-medium">{{ summary.cycle.name }}</p>
          </div>
          <div v-if="cycleBatchLabel(summary.cycle)">
            <p class="text-zinc-500 text-[11px] uppercase tracking-wide">Batch</p>
            <p class="text-white text-sm">{{ cycleBatchLabel(summary.cycle) }}</p>
          </div>
          <div>
            <p class="text-zinc-500 text-[11px] uppercase tracking-wide">Stage</p>
            <p class="text-white text-sm">{{ currentStageLabel }}</p>
          </div>
          <div>
            <p class="text-zinc-500 text-[11px] uppercase tracking-wide">Duration</p>
            <p class="text-white text-sm">{{ summary.duration_days }} day{{ summary.duration_days === 1 ? '' : 's' }}</p>
          </div>
          <div>
            <p class="text-zinc-500 text-[11px] uppercase tracking-wide">Status</p>
            <p :class="summary.cycle.is_active ? 'text-emerald-400' : 'text-zinc-300'" class="text-sm">
              {{ summary.cycle.is_active ? 'Active' : 'Harvested' }}
            </p>
          </div>
        </div>
      </section>

      <section
        v-if="harvestEconomics"
        data-test="summary-harvest-economics"
        class="bg-emerald-950/30 border border-emerald-800/50 rounded-xl p-5 mb-5"
      >
        <h2 class="text-emerald-300 text-sm font-semibold mb-2">Harvest economics</h2>
        <div class="grid grid-cols-1 sm:grid-cols-3 gap-4 text-sm">
          <Metric label="Spent on this grow" :value="fmt(harvestEconomics.expenses)" :suffix="' ' + harvestEconomics.currency" />
          <Metric label="Income tagged" :value="fmt(harvestEconomics.income)" :suffix="' ' + harvestEconomics.currency" />
          <Metric
            label="Net"
            :value="fmt(harvestEconomics.net)"
            :suffix="' ' + harvestEconomics.currency"
          />
        </div>
        <router-link
          v-if="summary.cycle"
          v-nav-hint="'/money'"
          :to="{ path: '/money', query: { cycle_id: summary.cycle.id, tab: 'summary' } }"
          class="inline-block mt-3 text-xs text-green-500 hover:text-green-300"
          data-test="summary-income-for-grow-link"
        >
          Income for this grow →
        </router-link>
      </section>

      <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <!-- Fertigation -->
        <section data-test="card-fertigation" class="bg-zinc-800 border border-zinc-700 rounded-xl p-5">
          <h2 class="text-white text-sm font-semibold flex items-center gap-2 mb-3">
            💧 Fertigation
            <HelpTip>Rolls up every fertigation_event with this <code>crop_cycle_id</code>. EC averages use the <em>after-feed</em> reading so the number reflects what the plants experienced.</HelpTip>
          </h2>
          <div class="grid grid-cols-2 gap-3 text-sm">
            <Metric label="Events"          :value="summary.fertigation.event_count" />
            <Metric label="Liters delivered" :value="fmt(summary.fertigation.total_liters)" suffix=" L" />
            <Metric label="Avg EC"          :value="fmt(summary.fertigation.avg_ec_mscm)" suffix=" mS/cm" />
            <Metric label="Avg pH"          :value="fmt(summary.fertigation.avg_ph)" />
            <Metric label="EC min / max"    :value="`${fmt(summary.fertigation.min_ec_mscm)} / ${fmt(summary.fertigation.max_ec_mscm)}`" />
          </div>
        </section>

        <!-- Cost -->
        <section data-test="card-cost" class="bg-zinc-800 border border-zinc-700 rounded-xl p-5">
          <h2 class="text-white text-sm font-semibold flex items-center gap-2 mb-3">
            💰 Cost
            <HelpTip>Only cost_transactions explicitly linked to this cycle. Zone-level costs that don't reference the cycle are excluded.</HelpTip>
          </h2>
          <div v-if="!costTotals.length" class="text-zinc-500 text-xs">No costs tagged to this cycle yet.</div>
          <div v-else class="space-y-3">
            <div v-for="t in costTotals" :key="t.currency" class="grid grid-cols-3 gap-3 text-sm">
              <Metric label="Expenses" :value="fmt(t.total_expenses)" :suffix="' ' + t.currency" />
              <Metric label="Income"   :value="fmt(t.total_income)"   :suffix="' ' + t.currency" />
              <Metric label="Net"      :value="fmt(t.net)"            :suffix="' ' + t.currency" />
            </div>
            <div v-if="costByCategory.length" class="border-t border-zinc-700 pt-3">
              <p class="text-zinc-500 text-[11px] uppercase tracking-wide mb-2">By category</p>
              <ul class="text-xs text-zinc-300 space-y-1">
                <li v-for="row in costByCategory" :key="row.category + row.currency" class="flex items-center justify-between">
                  <span>{{ row.category }} <span class="text-zinc-500">({{ row.tx_count }} tx)</span></span>
                  <span class="font-mono">{{ fmt(row.expense - row.income) }} {{ row.currency }}</span>
                </li>
              </ul>
            </div>
          </div>
        </section>

        <!-- Yield -->
        <section data-test="card-yield" class="bg-zinc-800 border border-zinc-700 rounded-xl p-5">
          <h2 class="text-white text-sm font-semibold flex items-center gap-2 mb-3">
            🌾 Yield
            <HelpTip><code>cost_per_gram</code> is only emitted when the cycle has costs in a single currency — mixing currencies blindly would be misleading.</HelpTip>
          </h2>
          <div v-if="!summary.yield.grams" class="text-zinc-500 text-xs">No yield recorded yet (set <code>yield_grams</code> on the cycle).</div>
          <div v-else class="grid grid-cols-2 gap-3 text-sm">
            <Metric label="Yield"            :value="fmt(summary.yield.grams)" suffix=" g" />
            <Metric label="g per liter"      :value="optFmt(summary.yield.grams_per_liter)" />
            <Metric label="g per day"        :value="optFmt(summary.yield.grams_per_day)" />
            <Metric label="Cost per gram"    :value="optFmt(summary.yield.cost_per_gram, 'mixed currencies')" />
          </div>
        </section>

        <!-- Stages -->
        <section data-test="card-stages" class="bg-zinc-800 border border-zinc-700 rounded-xl p-5">
          <h2 class="text-white text-sm font-semibold flex items-center gap-2 mb-3">
            ⏳ Stage timeline
            <HelpTip>
              Stage transitions recorded when you advance a grow appear here with dates.
              Older cycles may show only the current stage until history is backfilled.
            </HelpTip>
          </h2>
          <ol class="border-l border-zinc-700 ml-2 space-y-3">
            <li v-for="(s, idx) in summary.stages" :key="idx" class="pl-4 relative">
              <span class="absolute -left-1.5 top-1 w-3 h-3 rounded-full bg-emerald-500/70"></span>
              <p class="text-white text-sm">{{ s.stage }}</p>
              <p class="text-zinc-500 text-xs">{{ s.entered_at || '—' }}</p>
            </li>
          </ol>
          <p v-if="!summary.stage_history_supported" class="text-amber-400/80 text-[11px] mt-3">
            Stage history will fill in as you advance stages on this grow.
          </p>
        </section>
      </div>

      <CropOpsTimeline
        v-if="summary.cycle?.farm_id"
        class="mt-5"
        :farm-id="summary.cycle.farm_id"
        :cycle-id="cycleId"
        :default-from="summary.cycle.started_at"
        :default-to="summary.cycle.harvested_at"
      />
    </template>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { useFarmStore } from '../stores/farm'
import { buildPostHarvestCompareRoute, cycleBatchLabel, formatStageLabel } from '../lib/growHub.js'
import HelpTip from '../components/HelpTip.vue'
import Metric from '../components/MetricChip.vue'
import AskGuardianButton from '../components/AskGuardianButton.vue'
import CropOpsTimeline from '../components/CropOpsTimeline.vue'
import { downloadWithAuth } from '../lib/downloadAuth.js'

const route = useRoute()
const store = useFarmStore()

const cycleId = computed(() => Number(route.params.id))
const summary = ref(null)
const farmCycles = ref([])
const loading = ref(false)
const loadError = ref('')

const compareRoute = computed(() => {
  const cycle = summary.value?.cycle
  if (!cycle?.farm_id || !cycleId.value) return null
  return buildPostHarvestCompareRoute(cycle.farm_id, farmCycles.value, cycleId.value, cycle.zone_id)
})

const costTotals = computed(() => summary.value?.cost?.totals ?? [])
const costByCategory = computed(() => summary.value?.cost?.by_category ?? [])

const harvestEconomics = computed(() => {
  const totals = costTotals.value
  const row = totals.find((t) => Number(t.total_income) > 0)
  if (!row) return null
  return {
    currency: row.currency || 'USD',
    expenses: row.total_expenses,
    income: row.total_income,
    net: row.net,
  }
})

async function downloadCsv() {
  if (!cycleId.value) return
  const base = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'
  await downloadWithAuth(
    `${base}/crop-cycles/${cycleId.value}/summary.csv`,
    `crop-cycle-${cycleId.value}-summary.csv`,
  )
}

const currentStageLabel = computed(() => {
  if (!summary.value || !summary.value.cycle) return '—'
  const s = summary.value.cycle.current_stage
  if (!s) return '—'
  if (typeof s === 'string') return s
  if (s.Valid && s.Gr33nfertigationGrowthStageEnum) {
    return formatStageLabel(s.Gr33nfertigationGrowthStageEnum)
  }
  return '—'
})

function fmt(n) {
  if (n == null || Number.isNaN(Number(n))) return '0'
  const num = Number(n)
  if (Math.abs(num) >= 100) return num.toFixed(0)
  if (Math.abs(num) >= 10) return num.toFixed(1)
  return num.toFixed(2)
}

function optFmt(n, fallback = '—') {
  if (n == null) return fallback
  return fmt(n)
}

async function load() {
  if (!cycleId.value) return
  loading.value = true
  loadError.value = ''
  try {
    const sum = await store.loadCropCycleSummary(cycleId.value)
    summary.value = sum
    if (sum?.cycle?.farm_id) {
      farmCycles.value = await store.loadCropCycles(sum.cycle.farm_id).catch(() => [])
    }
  } catch (err) {
    loadError.value = err?.response?.data?.error || err?.message || 'Failed to load cycle summary'
  } finally {
    loading.value = false
  }
}

async function scrollToOpsAnchor() {
  if (route.hash !== '#crop-ops-timeline') return
  await nextTick()
  document.getElementById('crop-ops-timeline')?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

async function loadAndScroll() {
  await load()
  await scrollToOpsAnchor()
}

onMounted(loadAndScroll)
watch(cycleId, loadAndScroll)
watch(() => route.hash, scrollToOpsAnchor)
</script>
