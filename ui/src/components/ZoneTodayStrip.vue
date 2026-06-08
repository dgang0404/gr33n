<template>
  <div
    class="bg-zinc-900 border border-zinc-800 rounded-xl p-4"
    data-test="zone-today-strip"
  >
    <div class="flex items-center justify-between gap-2 mb-3">
      <h2 class="text-sm font-semibold text-white">Today in this zone</h2>
      <span class="text-zinc-600 text-xs">What matters now</span>
    </div>
    <div class="flex flex-wrap gap-2">
      <component
        :is="chip.to ? 'router-link' : 'div'"
        v-for="chip in chips"
        :key="chip.id"
        v-bind="chip.to ? { to: chip.to } : {}"
        class="flex items-center gap-2 rounded-lg border px-3 py-2 min-w-[8.5rem]"
        :class="[chipClass(chip), chip.to ? 'transition-colors hover:border-gr33n-800/80' : '']"
        :data-test="`zone-today-chip-${chip.id}`"
      >
        <span class="text-base leading-none" aria-hidden="true">{{ chip.icon }}</span>
        <div class="min-w-0">
          <p class="text-[10px] uppercase tracking-wide text-zinc-500">{{ chip.label }}</p>
          <p class="text-sm text-zinc-100 font-medium truncate" :title="chip.detail || chip.value">
            {{ chip.value }}
          </p>
          <p v-if="chip.detail" class="text-[10px] text-zinc-600 truncate">{{ chip.detail }}</p>
        </div>
      </component>
    </div>
  </div>
</template>

<script setup>
defineProps({
  chips: { type: Array, default: () => [] },
})

function chipClass(chip) {
  if (chip.tone === 'warn') return 'border-amber-800/60 bg-amber-950/30'
  if (chip.tone === 'ok') return 'border-green-900/50 bg-green-950/20'
  if (chip.tone === 'muted') return 'border-zinc-800 bg-zinc-950/50'
  return 'border-zinc-800 bg-zinc-950/60'
}
</script>
