<template>
  <div
    v-if="open"
    class="fixed inset-0 z-50 bg-black/70 p-4 flex items-center justify-center overflow-y-auto"
    data-test="post-harvest-screen"
    @click.self="close"
  >
    <div class="w-full max-w-2xl bg-zinc-900 border border-zinc-700 rounded-xl p-5 space-y-4 my-4">
      <div class="flex items-start justify-between gap-2">
        <div>
          <h2 class="text-white font-semibold">Harvest complete</h2>
          <p class="text-zinc-500 text-xs mt-0.5">Quick recap — compare to your last run in this zone when ready.</p>
        </div>
        <button type="button" class="text-zinc-500 hover:text-zinc-300 text-sm" @click="close">✕</button>
      </div>

      <div v-if="loading" class="text-zinc-500 text-sm">Loading summary…</div>
      <p v-else-if="loadError" class="text-red-400 text-sm">{{ loadError }}</p>

      <template v-else-if="summary">
        <section class="bg-zinc-950 border border-zinc-800 rounded-xl p-4">
          <p class="text-white text-sm font-medium">{{ summary.cycle.name }}</p>
          <p class="text-zinc-500 text-xs mt-0.5 capitalize">
            {{ formatStageLabel(summary.cycle.current_stage) }}
            · {{ summary.duration_days }} day{{ summary.duration_days === 1 ? '' : 's' }}
          </p>
        </section>

        <div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
          <div class="bg-zinc-950 border border-zinc-800 rounded-lg p-3" data-test="post-harvest-yield">
            <p class="text-[10px] uppercase tracking-wide text-zinc-500 mb-1">Yield</p>
            <p class="text-white text-lg font-mono">
              {{ summary.yield?.grams != null ? `${fmt(summary.yield.grams)} g` : '—' }}
            </p>
            <p v-if="summary.yield?.grams_per_day" class="text-zinc-500 text-[10px] mt-0.5">
              {{ fmt(summary.yield.grams_per_day) }} g/day
            </p>
          </div>
          <div class="bg-zinc-950 border border-zinc-800 rounded-lg p-3" data-test="post-harvest-fert">
            <p class="text-[10px] uppercase tracking-wide text-zinc-500 mb-1">Feeds</p>
            <p class="text-white text-lg font-mono">{{ summary.fertigation?.event_count ?? 0 }}</p>
            <p class="text-zinc-500 text-[10px] mt-0.5">
              {{ fmt(summary.fertigation?.total_liters) }} L delivered
            </p>
          </div>
          <div class="bg-zinc-950 border border-zinc-800 rounded-lg p-3" data-test="post-harvest-cost">
            <p class="text-[10px] uppercase tracking-wide text-zinc-500 mb-1">Tagged cost</p>
            <p class="text-white text-lg font-mono">
              {{ primaryCostLabel }}
            </p>
            <p v-if="summary.yield?.cost_per_gram" class="text-zinc-500 text-[10px] mt-0.5">
              {{ fmt(summary.yield.cost_per_gram) }}/g
            </p>
          </div>
        </div>
      </template>

      <GuardianStarterChips
        v-if="postHarvestStarters.length"
        :starters="postHarvestStarters"
        class="border-t border-zinc-800 pt-3"
      />

      <div class="flex flex-wrap items-center gap-2 border-t border-zinc-800 pt-3">
        <router-link
          v-if="cycleId"
          v-nav-hint="{ path: `/crop-cycles/${cycleId}/summary` }"
          :to="{ path: `/crop-cycles/${cycleId}/summary` }"
          class="text-xs px-3 py-1.5 rounded-lg border border-zinc-700 text-zinc-300 hover:text-green-300"
          data-test="post-harvest-full-summary"
        >
          Full summary →
        </router-link>
        <router-link
          v-nav-hint="compareRoute"
          :to="compareRoute"
          class="text-xs px-3 py-1.5 rounded-lg bg-green-900/50 text-green-300 border border-green-800 hover:bg-green-900/70"
          data-test="post-harvest-compare"
        >
          {{ compareLabel }} →
        </router-link>
        <button
          type="button"
          class="text-xs px-3 py-1.5 rounded-lg text-zinc-400 hover:text-zinc-200 ml-auto"
          @click="close"
        >
          Done
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { useFarmStore } from '../stores/farm.js'
import {
  buildPostHarvestCompareRoute,
  formatStageLabel,
  lastHarvestedCycleInZone,
} from '../lib/growHub.js'
import { buildHarvestFlowStarters } from '../lib/guardianStarters.js'
import GuardianStarterChips from './GuardianStarterChips.vue'

const props = defineProps({
  open: { type: Boolean, default: false },
  farmId: { type: Number, required: true },
  cycleId: { type: Number, default: null },
  zoneId: { type: Number, default: null },
  cycles: { type: Array, default: () => [] },
})

const emit = defineEmits(['close'])

const store = useFarmStore()
const summary = ref(null)
const loading = ref(false)
const loadError = ref('')

const priorCycle = computed(() => {
  if (!props.zoneId || !props.cycleId) return null
  return lastHarvestedCycleInZone(props.cycles, props.zoneId, props.cycleId)
})

const harvestedCycle = computed(() =>
  props.cycles.find((c) => Number(c.id) === Number(props.cycleId)) || null,
)

const postHarvestStarters = computed(() => {
  const zone = props.zoneId
    ? { id: props.zoneId, name: summary.value?.cycle?.zone_name || `Zone ${props.zoneId}` }
    : null
  return buildHarvestFlowStarters({
    zone,
    activeCycle: harvestedCycle.value,
    priorHarvestedCycle: priorCycle.value,
    farmId: props.farmId,
    allCycles: props.cycles,
    surface: 'post_harvest',
  })
})

const compareRoute = computed(() =>
  buildPostHarvestCompareRoute(props.farmId, props.cycles, props.cycleId, props.zoneId),
)

const compareLabel = computed(() =>
  priorCycle.value ? 'Compare to last cycle' : 'Open compare view',
)

const primaryCostLabel = computed(() => {
  const totals = summary.value?.cost?.totals
  if (!totals?.length) return '—'
  if (totals.length === 1) {
    return `${fmt(totals[0].total_expenses)} ${totals[0].currency}`
  }
  return 'Multiple currencies'
})

function fmt(n) {
  if (n == null || Number.isNaN(Number(n))) return '—'
  const num = Number(n)
  if (Math.abs(num) >= 100) return num.toFixed(0)
  if (Math.abs(num) >= 10) return num.toFixed(1)
  return num.toFixed(2)
}

function close() {
  emit('close')
}

async function load() {
  if (!props.cycleId) return
  loading.value = true
  loadError.value = ''
  summary.value = null
  try {
    summary.value = await store.loadCropCycleSummary(props.cycleId)
  } catch (e) {
    loadError.value = e?.response?.data?.error || e?.message || 'Failed to load summary'
  } finally {
    loading.value = false
  }
}

watch(
  () => [props.open, props.cycleId],
  ([isOpen, id]) => {
    if (isOpen && id) load()
  },
  { immediate: true },
)
</script>
