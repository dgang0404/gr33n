<template>
  <div
    class="bg-zinc-900 border border-zinc-800 rounded-xl p-4"
    data-test="zone-current-grow-strip"
  >
    <div class="flex items-start justify-between gap-3 flex-wrap">
      <div class="min-w-0">
        <h2 class="text-sm font-semibold text-white flex items-center gap-2">
          🌱 Current grow
        </h2>
        <p v-if="loading" class="text-zinc-500 text-xs mt-1">Loading crop cycles…</p>
        <template v-else-if="activeCycle">
          <p class="text-white text-sm font-medium mt-2 truncate">{{ activeCycle.name }}</p>
          <p class="text-zinc-400 text-xs mt-0.5">
            <span class="capitalize">{{ stageLabel }}</span>
            <span v-if="dayCount != null" class="text-zinc-600"> · day {{ dayCount }}</span>
            <span v-if="activeCycle.strain_or_variety" class="text-zinc-600">
              · {{ activeCycle.strain_or_variety }}
            </span>
          </p>
        </template>
        <p v-else class="text-zinc-500 text-xs mt-2">
          No active grow in this zone yet.
        </p>
      </div>
      <div class="flex flex-wrap items-center gap-2 shrink-0">
        <template v-if="activeCycle">
          <router-link
            v-nav-hint="summaryLink"
            :to="summaryLink"
            class="text-xs px-3 py-1.5 rounded-lg border border-zinc-700 text-zinc-300 hover:text-green-300 hover:border-green-800"
            data-test="grow-strip-summary-link"
          >
            Summary →
          </router-link>
          <button
            type="button"
            v-nav-hint="'/fertigation'"
            class="text-xs px-3 py-1.5 rounded-lg bg-amber-900/60 text-amber-100 border border-amber-800 hover:bg-amber-900/80"
            data-test="grow-strip-harvest-btn"
            @click="$emit('harvest', activeCycle)"
          >
            Harvest weigh-in
          </button>
        </template>
        <button
          v-else
          type="button"
          v-nav-hint="'/plants'"
          class="text-xs px-3 py-1.5 rounded-lg bg-green-900/60 text-green-300 border border-green-800 hover:bg-green-900/80 font-medium"
          data-test="grow-strip-start-btn"
          @click="$emit('start-grow')"
        >
          Start a grow
        </button>
      </div>
    </div>

    <GuardianStarterChips
      v-if="activeCycle && growStarters.length"
      :starters="growStarters"
      data-test="zone-grow-strip-starters"
    />
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { useFarmStore } from '../stores/farm.js'
import {
  activeCycleForZone,
  daysSinceStart,
  formatStageLabel,
  lastHarvestedCycleInZone,
} from '../lib/growHub.js'
import { buildZoneGrowStripStarters } from '../lib/guardianStarters.js'
import GuardianStarterChips from './GuardianStarterChips.vue'

const props = defineProps({
  zoneId: { type: Number, required: true },
  farmId: { type: Number, required: true },
  zone: { type: Object, default: null },
  /** Optional parent-owned cycles to avoid duplicate fetch. */
  cycles: { type: Array, default: null },
})

const emit = defineEmits(['start-grow', 'harvest', 'cycles-loaded'])

const store = useFarmStore()
const localCycles = ref([])
const loading = ref(false)

const cycleList = computed(() => props.cycles ?? localCycles.value)
const activeCycle = computed(() => activeCycleForZone(cycleList.value, props.zoneId))
const stageLabel = computed(() => formatStageLabel(activeCycle.value?.current_stage))
const dayCount = computed(() =>
  activeCycle.value ? daysSinceStart(activeCycle.value) : null,
)
const summaryLink = computed(() => ({
  path: `/crop-cycles/${activeCycle.value?.id}/summary`,
}))

const growStarters = computed(() => {
  if (!activeCycle.value) return []
  const zoneRef = props.zone || { id: props.zoneId, name: `Zone ${props.zoneId}` }
  return buildZoneGrowStripStarters({
    zone: zoneRef,
    activeCycle: activeCycle.value,
    farmId: props.farmId,
    priorHarvestedCycle: lastHarvestedCycleInZone(
      cycleList.value,
      props.zoneId,
      activeCycle.value.id,
    ),
  })
})

async function loadCycles() {
  if (props.cycles || !props.farmId) return
  loading.value = true
  try {
    localCycles.value = await store.loadCropCycles(props.farmId)
    emitLoaded()
  } finally {
    loading.value = false
  }
}

function emitLoaded() {
  if (props.cycles) return
  emit('cycles-loaded', localCycles.value)
}

watch(
  () => [props.farmId, props.cycles],
  () => {
    if (props.cycles) {
      emitLoaded()
      return
    }
    loadCycles()
  },
  { immediate: true },
)

defineExpose({ reload: loadCycles })
</script>
