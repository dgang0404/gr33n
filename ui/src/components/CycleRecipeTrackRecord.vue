<template>
  <p
    v-if="line"
    class="text-xs text-violet-300/90 mt-3 border-t border-zinc-700/80 pt-3"
    data-test="cycle-recipe-track-record"
  >
    {{ line }}
  </p>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useFarmStore } from '../stores/farm.js'
import { useFarmCaps } from '../composables/useFarmCaps.js'
import { FARM_SCOPES } from '../lib/farmScopes.js'
import {
  attributeRecipeFromHits,
  attributionHitsFromOpsEvents,
  findRecipeOutcome,
  formatCycleRecipeTrackRecord,
} from '../lib/recipeTrackRecord.js'

const props = defineProps({
  farmId: { type: Number, required: true },
  cycleId: { type: Number, required: true },
  cropKey: { type: String, default: '' },
})

const store = useFarmStore()
const { has: hasScope } = useFarmCaps(() => props.farmId)

const line = ref('')

const showCosts = computed(() => hasScope(FARM_SCOPES.moneyRead))

async function refresh() {
  line.value = ''
  if (!props.farmId || !props.cycleId) return
  try {
    const [ops, data] = await Promise.all([
      store.loadCropCycleOpsTimeline(props.farmId, props.cycleId),
      store.loadRecipeOutcomes(props.farmId, props.cropKey ? { cropKey: props.cropKey } : {}),
    ])
    const attr = attributeRecipeFromHits(attributionHitsFromOpsEvents(ops?.events))
    if (attr.empty || attr.mixed || !attr.application_recipe_id) return
    const outcome = findRecipeOutcome(
      data?.outcomes,
      attr.application_recipe_id,
      attr.application_recipe_revision_id,
    )
    line.value = formatCycleRecipeTrackRecord(outcome, {
      showCosts: showCosts.value,
      revisionId: attr.application_recipe_revision_id,
    })
  } catch {
    line.value = ''
  }
}

onMounted(refresh)
watch(() => [props.farmId, props.cycleId, props.cropKey, showCosts.value], refresh)
</script>
