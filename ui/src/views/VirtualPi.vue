<template>
  <div
    class="p-4 sm:p-6 w-full space-y-6 pb-24 md:pb-10 virtual-pi-root"
    :class="{ 'virtual-pi-print-mode': printMode }"
    data-test="virtual-pi-view"
  >
    <header class="space-y-2 virtual-pi-screen-only">
      <p class="text-sm text-zinc-400 leading-relaxed max-w-5xl">
        See what's wired to each pin on your edge device — the same assignments you edit on zone pages.
        Tap GPIO pins to wire relays and sensors, then download config.yaml or notify the Pi to reload.
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
          <label class="text-[10px] text-zinc-500 block mb-1" for="virtual-pi-device" title="Choose which Raspberry Pi or edge device to configure">Edge device</label>
          <select
            id="virtual-pi-device"
            v-model.number="selectedDeviceId"
            class="w-full bg-zinc-900 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-zinc-200"
            data-test="virtual-pi-device-select"
            title="Select the Raspberry Pi or edge device to view and manage GPIO wiring"
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
          title="Export the current wiring configuration (GPIO assignments) as config.yaml to use with your Pi"
          @click="downloadConfig"
        >
          {{ configDownloading ? 'Generating…' : 'Download config.yaml' }}
        </button>
        <button
          type="button"
          class="text-xs border border-zinc-700 rounded-lg px-3 py-2 text-zinc-300 hover:border-green-600 hover:text-green-400"
          data-test="virtual-pi-print"
          title="Print a labeled wiring reference sheet showing all GPIO pin assignments"
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
          title="Tell this Pi to reload its wiring configuration on the next poll (typically within 30s). Does not affect IP reconnection, which happens automatically."
          @click="pushConfigToPi"
        >
          {{ pushConfigLoading ? 'Notifying…' : 'Notify Pi to reload' }}
        </button>
        <router-link
          v-nav-hint="'/hardware'"
          :to="{ path: '/hardware', query: { tab: 'board' } }"
          class="text-xs text-zinc-500 hover:text-green-400 pb-2"
          title="Switch to list view of all connected devices and their sensors/actuators"
        >
          List view →
        </router-link>
        <router-link
          v-nav-hint="'/pi-setup'"
          :to="{ name: 'pi-setup' }"
          class="text-xs text-zinc-500 hover:text-green-400 pb-2"
          title="Help registering a new Raspberry Pi and getting it online"
        >
          Pi setup →
        </router-link>
      </div>

      <div
        v-if="pushConfigMessage"
        class="rounded-lg border border-amber-800/50 bg-amber-950/20 px-3 py-2 text-xs text-amber-200 virtual-pi-screen-only"
        data-test="virtual-pi-push-ok"
        title="Notification has been sent. The Pi will reload its wiring configuration on the next config poll (usually within 30 seconds)."
      >
        {{ pushConfigMessage }}
      </div>

      <div
        v-if="selectedDeviceId && validation"
        class="rounded-lg border px-3 py-3 text-xs space-y-2 virtual-pi-screen-only"
        :class="validationBannerClass(validation.status)"
        data-test="virtual-pi-validation-banner"
        :title="`Status: ${validation.status} — ${validation.hint}`"
      >
        <p class="font-semibold" data-test="virtual-pi-validation-title" :title="validation.hint">{{ validation.title }}</p>
        <ul class="space-y-1">
          <li
            v-for="item in validation.checklist"
            :key="item.id"
            :class="item.ok ? 'text-green-300/90' : 'text-amber-200/80'"
            :data-test="`virtual-pi-validation-${item.id}`"
            :title="item.id === 'wiring_assigned' ? 'At least one GPIO pin or relay channel must be wired' : item.id === 'config_ready' ? 'Configuration file is ready to export' : item.id === 'pi_connected' ? 'Pi is currently online and reachable' : ''"
          >
            {{ item.ok ? '✓' : '○' }} {{ item.label }}
          </li>
        </ul>
        <p class="text-[10px] opacity-90 leading-snug">{{ validation.hint }}</p>
      </div>

      <div
        v-if="wiringDrift === 'stale'"
        class="rounded-lg border border-amber-700/70 bg-amber-950/30 px-3 py-2 text-xs text-amber-200"
        data-test="virtual-pi-wiring-stale"
        title="Pi's wiring is out of sync — click 'Notify Pi to reload' to update it"
      >
        {{ wiringDriftLabel(wiringDrift) }}
      </div>
      <div
        v-else-if="wiringDrift === 'synced'"
        class="rounded-lg border border-green-800/50 bg-green-950/20 px-3 py-2 text-xs text-green-300 virtual-pi-screen-only"
        data-test="virtual-pi-wiring-synced"
        title="Pi's wiring configuration matches what you configured here"
      >
        {{ wiringDriftLabel(wiringDrift) }}
      </div>

      <section
        v-if="selectedDeviceId"
        class="bg-zinc-900 border border-zinc-800 rounded-xl p-4 space-y-3 virtual-pi-screen-only"
        data-test="virtual-pi-connectivity"
      >
        <div class="flex items-center justify-between gap-2 flex-wrap">
          <h3 class="text-sm font-semibold text-white">Connectivity</h3>
          <span class="text-[10px] text-zinc-500 font-mono" data-test="virtual-pi-current-ip">
            Current IP: {{ selectedDevice?.ip_address || 'unknown' }}
          </span>
        </div>
        <p class="text-[10px] text-zinc-500 leading-snug">
          Editing this only corrects the platform's record — it does not dial the Pi. The Pi always
          initiates contact (commands/config are polled, not pushed), so a live Pi's next heartbeat
          auto-corrects this anyway. Use the override when a device won't be checking in soon and you
          need the record to match a new DHCP reservation or replacement.
        </p>
        <div class="flex flex-wrap items-center gap-2">
          <input
            v-model="ipOverrideInput"
            type="text"
            placeholder="192.168.1.55"
            class="bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-xs text-zinc-200 font-mono w-40"
            data-test="virtual-pi-ip-override-input"
            title="New IP address to record for this device"
          />
          <button
            type="button"
            class="text-xs border border-zinc-700 rounded-lg px-3 py-2 text-zinc-300 hover:border-amber-600 hover:text-amber-300 disabled:opacity-40"
            data-test="virtual-pi-ip-override-save"
            :disabled="ipOverrideSaving || !ipOverrideInput.trim()"
            title="Manually set this device's recorded IP address"
            @click="saveIPOverride"
          >
            {{ ipOverrideSaving ? 'Saving…' : 'Set IP' }}
          </button>
        </div>
        <p v-if="ipOverrideMessage" class="text-[10px] text-green-400" data-test="virtual-pi-ip-override-ok">{{ ipOverrideMessage }}</p>
        <p v-if="ipOverrideError" class="text-[10px] text-red-400" data-test="virtual-pi-ip-override-error">{{ ipOverrideError }}</p>

        <DeviceIPHistoryPanel :key="selectedDeviceId" :device-id="selectedDeviceId" />
      </section>

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
import DeviceIPHistoryPanel from '../components/DeviceIPHistoryPanel.vue'
import { devicesWithWiring } from '../lib/piPinMap.js'
import { loadDeviceTaxonomy } from '../lib/deviceTaxonomy.js'
import { wiringDriftStatus, wiringDriftLabel } from '../lib/piConfigDrift.js'
import { computeVirtualPiValidation, validationBannerClass } from '../lib/virtualPiValidation.js'
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
const configDownloaded = ref(false)
const ipOverrideInput = ref('')
const ipOverrideSaving = ref(false)
const ipOverrideMessage = ref('')
const ipOverrideError = ref('')

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

