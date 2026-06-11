<template>
  <p
    v-if="line"
    class="text-[10px] text-green-500/90 mt-1 font-medium truncate"
    :title="line"
    data-test="alert-hardware-line"
  >
    <span aria-hidden="true">🔌</span> {{ line }}
  </p>
</template>

<script setup>
import { computed } from 'vue'
import { formatAlertHardwareLine } from '../lib/alertHardwareContext.js'

const props = defineProps({
  alert: { type: Object, required: true },
  sensors: { type: Array, default: () => [] },
  actuators: { type: Array, default: () => [] },
})

const line = computed(() =>
  formatAlertHardwareLine(props.alert, {
    sensors: props.sensors,
    actuators: props.actuators,
  }),
)
</script>
