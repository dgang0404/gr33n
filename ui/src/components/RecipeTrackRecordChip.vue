<template>
  <p
    v-if="text"
    class="text-[11px] text-violet-300/90 bg-violet-950/20 border border-violet-900/40 rounded-md px-2 py-1"
    :data-test="dataTest"
  >
    <span class="text-violet-400/80 uppercase tracking-wide text-[10px] mr-1">Track record</span>
    {{ text }}
  </p>
  <p
    v-else-if="insufficientText"
    class="text-[11px] text-zinc-500"
    :data-test="dataTest"
  >
    {{ insufficientText }}
  </p>
</template>

<script setup>
import { computed } from 'vue'
import {
  RECIPE_OUTCOME_MIN_SAMPLE,
  formatRecipeTrackRecord,
} from '../lib/recipeTrackRecord.js'

const props = defineProps({
  outcome: { type: Object, default: null },
  showCosts: { type: Boolean, default: true },
  dataTest: { type: String, default: 'recipe-track-record' },
})

const text = computed(() => formatRecipeTrackRecord(props.outcome, { showCosts: props.showCosts }))

const insufficientText = computed(() => {
  const o = props.outcome
  if (!o || Number(o.cycle_count) <= 0 || Number(o.cycle_count) >= RECIPE_OUTCOME_MIN_SAMPLE) return ''
  return `Only ${o.cycle_count} harvested cycle — need ${RECIPE_OUTCOME_MIN_SAMPLE}+ for a track record average.`
})
</script>
