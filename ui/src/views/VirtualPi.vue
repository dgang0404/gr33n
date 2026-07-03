<template>
  <div class="p-4 sm:p-6 max-w-3xl mx-auto space-y-6 pb-24 md:pb-10" data-test="virtual-pi-view">
    <header class="space-y-2">
      <h1 class="text-xl font-bold text-white">Virtual Pi</h1>
      <p class="text-sm text-zinc-400 leading-relaxed max-w-2xl">
        See what's wired to each pin on your edge device — driven from the same wiring
        you edit on zone pages. Read-only here; use zone hardware panels to change assignments.
      </p>
    </header>

    <div v-if="loading" class="text-sm text-zinc-500">Loading farm hardware…</div>
    <div v-else-if="loadError" class="text-sm text-red-400">{{ loadError }}</div>

    <EmptyStateHint
      v-else-if="!piDevices.length"
      reason="no_telemetry"
      message="No Pi devices with wiring yet. Register a Pi and assign GPIO or relay channels from a zone."
      action-label="Pi setup guide"
      :action-to="{ name: 'pi-setup' }"
    />

    <template v-else>
      <div class="flex flex-wrap items-end gap-3">
        <div class="flex-1 min-w-[12rem]">
          <label class="text-[10px] text-zinc-500 block mb-1" for="virtual-pi-device">Edge device</label>
          <select
            id="virtual-pi-device"
            v-model.number="selectedDeviceId"
            class="w-full bg-zinc-900 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-zinc-200"
            data-test="virtual-pi-device-select"
          >
            <option v-for="d in piDevices" :key="d.id" :value="d.id">
              {{ deviceLabel(d) }}
            </option>
          </select>
        </div>
        <router-link
          v-nav-hint="'/hardware'"
          :to="{ path: '/hardware', query: { tab: 'board' } }"
          class="text-xs text-zinc-500 hover:text-green-400 pb-2"
        >
          List view →
        </router-link>
        <router-link
          v-nav-hint="'/pi-setup'"
          :to="{ name: 'pi-setup' }"
          class="text-xs text-zinc-500 hover:text-green-400 pb-2"
        >
          Pi setup →
        </router-link>
      </div>

      <VirtualPiBoard
        v-if="selectedDeviceId"
        :device-id="selectedDeviceId"
        :sensors="store.sensors"
        :actuators="store.actuators"
        :zones="store.zones"
      />
    </template>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useFarmStore } from '../stores/farm.js'
import { useFarmContextStore } from '../stores/farmContext.js'
import VirtualPiBoard from '../components/VirtualPiBoard.vue'
import EmptyStateHint from '../components/EmptyStateHint.vue'
import { devicesWithWiring } from '../lib/piPinMap.js'

const store = useFarmStore()
const farmContext = useFarmContextStore()
const loading = ref(true)
const loadError = ref('')
const selectedDeviceId = ref(null)

const piDevices = computed(() => {
  const wired = devicesWithWiring(store.devices, store.sensors, store.actuators)
  if (wired.length) return wired
  return store.devices.filter((d) =>
    String(d.device_type || '').toLowerCase().includes('raspberry')
    || String(d.device_type || '').toLowerCase().includes('pi'),
  )
})

function deviceLabel(d) {
  const status = d.status === 'online' ? ' · online' : ''
  return `${d.name || d.device_uid || 'Device ' + d.id}${status}`
}

watch(piDevices, (list) => {
  if (!list.length) {
    selectedDeviceId.value = null
    return
  }
  if (!list.some((d) => d.id === selectedDeviceId.value)) {
    selectedDeviceId.value = list[0].id
  }
}, { immediate: true })

onMounted(async () => {
  const fid = farmContext.farmId
  if (!fid) {
    loading.value = false
    return
  }
  try {
    if (!store.devices.length) await store.loadAll(fid)
  } catch (e) {
    loadError.value = e?.message || 'Failed to load devices'
  } finally {
    loading.value = false
  }
})
</script>
