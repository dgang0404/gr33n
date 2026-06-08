<template>
  <p class="text-zinc-600 text-xs" data-test="zone-connection-pipeline">
    How it connects:
    <template v-for="(seg, i) in segments" :key="seg.id">
      <span v-if="i > 0" class="text-zinc-700" aria-hidden="true"> → </span>
      <button
        type="button"
        v-nav-hint="seg.hint"
        class="font-medium text-zinc-400 hover:text-green-400 hover:underline motion-reduce:hover:no-underline focus:outline-none focus-visible:text-green-400"
        :data-test="`pipeline-segment-${seg.id}`"
      >
        {{ seg.label }}
      </button>
    </template>
  </p>
</template>

<script setup>
import { computed } from 'vue'
import {
  buildZoneConnectionSegments,
  resolvePipelineDeviceHint,
} from '../lib/zoneConnectionPipeline.js'

const props = defineProps({
  need: { type: String, default: '' },
  devices: { type: Array, default: () => [] },
})

const segments = computed(() =>
  buildZoneConnectionSegments({
    need: props.need,
    deviceHint: resolvePipelineDeviceHint(props.devices),
  }),
)
</script>
