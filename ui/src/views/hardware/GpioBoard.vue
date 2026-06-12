<template>
  <div class="p-4 sm:p-6 max-w-4xl mx-auto space-y-6 pb-24 md:pb-10" data-test="gpio-board">
    <header class="space-y-2">
      <h1 class="text-xl font-bold text-white">GPIO board</h1>
      <p class="text-sm text-zinc-400 leading-relaxed max-w-2xl">
        Live pin and relay map for this farm — what each channel is wired to, which zone it serves, and current state.
        For typical HAT wiring diagrams, see the
        <router-link v-nav-hint="'/hardware'" :to="{ path: '/hardware', query: { tab: 'reference' } }" class="text-green-500 hover:text-green-400">Reference</router-link>
        tab.
      </p>
    </header>

    <div v-if="loading" class="text-sm text-zinc-500">Loading devices and wiring…</div>
    <div v-else-if="loadError" class="text-sm text-red-400">{{ loadError }}</div>

    <div v-else-if="!boardDevices.length" class="rounded-xl border border-zinc-800 bg-zinc-950/50 px-4 py-5 text-center space-y-2">
      <p class="text-sm text-zinc-500">No wiring set up yet.</p>
      <p class="text-xs text-zinc-600">
        Register a Pi under
        <router-link v-nav-hint="'/hardware'" :to="{ path: '/hardware', query: { tab: 'devices' } }" class="text-green-600 hover:text-green-400">Pi devices</router-link>,
        then assign relay channels or GPIO pins on actuators and sensors.
      </p>
    </div>

    <div v-else class="space-y-6">
      <section
        v-for="device in boardDevices"
        :key="device.id"
        class="bg-zinc-900 border border-zinc-800 rounded-xl overflow-hidden"
        :data-test="'gpio-board-device-' + device.id"
      >
        <div class="flex flex-wrap items-center gap-3 px-4 py-3 border-b border-zinc-800 bg-zinc-950/40">
          <span class="text-base">🖥️</span>
          <span class="text-sm font-semibold text-white">{{ deviceLabel(device) }}</span>
          <span class="text-[10px] font-mono text-zinc-500">{{ device.device_uid }}</span>
          <span
            class="ml-auto text-[10px] px-2 py-0.5 rounded-full font-medium"
            :class="device.status === 'online' ? 'bg-green-900/40 text-green-400' : 'bg-zinc-800 text-zinc-500'"
          >{{ device.status || 'offline' }}</span>
          <router-link
            v-nav-hint="'/hardware'"
            :to="{ path: '/hardware', query: { tab: 'devices' } }"
            class="text-[10px] text-green-400 hover:text-green-300 border border-zinc-700 rounded px-2 py-0.5"
          >Pi devices</router-link>
        </div>

        <div v-if="relayRows(device.id).length" class="px-4 py-3 border-b border-zinc-800/60">
          <p class="text-[10px] uppercase tracking-wide text-zinc-600 mb-2">Relay channels</p>
          <div class="space-y-1">
            <div
              v-for="row in relayRows(device.id)"
              :key="'relay-' + row.key"
              class="flex items-center gap-3 rounded-lg px-3 py-2 bg-zinc-950/30"
              data-test="gpio-board-relay-row"
            >
              <span class="font-mono text-[11px] text-gr33n-400 shrink-0 w-20">{{ row.label }}</span>
              <span class="text-xs text-zinc-200 truncate flex-1">{{ row.actuator.name }}</span>
              <span class="text-[10px] text-zinc-500 capitalize">{{ row.actuator.actuator_type }}</span>
              <span class="text-[10px] text-zinc-500">{{ zoneName(row.actuator.zone_id) }}</span>
              <span
                class="text-[10px] px-1.5 py-0.5 rounded"
                :class="row.actuator.current_state_text === 'on' ? 'bg-green-900/50 text-green-300' : 'bg-zinc-800 text-zinc-500'"
              >{{ row.actuator.current_state_text || 'off' }}</span>
            </div>
          </div>
        </div>

        <div v-if="gpioActuatorRows(device.id).length" class="px-4 py-3 border-b border-zinc-800/60">
          <p class="text-[10px] uppercase tracking-wide text-zinc-600 mb-2">Direct GPIO actuators</p>
          <div class="space-y-1">
            <div
              v-for="row in gpioActuatorRows(device.id)"
              :key="'gpio-act-' + row.actuator.id"
              class="flex items-center gap-3 rounded-lg px-3 py-2 bg-zinc-950/30"
            >
              <span class="font-mono text-[11px] text-gr33n-400 shrink-0 w-20">{{ row.label }}</span>
              <span class="text-xs text-zinc-200 truncate flex-1">{{ row.actuator.name }}</span>
              <span class="text-[10px] text-zinc-500">{{ zoneName(row.actuator.zone_id) }}</span>
            </div>
          </div>
        </div>

        <div v-if="sensorRows(device.id).length" class="px-4 py-3">
          <p class="text-[10px] uppercase tracking-wide text-zinc-600 mb-2">Sensor pins</p>
          <div class="space-y-1">
            <router-link
              v-for="row in sensorRows(device.id)"
              :key="'sensor-' + row.sensor.id"
              v-nav-hint="'/sensors'"
              :to="{ name: 'sensor-detail', params: { id: row.sensor.id } }"
              class="flex items-center gap-3 rounded-lg px-3 py-2 hover:bg-zinc-800/60 transition-colors"
            >
              <span class="font-mono text-[11px] text-blue-400 shrink-0 w-20 truncate">{{ row.label }}</span>
              <span class="text-xs text-zinc-200 truncate flex-1">{{ row.sensor.name }}</span>
            </router-link>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useFarmStore } from '../../stores/farm'
