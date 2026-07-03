<template>
  <div
    class="p-4 sm:p-6 max-w-3xl mx-auto space-y-6 pb-24 md:pb-10 virtual-pi-root"
    :class="{ 'virtual-pi-print-mode': printMode }"
    data-test="virtual-pi-view"
  >
    <header class="space-y-2 virtual-pi-screen-only">
      <h1 class="text-xl font-bold text-white">Virtual Pi</h1>
      <p class="text-sm text-zinc-400 leading-relaxed max-w-2xl">
        See what's wired to each pin on your edge device — driven from the same wiring
        you edit on zone pages. Tap GPIO pins to wire; download config.yaml for the Pi.
      </p>
    </header>

    <header v-if="printMode" class="virtual-pi-print-only space-y-1 border-b border-zinc-700 pb-3">
      <h1 class="text-lg font-bold text-black">gr33n — Pi wiring sheet</h1>
      <p v-if="selectedDevice" class="text-sm text-zinc-700">
        {{ deviceLabel(selectedDevice) }}
        <span v-if="farmContext.selectedFarm?.name"> · {{ farmContext.selectedFarm.name }}</span>
      </p>
      <p class="text-xs text-zinc-600">Generated {{ printDate }}</p>
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
      <div class="flex flex-wrap items-end gap-3 virtual-pi-screen-only">
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
        <button
          type="button"
          class="text-xs border border-zinc-700 rounded-lg px-3 py-2 text-zinc-300 hover:border-green-600 hover:text-green-400"
          data-test="virtual-pi-download-config"
          :disabled="configDownloading || !selectedDeviceId"
          @click="downloadConfig"
        >
          {{ configDownloading ? 'Generating…' : 'Download config.yaml' }}
        </button>
        <button
          type="button"
          class="text-xs border border-zinc-700 rounded-lg px-3 py-2 text-zinc-300 hover:border-green-600 hover:text-green-400"
          data-test="virtual-pi-print"
          @click="openPrintView"
        >
          Print wiring sheet
        </button>
        <button
          v-if="canPushToPi"
          type="button"
          class="text-xs border border-zinc-700 rounded-lg px-3 py-2 text-zinc-300 hover:border-amber-600 hover:text-amber-300"
          data-test="virtual-pi-push-config"
          :disabled="pushConfigLoading || !selectedDeviceId"
          @click="pushConfigToPi"
        >
          {{ pushConfigLoading ? 'Notifying…' : 'Notify Pi to reload' }}
        </button>
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

      <div
        v-if="pushConfigMessage"
        class="rounded-lg border border-amber-800/50 bg-amber-950/20 px-3 py-2 text-xs text-amber-200 virtual-pi-screen-only"
        data-test="virtual-pi-push-ok"
      >
        {{ pushConfigMessage }}
      </div>

      <div
        v-if="wiringDrift === 'stale'"
        class="rounded-lg border border-amber-700/70 bg-amber-950/30 px-3 py-2 text-xs text-amber-200"
        data-test="virtual-pi-wiring-stale"
      >
        {{ wiringDriftLabel(wiringDrift) }}
      </div>
      <div
        v-else-if="wiringDrift === 'synced'"
        class="rounded-lg border border-green-800/50 bg-green-950/20 px-3 py-2 text-xs text-green-300 virtual-pi-screen-only"
        data-test="virtual-pi-wiring-synced"
      >
        {{ wiringDriftLabel(wiringDrift) }}
      </div>

      <VirtualPiBoard
        v-if="selectedDeviceId"
        :device-id="selectedDeviceId"
        :sensors="store.sensors"
        :actuators="store.actuators"
        :zones="store.zones"
        :devices="store.devices"
        @updated="onHardwareUpdated"
      />
    </template>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useFarmStore } from '../stores/farm.js'
import { useFarmContextStore } from '../stores/farmContext.js'
import VirtualPiBoard from '../components/VirtualPiBoard.vue'
import EmptyStateHint from '../components/EmptyStateHint.vue'
import { devicesWithWiring } from '../lib/piPinMap.js'
import { loadDeviceTaxonomy } from '../lib/deviceTaxonomy.js'
import { wiringDriftStatus, wiringDriftLabel } from '../lib/piConfigDrift.js'
import { deviceUsesPlatformSync } from '../lib/deviceConfigSync.js'
import api from '../api'

const store = useFarmStore()
const farmContext = useFarmContextStore()
const route = useRoute()
const loading = ref(true)
const loadError = ref('')
const selectedDeviceId = ref(null)
const configDownloading = ref(false)
const pushConfigLoading = ref(false)
const pushConfigMessage = ref('')
const expectedConfigSha = ref('')

const printMode = computed(() => route.query.print === '1')
const printDate = computed(() => new Date().toLocaleString())

