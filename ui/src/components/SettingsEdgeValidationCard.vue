<template>
  <section
    v-if="farmId && piDeviceCount > 0"
    class="bg-zinc-800 border border-zinc-700 rounded-xl p-5 mb-5"
    data-test="settings-edge-validation"
  >
    <h2 class="text-white font-semibold mb-2 flex items-center gap-2">
      <span>🔌</span> Field validation — Virtual Pi path
    </h2>
    <p class="text-xs text-zinc-500 mb-4 leading-relaxed">
      Wire on <router-link v-nav-hint="'/virtual-pi'" to="/virtual-pi" class="text-gr33n-500 hover:underline">Virtual Pi</router-link>,
      dry-run on the LED simulation rig, then promote to live relays.
      Guide: <code class="text-zinc-400">docs/virtual-pi-field-validation-path.md</code>
    </p>

    <ol class="text-xs text-zinc-400 space-y-2 list-decimal list-inside mb-4">
      <li>Virtual Pi — download <code class="text-zinc-500">config.yaml</code></li>
      <li>
        <code class="text-zinc-500">docs/pi-light-simulation-runbook.md</code> — Demo A moisture loop
      </li>
      <li>Phase 31 edge checklist — live sensor on dashboard</li>
    </ol>

    <div class="flex flex-wrap gap-2 text-xs">
      <router-link
        v-nav-hint="'/virtual-pi'"
        to="/virtual-pi"
        class="px-3 py-1.5 rounded-lg border border-zinc-600 text-zinc-200 hover:border-green-700 hover:text-green-400"
        data-test="settings-edge-validation-virtual-pi"
      >
        Open Virtual Pi
      </router-link>
      <router-link
        v-nav-hint="'/pi-setup'"
        :to="{ name: 'pi-setup' }"
        class="px-3 py-1.5 rounded-lg border border-zinc-600 text-zinc-200 hover:border-green-700 hover:text-green-400"
        data-test="settings-edge-validation-pi-setup"
      >
        Pi setup guide
      </router-link>
    </div>

    <p class="text-[10px] text-zinc-600 mt-3">
      {{ piDeviceCount }} Pi device{{ piDeviceCount === 1 ? '' : 's' }} on this farm ·
      {{ wiredCount }} with wiring assigned
    </p>
  </section>
</template>

<script setup>
import { computed } from 'vue'
import { useFarmContextStore } from '../stores/farmContext'
import { useFarmStore } from '../stores/farm'
import { devicesWithWiring } from '../lib/piPinMap.js'

const farmContext = useFarmContextStore()
const farmStore = useFarmStore()

const farmId = computed(() => farmContext.farmId)

const piDevices = computed(() =>
  farmStore.devices.filter((d) => {
    const t = String(d.device_type || '').toLowerCase()
    return t.includes('raspberry') || t.includes('pi') || t.includes('relay')
  }),
)

const piDeviceCount = computed(() => piDevices.value.length)

const wiredCount = computed(() =>
  devicesWithWiring(farmStore.devices, farmStore.sensors, farmStore.actuators).length,
)
</script>