import { formatEntityHardwareLabel, formatWiringLabel, resolveWiring } from '../../lib/hardwareWiring.js'

const store = useFarmStore()
const loading = ref(true)
const loadError = ref('')

onMounted(async () => {
  try {
    await store.loadAll()
  } catch (e) {
    loadError.value = e?.message || 'Failed to load farm hardware'
  } finally {
    loading.value = false
  }
})

function channelFromActuator(a) {
  const hi = a?.hardware_identifier
  if (hi == null || hi === '') return null
  const m = String(hi).match(/(\d+)$/)
  if (!m) return null
  const n = parseInt(m[1], 10)
  return Number.isFinite(n) && n >= 0 ? n : null
}

const relayByDevice = computed(() => {
  const map = {}
  for (const a of store.actuators) {
    if (!a.device_id) continue
    const ch = channelFromActuator(a)
    if (ch == null) continue
    if (!map[a.device_id]) map[a.device_id] = {}
    map[a.device_id][ch] = a
  }
  return map
})

const gpioActuatorsByDevice = computed(() => {
  const map = {}
  for (const a of store.actuators) {
    if (channelFromActuator(a) != null) continue
    const w = resolveWiring(a)
    const devId = w?.device_id ?? a.device_id
    if (!devId) continue
    if (!map[devId]) map[devId] = []
    map[devId].push({ actuator: a, label: formatWiringLabel(w) || formatEntityHardwareLabel(a) || 'GPIO' })
  }
  return map
})

const sensorsByDevice = computed(() => {
  const map = {}
  for (const s of store.sensors) {
    const w = resolveWiring(s)
    if (!w?.device_id) continue
    if (!map[w.device_id]) map[w.device_id] = []
    map[w.device_id].push({ sensor: s, label: formatWiringLabel(w) || 'wired' })
  }
  for (const rows of Object.values(map)) {
    rows.sort((a, b) => (a.label > b.label ? 1 : -1))
  }
  return map
})

const boardDevices = computed(() => {
  const ids = new Set([
    ...Object.keys(relayByDevice.value).map(Number),
    ...Object.keys(gpioActuatorsByDevice.value).map(Number),
    ...Object.keys(sensorsByDevice.value).map(Number),
  ])
  return store.devices.filter((d) => ids.has(d.id))
})

function deviceLabel(d) {
  return d.name || d.device_uid || `Device ${d.id}`
}

function zoneName(zoneId) {
  if (!zoneId) return ''
  const z = store.zones.find((x) => Number(x.id) === Number(zoneId))
  return z?.name ? `· ${z.name}` : ''
}

function relayRows(deviceId) {
  const chMap = relayByDevice.value[deviceId] || {}
  return Object.entries(chMap)
    .map(([ch, actuator]) => ({
      key: ch,
      label: `ch ${ch}`,
      actuator,
    }))
    .sort((a, b) => Number(a.key) - Number(b.key))
}

function gpioActuatorRows(deviceId) {
  return gpioActuatorsByDevice.value[deviceId] || []
}

function sensorRows(deviceId) {
  return sensorsByDevice.value[deviceId] || []
}
</script>