const piDevices = computed(() => {
  const wired = devicesWithWiring(store.devices, store.sensors, store.actuators)
  if (wired.length) return wired
  return store.devices.filter((d) =>
    String(d.device_type || '').toLowerCase().includes('raspberry')
    || String(d.device_type || '').toLowerCase().includes('pi'),
  )
})

const selectedDevice = computed(() =>
  piDevices.value.find((d) => d.id === selectedDeviceId.value) || null,
)

const wiringDrift = computed(() =>
  wiringDriftStatus(selectedDevice.value, expectedConfigSha.value),
)

const canPushToPi = computed(() => deviceUsesPlatformSync(selectedDevice.value))

function deviceLabel(d) {
  const status = d.status === 'online' ? ' · online' : ''
  return `${d.name || d.device_uid || 'Device ' + d.id}${status}`
}

const apiBaseUrl = computed(() => {
  if (typeof window === 'undefined') return 'http://<api-lan-ip>:8080'
  return `${window.location.origin.replace(/:\d+$/, ':8080')}`
})

async function fetchExpectedConfigSha(deviceId) {
  if (!deviceId) {
    expectedConfigSha.value = ''
    return
  }
  try {
    const r = await api.get(`/devices/${deviceId}/pi-config`, {
      params: { base_url: apiBaseUrl.value },
    })
    expectedConfigSha.value = r.data?.config_sha256 || ''
  } catch {
    expectedConfigSha.value = ''
  }
}

async function pushConfigToPi() {
  if (!selectedDeviceId.value) return
  pushConfigLoading.value = true
  pushConfigMessage.value = ''
  loadError.value = ''
  try {
    const r = await api.post(`/devices/${selectedDeviceId.value}/push-config`)
    pushConfigMessage.value = r.data?.message || 'Pi notified — wiring reloads on next poll.'
    const fid = farmContext.farmId
    if (fid) await store.loadAll(fid)
  } catch (e) {
    loadError.value = e?.response?.data?.error || e?.message || 'Could not notify Pi'
  } finally {
    pushConfigLoading.value = false
  }
}

async function downloadConfig() {
  if (!selectedDeviceId.value) return
  configDownloading.value = true
  try {
    const r = await api.get(`/devices/${selectedDeviceId.value}/pi-config`, {
      params: { base_url: apiBaseUrl.value },
    })
    const yaml = r.data?.yaml || ''
    const filename = r.data?.filename || `config-device-${selectedDeviceId.value}.yaml`
    expectedConfigSha.value = r.data?.config_sha256 || expectedConfigSha.value
    const blob = new Blob([yaml], { type: 'text/yaml;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = filename
    a.click()
    URL.revokeObjectURL(url)
  } catch (e) {
    loadError.value = e?.response?.data?.error || e?.message || 'Config download failed'
  } finally {
    configDownloading.value = false
  }
}

function openPrintView() {
  const q = selectedDeviceId.value ? `?print=1&device=${selectedDeviceId.value}` : '?print=1'
  window.open(`/virtual-pi${q}`, '_blank', 'noopener')
}

async function onHardwareUpdated() {
  const fid = farmContext.farmId
  if (!fid) return
  try {
    await store.loadAll(fid)
    await fetchExpectedConfigSha(selectedDeviceId.value)
  } catch { /* best-effort refresh */ }
}

watch(piDevices, (list) => {
  if (!list.length) {
    selectedDeviceId.value = null
    return
  }
  const fromQuery = Number(route.query.device)
  if (fromQuery && list.some((d) => d.id === fromQuery)) {
    selectedDeviceId.value = fromQuery
    return
  }
  if (!list.some((d) => d.id === selectedDeviceId.value)) {
    selectedDeviceId.value = list[0].id
  }
}, { immediate: true })

watch(selectedDeviceId, (id) => {
  fetchExpectedConfigSha(id)
})

watch(printMode, (isPrint) => {
  if (isPrint && typeof window !== 'undefined') {
    window.addEventListener('load', () => window.print(), { once: true })
    setTimeout(() => window.print(), 500)
  }
}, { immediate: true })

onMounted(async () => {
  const fid = farmContext.farmId
  if (!fid) {
    loading.value = false
    return
  }
  try {
    await loadDeviceTaxonomy(api)
    if (!store.devices.length) await store.loadAll(fid)
    await fetchExpectedConfigSha(selectedDeviceId.value)
  } catch (e) {
    loadError.value = e?.message || 'Failed to load devices'
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.virtual-pi-print-only {
  display: none;
}

@media print {
  .virtual-pi-root {
    max-width: none;
    padding: 0.5in;
    color: #111;
  }

  .virtual-pi-screen-only {
    display: none !important;
  }

  .virtual-pi-print-only {
    display: block !important;
  }

  :global(body) {
    background: white !important;
  }

  :global(.sidebar),
  :global(.app-sidebar),
  :global(nav),
  :global(header.app-header) {
    display: none !important;
  }
}
</style>
