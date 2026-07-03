<template>
  <div
    v-if="steps.length"
    class="rounded-lg border border-zinc-800 bg-zinc-950/60 p-3 space-y-2"
    data-test="driver-hookup-steps"
  >
    <h3 class="text-xs font-semibold text-zinc-300">
      Hookup · {{ driverLabel }}
    </h3>
    <ol class="space-y-2 text-xs">
      <li
        v-for="(step, i) in steps"
        :key="step.role + '-' + i"
        class="flex gap-2"
      >
        <span class="text-zinc-600 shrink-0">{{ i + 1 }}.</span>
        <span>
          <span class="font-mono text-amber-400/90">{{ step.wire }}</span>
          <span class="text-zinc-500"> → </span>
          <span class="text-zinc-300">{{ step.to }}</span>
        </span>
      </li>
    </ol>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { hookupStepsForDriver, wiringSourceForEntity } from '../lib/driverHookups.js'
import { getDeviceTaxonomy } from '../lib/deviceTaxonomy.js'
import { driverHookupsFromTaxonomy } from '../lib/driverHookups.js'

const props = defineProps({
  driver: { type: String, default: '' },
})

const hookups = computed(() => driverHookupsFromTaxonomy(getDeviceTaxonomy()))

const steps = computed(() => hookupStepsForDriver(hookups.value, props.driver))

const driverLabel = computed(() => {
  const d = String(props.driver || '').replace(/_/g, ' ')
  return d || 'driver'
})
</script>
