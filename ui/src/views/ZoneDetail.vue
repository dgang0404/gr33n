<template>
  <div class="p-6 space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <router-link to="/zones" class="text-xs text-zinc-500 hover:text-zinc-300">← Back to zones</router-link>
        <h1 class="text-xl font-semibold text-white mt-1">{{ zone?.name || 'Zone' }}</h1>
        <p class="text-zinc-500 text-sm">{{ zone?.description || 'No description' }}</p>
      </div>
      <span :class="zoneBadge(zone?.zone_type)" class="text-xs font-medium px-2 py-1 rounded-full capitalize">
        {{ zone?.zone_type || 'unknown' }}
      </span>
    </div>

    <div v-if="!zone" class="text-zinc-500 text-sm">Zone not found.</div>

    <template v-else>
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
          <p class="text-zinc-400 text-xs mb-2">Sensors</p>
          <p class="text-white text-2xl font-semibold">{{ sensors.length }}</p>
        </div>
        <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
          <p class="text-zinc-400 text-xs mb-2">Devices</p>
          <p class="text-white text-2xl font-semibold">{{ devices.length }}</p>
        </div>
        <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
          <p class="text-zinc-400 text-xs mb-2">Actuators</p>
          <p class="text-white text-2xl font-semibold">{{ actuators.length }}</p>
        </div>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
          <h2 class="text-sm font-semibold text-white mb-3">Sensors</h2>
          <p v-if="!sensors.length" class="text-zinc-500 text-sm">No sensors assigned.</p>
          <ul v-else class="space-y-2">
            <li v-for="s in sensors" :key="s.id" class="text-sm text-zinc-300">
              {{ s.name }} <span class="text-zinc-500">({{ s.sensor_type }})</span>
            </li>
          </ul>
        </div>

        <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
          <h2 class="text-sm font-semibold text-white mb-3">Controls</h2>
          <p v-if="!actuators.length" class="text-zinc-500 text-sm">No actuators assigned.</p>
          <ul v-else class="space-y-2">
            <li v-for="a in actuators" :key="a.id" class="text-sm text-zinc-300">
              {{ a.name }} <span class="text-zinc-500">({{ a.actuator_type }})</span>
            </li>
          </ul>
        </div>
      </div>

      <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
        <h2 class="text-sm font-semibold text-white mb-3">Fertigation Summary</h2>
        <p class="text-zinc-500 text-sm mb-3">Active programs and latest event for this zone.</p>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
          <div class="bg-zinc-950 border border-zinc-800 rounded-lg p-3">
            <p class="text-zinc-400 text-xs mb-1">Active Program</p>
            <p class="text-zinc-200 text-sm">{{ activeProgram?.name || 'None' }}</p>
          </div>
          <div class="bg-zinc-950 border border-zinc-800 rounded-lg p-3">
            <p class="text-zinc-400 text-xs mb-1">Latest Event</p>
            <p class="text-zinc-200 text-sm">
              {{ latestEvent ? `${latestEvent.applied_at} · ${latestEvent.volume_applied_liters || '0'}L` : 'No events' }}
            </p>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { useFarmStore } from '../stores/farm'
import api from '../api'

const route = useRoute()
const store = useFarmStore()

const programs = ref([])
const events = ref([])

const zoneId = computed(() => Number(route.params.id))
const zone = computed(() => store.zones.find(z => z.id === zoneId.value))
const sensors = computed(() => store.sensorsByZone(zoneId.value))
const devices = computed(() => store.devicesByZone(zoneId.value))
const actuators = computed(() => store.actuatorsByZone(zoneId.value))
const activeProgram = computed(() => programs.value.find(p => p.target_zone_id === zoneId.value && p.is_active))
const latestEvent = computed(() => events.value.find(e => e.zone_id === zoneId.value))

onMounted(async () => {
  if (!store.zones.length) await store.loadAll(1)
  try {
    const [p, e] = await Promise.all([
      api.get('/farms/1/fertigation/programs'),
      api.get('/farms/1/fertigation/events'),
    ])
    programs.value = Array.isArray(p.data) ? p.data : []
    events.value = Array.isArray(e.data) ? e.data : []
  } catch {
    programs.value = []
    events.value = []
  }
})

const BADGE = {
  indoor: 'bg-indigo-900/60 text-indigo-300',
  outdoor: 'bg-emerald-900/60 text-emerald-300',
  greenhouse: 'bg-green-900/60 text-green-300',
}
function zoneBadge(type) {
  if (!type) return 'bg-zinc-800 text-zinc-400'
  const k = String(type).toLowerCase()
  for (const [name, cls] of Object.entries(BADGE)) {
    if (k.includes(name)) return cls
  }
  return 'bg-zinc-800 text-zinc-400'
}
</script>
