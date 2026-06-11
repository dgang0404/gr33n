<template>
  <span
    v-nav-hint="hintPath"
    class="inline-flex cursor-default"
    :title="hintTitle"
    data-test="hardware-wiring-badge"
  >
    <span
      v-if="label"
      class="inline-flex items-center gap-1 text-[10px] font-medium px-2 py-0.5 rounded-full bg-zinc-800/80 text-zinc-300 border border-zinc-700/80"
    >
      <span aria-hidden="true">🔌</span>
      <span class="truncate max-w-[14rem]">{{ label }}</span>
    </span>
    <span
      v-else-if="showEmpty"
      class="inline-flex items-center text-[10px] text-zinc-500 italic"
    >
      Not wired yet
    </span>
  </span>
</template>

<script setup>
import { computed } from 'vue'
import { formatWiringLabel, resolveWiring } from '../lib/hardwareWiring.js'

const props = defineProps({
  entity: { type: Object, default: null },
  wiring: { type: Object, default: null },
  showEmpty: { type: Boolean, default: false },
  hintPath: { type: String, default: '/zones' },
})

const label = computed(() => {
  const w = props.wiring ?? resolveWiring(props.entity)
  return formatWiringLabel(w)
})

const hintTitle = computed(() => {
  if (label.value) return `${label.value} — wiring lives on the zone page`
  return 'Not wired yet — edit wiring on the zone page'
})
</script>
