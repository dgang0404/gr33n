<template>
  <div class="p-6">
    <h1 class="text-xl font-semibold text-white mb-6">Zones</h1>

    <div v-if="store.loading" class="text-zinc-400 text-sm">Loading zones…</div>
    <div v-else-if="!store.zones.length" class="text-zinc-500 text-sm">No zones found.</div>

    <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
      <router-link
        v-for="zone in store.zones"
        :key="zone.id"
        :to="`/zones/${zone.id}`"
        class="bg-zinc-900 border border-zinc-800 rounded-xl p-5 hover:border-green-700 transition-colors group block"
      >
        <div class="flex items-start justify-between mb-3">
          <span class="text-white font-medium group-hover:text-green-400 transition-colors">
            {{ zone.name }}
          </span>
          <span :class="zoneBadge(zone.zone_type)" class="text-xs font-medium px-2 py-0.5 rounded-full capitalize">
            {{ zone.zone_type || 'unknown' }}
          </span>
        </div>

        <p v-if="zone.description" class="text-zinc-500 text-sm mb-3 line-clamp-2">
          {{ zone.description }}
        </p>

        <div class="flex items-center gap-4 text-xs text-zinc-400">
          <span>🌡 {{ store.sensorsByZone(zone.id).length }} sensors</span>
          <span>⚙️ {{ store.devicesByZone(zone.id).length }} devices</span>
          <span v-if="zone.area_sqm">📐 {{ zone.area_sqm }} m²</span>
        </div>
      </router-link>
    </div>
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { useFarmStore } from '../stores/farm'

const store = useFarmStore()
onMounted(() => { if (!store.zones.length) store.loadAll() })

const BADGE = {
  indoor:     'bg-indigo-900/60 text-indigo-300',
  outdoor:    'bg-emerald-900/60 text-emerald-300',
  greenhouse: 'bg-green-900/60 text-green-300',
  nursery:    'bg-yellow-900/60 text-yellow-300',
  seedling:   'bg-lime-900/60 text-lime-300',
  veg:        'bg-teal-900/60 text-teal-300',
  flower:     'bg-pink-900/60 text-pink-300',
  storage:    'bg-zinc-700/60 text-zinc-300',
}
function zoneBadge(type) {
  if (!type) return 'bg-zinc-800 text-zinc-400'
  const k = type.toLowerCase()
  for (const [name, cls] of Object.entries(BADGE)) {
    if (k.includes(name)) return cls
  }
  return 'bg-zinc-800 text-zinc-400'
}
</script>
