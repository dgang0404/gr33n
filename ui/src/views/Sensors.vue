<template>
  <div class="p-6">
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-xl font-semibold text-white">Sensors</h1>
      <span class="text-xs text-zinc-500">{{ store.sensors.length }} total</span>
    </div>

    <div v-if="store.loading" class="text-zinc-400 text-sm">Loading sensors…</div>
    <div v-else-if="!store.sensors.length" class="text-zinc-500 text-sm">No sensors found.</div>

    <div v-else class="overflow-hidden rounded-xl border border-zinc-800">
      <table class="w-full text-sm">
        <thead class="bg-zinc-900 text-zinc-400 text-xs uppercase tracking-wide">
          <tr>
            <th class="px-4 py-3 text-left">Sensor</th>
            <th class="px-4 py-3 text-left">Type</th>
            <th class="px-4 py-3 text-left">Zone</th>
            <th class="px-4 py-3 text-left">Last Reading</th>
            <th class="px-4 py-3 text-left">Status</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-zinc-800/60">
          <tr
            v-for="sensor in store.sensors"
            :key="sensor.id"
            class="bg-zinc-950 hover:bg-zinc-900/60 transition-colors"
          >
            <td class="px-4 py-3">
              <router-link
                :to="{ name: 'sensor-detail', params: { id: sensor.id } }"
                class="text-white font-medium hover:text-green-400"
              >
                {{ sensor.name }}
              </router-link>
            </td>
            <td class="px-4 py-3 text-zinc-300">
              <span class="flex items-center gap-1.5">
                <span>{{ sensorIcon(sensor.sensor_type) }}</span>
                <span class="capitalize">{{ sensor.sensor_type }}</span>
              </span>
            </td>
            <td class="px-4 py-3 text-zinc-400">{{ zoneName(sensor.zone_id) }}</td>
            <td class="px-4 py-3 font-mono text-zinc-200 tabular-nums">
              {{ formatReading(sensor.id) }}
            </td>
            <td class="px-4 py-3">
              <span :class="statusBadge(store.sensorStatus(sensor.id))"
                class="text-xs font-medium px-2 py-0.5 rounded-full">
                {{ store.sensorStatus(sensor.id) }}
              </span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { useFarmStore } from '../stores/farm'
import { useFarmContextStore } from '../stores/farmContext'

const store = useFarmStore()
const farmContext = useFarmContextStore()
onMounted(async () => {
  if (!store.sensors.length && farmContext.farmId) await store.loadAll(farmContext.farmId)
  store.refreshReadings()
})

function zoneName(zoneId) {
  if (!zoneId) return '—'
  return store.zones.find(z => z.id === zoneId)?.name ?? `Zone ${zoneId}`
}
function formatReading(sensorId) {
  const r = store.readings[sensorId]
  if (!r) return 'NO DATA'
  const val = r.value_normalized ?? r.value_raw
  if (val == null) return 'NO DATA'
  const num = parseFloat(val)
  return isNaN(num) ? String(val) : num.toFixed(2)
}
const ICONS = { temperature:'🌡', humidity:'💧', co2:'🫧', ph:'⚗️',
  ec:'⚡', light:'☀️', moisture:'🌱', pressure:'🔵', flow:'🌊', default:'📡' }
function sensorIcon(type) {
  if (!type) return ICONS.default
  const k = type.toLowerCase()
  for (const [n, i] of Object.entries(ICONS)) { if (k.includes(n)) return i }
  return ICONS.default
}
const STATUS_BADGE = {
  ok:      'bg-green-900/50 text-green-400',
  unknown: 'bg-zinc-800 text-zinc-400',
  danger:  'bg-red-900/50 text-red-400',
  warn:    'bg-yellow-900/50 text-yellow-400',
}
function statusBadge(s) { return STATUS_BADGE[s] ?? STATUS_BADGE.unknown }
</script>
