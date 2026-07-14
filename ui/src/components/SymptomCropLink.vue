<template>
  <router-link
    v-if="to"
    v-nav-hint="'/operator-guide'"
    :to="to"
    class="text-[11px] text-green-600 hover:text-green-400"
    data-test="symptom-crop-link"
  >
    {{ label }}
  </router-link>
</template>

<script setup>
import { computed } from 'vue'
import { symptomGuideLabel, symptomGuideRoute } from '../lib/symptomGuideLink.js'

const props = defineProps({
  cropKey: { type: String, default: '' },
  labelOverride: { type: String, default: '' },
})

const to = computed(() => symptomGuideRoute(props.cropKey))

const label = computed(() => {
  if (props.labelOverride?.trim()) return props.labelOverride.trim()
  return symptomGuideLabel(props.cropKey) || 'Symptoms for this crop'
})
</script>
