<template>
  <section
    v-if="visible"
    class="flex items-center gap-2 overflow-x-auto pb-1 -mx-1 px-1"
    role="group"
    aria-label="Filter zones"
    data-test="farm-today-zone-filter-bar"
  >
    <button
      v-for="f in filters"
      :key="f.id"
      type="button"
      class="shrink-0 min-h-[36px] px-3 py-1.5 rounded-full text-xs font-medium border transition-colors"
      :class="chipClass(f.id)"
      :aria-pressed="f.id === modelValue"
      :data-test="`farm-today-zone-filter-${f.id}`"
      @click="select(f.id)"
    >
      {{ f.label }}
      <span v-if="counts[f.id] != null" class="ml-1 text-[10px] opacity-70">{{ counts[f.id] }}</span>
    </button>
  </section>
</template>

<script setup>
import { computed } from 'vue'
import {
  TODAY_ZONE_FILTERS,
  countZonesPerFilter,
  shouldShowTodayZoneFilterBar,
} from '../lib/farmTodayZoneFilter.js'

const props = defineProps({
  zones: { type: Array, default: () => [] },
  getStatus: { type: Function, required: true },
  modelValue: { type: String, default: 'all' },
})

const emit = defineEmits(['update:modelValue'])

const filters = TODAY_ZONE_FILTERS
const counts = computed(() => countZonesPerFilter(props.zones, props.getStatus))
const visible = computed(() => shouldShowTodayZoneFilterBar(props.zones.length))

function select(id) {
  if (id === props.modelValue) return
  emit('update:modelValue', id)
}

function chipClass(id) {
  return id === props.modelValue
    ? 'bg-green-900/50 text-green-300 border-green-700'
    : 'bg-zinc-800 text-zinc-400 border-zinc-700 hover:border-zinc-600'
}
</script>
