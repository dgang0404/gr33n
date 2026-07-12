<template>
  <section
    v-if="entries.length"
    class="rounded-xl border border-amber-900/50 bg-amber-950/20 px-4 py-3 space-y-2"
    data-test="farm-today-attention"
  >
    <div class="flex items-center justify-between gap-2">
      <h3 class="text-xs font-semibold text-amber-200/90 uppercase tracking-widest">
        Needs attention
      </h3>
      <span class="text-[10px] text-amber-300/70">{{ entries.length }} zone{{ entries.length === 1 ? '' : 's' }}</span>
    </div>
    <div class="flex flex-wrap gap-2">
      <button
        v-for="entry in entries"
        :key="entry.zone.id"
        type="button"
        class="min-h-[44px] px-3 py-2 rounded-lg border text-left text-sm transition-colors"
        :class="chipClass(entry.status)"
        :data-test="`farm-attention-chip-${entry.zone.id}`"
        :aria-label="`${entry.zone.name}: ${entry.summary}`"
        @click="$emit('select-zone', entry.zone, entry.status)"
      >
        <span class="font-medium text-white block truncate max-w-[200px] sm:max-w-none">{{ entry.zone.name }}</span>
        <span class="text-[11px] text-zinc-300 block truncate max-w-[220px] sm:max-w-none">{{ entry.summary }}</span>
      </button>
    </div>
  </section>
</template>

<script setup>
import { computed } from 'vue'
import { computeZoneVisualStatus } from '../lib/farmVisualStatus.js'
import { listAttentionZones, zoneAttentionSummary } from '../lib/zoneQuickActions.js'

const props = defineProps({
  zones: { type: Array, default: () => [] },
  sensors: { type: Array, default: () => [] },
  readings: { type: Object, default: () => ({}) },
  actuators: { type: Array, default: () => [] },
  tasks: { type: Array, default: () => [] },
  alerts: { type: Array, default: () => [] },
  schedules: { type: Array, default: () => [] },
  programs: { type: Array, default: () => [] },
  cropCycles: { type: Array, default: () => [] },
  fertigationEvents: { type: Array, default: () => [] },
})

defineEmits(['select-zone'])

const statusCtx = computed(() => ({
  sensors: props.sensors,
  readings: props.readings,
  actuators: props.actuators,
  tasks: props.tasks,
  alerts: props.alerts,
  schedules: props.schedules,
  programs: props.programs,
  cropCycles: props.cropCycles,
  fertigationEvents: props.fertigationEvents,
}))

function statusFor(zone) {
  return computeZoneVisualStatus({ zone, ...statusCtx.value })
}

const entries = computed(() =>
  listAttentionZones(props.zones, statusFor).map(({ zone, status }) => ({
    zone,
    status,
    summary: zoneAttentionSummary(status),
  })),
)

function chipClass(status) {
  if (status?.health === 'alert') {
    return 'border-red-800/70 bg-red-950/30 hover:border-red-700'
  }
  return 'border-amber-800/60 bg-amber-950/30 hover:border-amber-700'
}
</script>
