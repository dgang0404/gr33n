<template>
  <div
    v-if="hasLine"
    class="flex flex-wrap items-center gap-2 text-xs text-zinc-400 border border-zinc-800 rounded-lg px-3 py-2 bg-zinc-950/60"
    data-test="zone-grow-connection-line"
  >
    <span class="text-[10px] uppercase tracking-wide text-zinc-600 shrink-0">Grow chain</span>
    <router-link
      v-if="growLink"
      v-nav-hint="growLink"
      :to="growLink"
      class="text-zinc-200 hover:text-green-300 truncate max-w-[10rem]"
      data-test="connection-grow"
    >
      🌱 {{ growLabel }}
    </router-link>
    <span v-else class="text-zinc-600">No active grow</span>
    <span class="text-zinc-700">→</span>
    <router-link
      v-if="feedingLink"
      v-nav-hint="feedingLink"
      :to="feedingLink"
      class="hover:text-green-300 truncate max-w-[10rem]"
      data-test="connection-feeding"
    >
      💧 {{ feedingLabel }}
    </router-link>
    <span v-else class="text-zinc-600">No feeding plan</span>
    <span class="text-zinc-700">→</span>
    <router-link
      v-if="costLink"
      v-nav-hint="'/operations/money'"
      :to="costLink"
      class="text-green-600 hover:text-green-400 shrink-0"
      data-test="connection-cost"
    >
      {{ costLabel }} →
    </router-link>
    <span v-else class="text-zinc-600">No cost yet</span>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { useFarmStore } from '../stores/farm.js'
import { activeCycleForZone, formatStageLabel } from '../lib/growHub.js'
import { formatMoney } from '../lib/moneyHub.js'

const props = defineProps({
  zoneId: { type: Number, required: true },
  farmId: { type: Number, required: true },
  activeProgram: { type: Object, default: null },
})

const store = useFarmStore()
const cycles = ref([])
const costSummary = ref(null)

const activeCycle = computed(() => activeCycleForZone(cycles.value, props.zoneId))

const growLabel = computed(() => {
  if (!activeCycle.value) return ''
  const stage = formatStageLabel(activeCycle.value.current_stage)
  return `${activeCycle.value.name} (${stage})`
})

const growLink = computed(() =>
  activeCycle.value
    ? { path: `/crop-cycles/${activeCycle.value.id}/summary` }
    : { path: `/zones/${props.zoneId}`, query: { start_grow: '1' } },
)

const feedingLabel = computed(() => props.activeProgram?.name || '')
const feedingLink = computed(() =>
  props.activeProgram
    ? { path: '/operations/feeding', query: { zone: String(props.zoneId) } }
    : { path: `/zones/${props.zoneId}` },
)

const costLink = computed(() => {
  if (!activeCycle.value) return null
  return {
    path: `/crop-cycles/${activeCycle.value.id}/summary`,
  }
})

const costLabel = computed(() => {
  if (!activeCycle.value) return ''
  if (costSummary.value?.total_expenses != null) {
    return `~$${formatMoney(costSummary.value.total_expenses)}`
  }
  return 'Cost details'
})

const hasLine = computed(() => Boolean(props.zoneId && props.farmId))

async function load() {
  if (!props.farmId || !props.zoneId) return
  try {
    cycles.value = await store.loadCropCycles(props.farmId)
    const cycle = activeCycleForZone(cycles.value, props.zoneId)
    costSummary.value = cycle
      ? await store.loadCropCycleCostSummary(cycle.id)
      : null
  } catch {
    costSummary.value = null
  }
}

watch(() => [props.zoneId, props.farmId, props.activeProgram?.id], load, { immediate: true })
</script>
