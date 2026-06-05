<template>
  <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4" data-test="zone-comfort-targets">
    <div class="flex items-center justify-between gap-2 mb-3 flex-wrap">
      <div>
        <h3 class="text-sm font-semibold text-white">Comfort targets</h3>
        <p class="text-zinc-600 text-xs mt-0.5">How comfortable this room should be — min, ideal, and max.</p>
      </div>
      <router-link to="/comfort-targets" class="text-xs text-zinc-500 hover:text-green-400">
        All rooms →
      </router-link>
    </div>

    <ComfortBandEditor
      :need="need"
      :zone-id="zoneId"
      :farm-id="farmId"
      :sensors="sensors"
      :setpoints="setpoints"
      :readings="readings"
      @updated="$emit('updated')"
    />
  </div>
</template>

<script setup>
import ComfortBandEditor from './ComfortBandEditor.vue'

defineProps({
  need: { type: String, required: true },
  zoneId: { type: Number, required: true },
  farmId: { type: Number, required: true },
  sensors: { type: Array, default: () => [] },
  setpoints: { type: Array, default: () => [] },
  readings: { type: Object, default: () => ({}) },
})

defineEmits(['updated'])
</script>
