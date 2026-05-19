<!--
  MetricChip — small labelled value used by CropCycleSummary and
  CropCycleCompare. Keeps the grid cells consistent without leaking
  Tailwind utilities into every parent.
-->
<template>
  <div class="flex flex-col gap-0.5">
    <p class="text-zinc-500 text-[10px] uppercase tracking-wide">{{ label }}</p>
    <p class="text-white text-sm font-mono">
      <span>{{ value }}</span><span v-if="suffix" class="text-zinc-400">{{ suffix }}</span>
    </p>
  </div>
</template>

<script setup>
defineProps({
  label: { type: String, required: true },
  value: { type: [String, Number], required: true },
  suffix: { type: String, default: '' },
})
</script>