const validation = computed(() => {
  if (!selectedDevice.value) return null
  return computeVirtualPiValidation({
    device: selectedDevice.value,
    sensors: store.sensors,
    actuators: store.actuators,
    expectedConfigSha: expectedConfigSha.value,
    configDownloaded: configDownloaded.value,
  })
})

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
    configDownloaded.value = true
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
  configDownloaded.value = false
  ipOverrideInput.value = ''
  ipOverrideMessage.value = ''
  ipOverrideError.value = ''
  fetchExpectedConfigSha(id)
})

async function saveIPOverride() {
  if (!selectedDeviceId.value || !ipOverrideInput.value.trim()) return
  ipOverrideSaving.value = true
  ipOverrideMessage.value = ''
  ipOverrideError.value = ''
  try {
    await api.patch(`/devices/${selectedDeviceId.value}/ip-address`, {
      ip_address: ipOverrideInput.value.trim(),
    })
    ipOverrideMessage.value = 'IP address updated — see history below.'
    ipOverrideInput.value = ''
    const fid = farmContext.farmId
    if (fid) await store.loadAll(fid)
  } catch (e) {
    ipOverrideError.value = e?.response?.data?.error || e?.message || 'Could not update IP address'
  } finally {
    ipOverrideSaving.value = false
  }
}

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
