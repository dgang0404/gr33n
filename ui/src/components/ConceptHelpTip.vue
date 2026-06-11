<template>
  <HelpTip v-if="concept" :position="position">
    <p class="font-medium text-zinc-200 mb-1">{{ concept.label }}</p>
    <p>{{ concept.shortTip }}</p>
    <p v-if="showTable" class="text-[10px] text-zinc-500 mt-1.5 font-mono">{{ concept.dbTable }}</p>
  </HelpTip>
</template>

<script setup>
import { computed } from 'vue'
import HelpTip from './HelpTip.vue'
import { operatorConcept } from '../lib/operatorConcepts.js'

const props = defineProps({
  conceptId: { type: String, required: true },
  position: { type: String, default: 'bottom' },
  showTable: { type: Boolean, default: false },
})

const concept = computed(() => operatorConcept(props.conceptId))
</script>
