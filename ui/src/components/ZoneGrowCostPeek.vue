<template>
  <div
    v-if="activeCycle"
    class="bg-zinc-950 border border-zinc-800 rounded-lg p-3 flex flex-wrap items-center justify-between gap-2"
    data-test="zone-grow-cost-peek"
  >
    <div>
      <p class="text-[10px] uppercase tracking-wide text-zinc-500 mb-0.5">This grow</p>
      <p v-if="loading" class="text-sm text-zinc-500">Loading cost…</p>
      <p v-else-if="loadError" class="text-xs text-zinc-500">{{ loadError }}</p>
      <p v-else class="text-sm text-zinc-200">
        <span class="font-mono tabular-nums">~${{ formatMoney(summary?.total_expenses) }}</span>
        spent
        <span v-if="summary?.net != null && summary.net !== summary?.total_expenses" class="text-zinc-500 text-xs ml-1">
          (net ${{ formatMoney(summary.net) }})
        </span>
      </p>
      <p class="text-[10px] text-zinc-600 mt-0.5">{{ cycleLabel }}</p>
    </div>
    <router-link
      v-nav-hint="'/operations/money'"
      :to="summaryLink"
      class="text-xs text-green-600 hover:text-green-400 shrink-0"
    >
      Cost details →
    </router-link>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { useFarmStore } from '../stores/farm.js'
import { formatCycleOptionLabel, formatMoney } from '../lib/moneyHub.js'

const props = defineProps({
  zoneId: { type: Number, required: true },
  farmId: { type: Number, required: true },
})

const store = useFarmStore()
const cycles = ref([])
const activeCycle = ref(null)
const summary = ref(null)
const loading = ref(false)
const loadError = ref('')

const cycleLabel = computed(() =>
  activeCycle.value ? formatCycleOptionLabel(activeCycle.value) : '',
)

const summaryLink = computed(() => ({
  path: `/crop-cycles/${activeCycle.value?.id}/summary`,
}))

async function load() {
  if (!props.farmId || !props.zoneId) return
  loading.value = true
  loadError.value = ''
  summary.value = null
  try {
    cycles.value = await store.loadCropCycles(props.farmId)
    activeCycle.value = cycles.value.find(
      (c) => c.is_active && Number(c.zone_id) === Number(props.zoneId),
    ) || null
    if (!activeCycle.value) return
    summary.value = await store.loadCropCycleCostSummary(activeCycle.value.id)
  } catch (e) {
    loadError.value = 'Cost summary unavailable'
  } finally {
    loading.value = false
  }
}

watch(() => [props.zoneId, props.farmId], load, { immediate: true })
</script>
