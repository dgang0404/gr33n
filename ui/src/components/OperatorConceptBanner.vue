<template>
  <details
    class="rounded-xl border border-zinc-800 bg-zinc-900/60 text-xs text-zinc-400"
    :open="defaultOpen"
    data-test="operator-concept-banner"
  >
    <summary class="cursor-pointer px-4 py-2.5 text-zinc-300 font-medium select-none">
      What's the difference?
      <span class="text-zinc-500 font-normal ml-1">— these are separate database records, not the same thing</span>
    </summary>
    <ul class="px-4 pb-3 space-y-2 border-t border-zinc-800/80 pt-2">
      <li
        v-for="(line, i) in relationships"
        :key="`rel-${i}`"
        class="text-zinc-400 leading-relaxed list-disc ml-4"
      >
        {{ line }}
      </li>
      <li v-for="id in conceptIds" :key="id" class="leading-relaxed">
        <span class="text-zinc-200 font-medium">{{ concepts[id].label }}:</span>
        {{ concepts[id].shortTip }}
        <span class="text-[10px] text-zinc-600 font-mono ml-1">({{ concepts[id].dbTable }})</span>
      </li>
    </ul>
  </details>
</template>

<script setup>
import { computed } from 'vue'
import { OPERATOR_CONCEPTS, OPERATOR_CONCEPT_RELATIONSHIPS } from '../lib/operatorConcepts.js'

const props = defineProps({
  conceptIds: { type: Array, required: true },
  relationships: { type: Array, default: () => OPERATOR_CONCEPT_RELATIONSHIPS },
  defaultOpen: { type: Boolean, default: false },
})

const concepts = computed(() => OPERATOR_CONCEPTS)
const relationships = computed(() => props.relationships)
</script>
