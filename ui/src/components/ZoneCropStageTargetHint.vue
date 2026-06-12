<template>
  <div
    v-if="lines.length"
    class="rounded-lg border border-green-900/50 bg-green-950/20 px-3 py-2 text-[11px] text-zinc-300 space-y-0.5"
    data-test="zone-crop-stage-target-hint"
  >
    <p class="text-[10px] uppercase tracking-wide text-zinc-500 mb-1">
      Crop targets ({{ cropLabel }} · {{ stageLabel }})
    </p>
    <p v-for="(line, i) in lines" :key="i" class="font-mono text-zinc-200">{{ line }}</p>
    <router-link to="/settings" class="text-green-500 hover:text-green-400 text-[10px]">
      Adjust targets in Settings → Crops &amp; targets
    </router-link>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { useFarmStore } from '../stores/farm.js'
import { activeCycleForZone, formatEcTargetChip, formatStageLabel } from '../lib/growHub.js'
import { formatStageTargetLine } from '../lib/cropLibraryPicker.js'

const props = defineProps({
  zoneId: { type: Number, required: true },
  farmId: { type: Number, required: true },
  cycles: { type: Array, default: null },
})

const store = useFarmStore()
const localCycles = ref([])
const lines = ref([])
const cropLabel = ref('')
const stageLabel = ref('')

const cycleList = computed(() => props.cycles ?? localCycles.value)
const activeCycle = computed(() => activeCycleForZone(cycleList.value, props.zoneId))

async function loadCycles() {
  if (props.cycles || !props.farmId) return
  localCycles.value = await store.loadCropCycles(props.farmId)
}

async function loadTargets() {
  lines.value = []
  cropLabel.value = ''
  stageLabel.value = ''
  const cycle = activeCycle.value
  if (!cycle?.plant_id || !cycle.current_stage) return
  stageLabel.value = formatStageLabel(cycle.current_stage)
  try {
    const plant = await store.getPlant(cycle.plant_id)
    cropLabel.value = plant.display_name || plant.crop_key || 'Crop'
    if (!plant.crop_profile_id) return
    const profile = await store.getCropProfile(plant.crop_profile_id)
    const stageRow = (profile.stages || []).find((s) => s.stage === cycle.current_stage)
    if (!stageRow) return
    const out = []
    const ec = formatEcTargetChip(stageRow)
    if (ec) out.push(ec)
    const full = formatStageTargetLine(stageRow)
    if (full && !out.includes(full)) out.push(full)
    lines.value = out.filter(Boolean)
  } catch {
    lines.value = []
  }
}

watch(
  () => [props.farmId, props.cycles],
  () => loadCycles(),
  { immediate: true },
)

watch(
  () => [activeCycle.value?.id, activeCycle.value?.current_stage, activeCycle.value?.plant_id],
  () => loadTargets(),
  { immediate: true },
)
</script>
