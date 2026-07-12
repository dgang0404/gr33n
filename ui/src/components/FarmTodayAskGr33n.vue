<template>
  <section
    v-if="starters.length"
    class="space-y-2"
    data-test="farm-today-ask-gr33n"
  >
    <p class="text-[10px] font-semibold text-zinc-500 uppercase tracking-widest">Ask gr33n</p>
    <GuardianStarterChips :starters="starters" />
  </section>
</template>

<script setup>
import GuardianStarterChips from './GuardianStarterChips.vue'

defineProps({
  starters: { type: Array, default: () => [] },
})
</script>
