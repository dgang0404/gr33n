<template>
  <div
    v-if="procedure"
    class="rounded-lg border border-green-900/60 bg-green-950/30 px-3 py-2 text-xs space-y-2"
    data-test="guardian-procedure-card"
  >
    <div class="flex items-center justify-between gap-2">
      <span class="font-medium text-green-300">{{ procedure.title }}</span>
      <span class="text-[10px] text-zinc-500">
        Step {{ procedure.step_n }}/{{ procedure.step_total }}
        <span v-if="procedure.safety_tier" class="text-amber-400/90"> · {{ procedure.safety_tier }}</span>
      </span>
    </div>
    <p v-if="procedure.safety_stopped" class="text-amber-200/90" data-test="guardian-procedure-safety-stop">
      Safety stop — use a licensed tradesperson before continuing physical work.
    </p>
    <p v-if="procedure.status === 'completed'" class="text-green-400/90">Procedure complete.</p>
    <a
      v-if="procedure.print_path"
      :href="printUrl"
      target="_blank"
      rel="noopener"
      class="inline-block text-[10px] text-green-500 hover:text-green-300 underline"
      data-test="guardian-procedure-print"
    >
      Print checklist
    </a>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  procedure: { type: Object, default: null },
})

const apiBase = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

const printUrl = computed(() => {
  if (!props.procedure?.print_path) return '#'
  const path = props.procedure.print_path.startsWith('/')
    ? props.procedure.print_path
    : '/' + props.procedure.print_path
  return apiBase + path
})
</script>
