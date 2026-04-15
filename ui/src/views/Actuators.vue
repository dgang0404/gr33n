<template>
  <div class="p-6">
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-xl font-semibold text-white">Controls (Actuators)</h1>
      <span class="text-xs text-zinc-500">{{ store.actuators.length }} actuators</span>
    </div>

    <div v-if="store.loading" class="text-zinc-400 text-sm">Loading controls…</div>
    <div v-else-if="!store.actuators.length" class="text-zinc-500 text-sm">
      No actuators found. This usually means no control hardware is registered yet.
      You can still run in simulation mode until a Pi/device is connected.
    </div>

    <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
      <div
        v-for="actuator in store.actuators"
        :key="actuator.id"
        class="bg-zinc-900 border rounded-xl p-4 flex flex-col gap-3 transition-colors"
        :class="actuator.current_state_text === 'online' ? 'border-green-800/70' : 'border-zinc-800'"
      >
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-2 min-w-0">
            <span class="text-xl shrink-0">{{ deviceIcon(actuator.actuator_type) }}</span>
            <div class="min-w-0">
              <p class="text-white text-sm font-medium truncate">{{ actuator.name }}</p>
              <p class="text-zinc-500 text-xs capitalize">{{ actuator.actuator_type }}</p>
            </div>
          </div>
          <!-- Toggle -->
          <button
            @click="toggle(actuator)"
            :disabled="toggling[actuator.id]"
            class="relative shrink-0 w-11 h-6 rounded-full transition-colors focus:outline-none focus:ring-2 focus:ring-green-600 disabled:opacity-40"
            :class="actuator.current_state_text === 'online' ? 'bg-green-600' : 'bg-zinc-700'"
            :title="actuator.current_state_text === 'online' ? 'Turn off' : 'Turn on'"
          >
            <span
              class="absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full shadow transition-transform duration-200"
              :class="actuator.current_state_text === 'online' ? 'translate-x-5' : 'translate-x-0'"
            />
          </button>
        </div>

        <div class="flex items-center justify-between text-xs">
          <span class="text-zinc-400 truncate">{{ zoneName(actuator.zone_id) }}</span>
          <span :class="statusBadge(actuator.current_state_text || 'offline')"
            class="shrink-0 ml-2 px-2 py-0.5 rounded-full font-medium">
            {{ (actuator.current_state_text || 'offline').replaceAll('_', ' ') }}
          </span>
        </div>

        <p v-if="actuator.last_known_state_time" class="text-zinc-600 text-xs">
          Last state update {{ timeAgo(actuator.last_known_state_time) }}
        </p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useFarmStore } from '../stores/farm'

const store = useFarmStore()
const toggling = ref({})
onMounted(() => { if (!store.actuators.length) store.loadAll() })

async function toggle(actuator) {
  toggling.value[actuator.id] = true
  try { await store.toggleActuator(actuator.id, actuator.current_state_text || 'offline') }
  finally { toggling.value[actuator.id] = false }
}
function zoneName(id) {
  if (!id) return 'Unassigned'
  return store.zones.find(z => z.id === id)?.name ?? `Zone ${id}`
}
const DEVICE_ICONS = { pump:'🔧', fan:'🌀', light:'💡', valve:'🚰',
  heater:'🔥', cooler:'❄️', humidifier:'💨', co2:'🫧',
  relay:'⚡', controller:'🖥', pi:'🍓', sensor:'📡', default:'⚙️' }
function deviceIcon(type) {
  if (!type) return DEVICE_ICONS.default
  const k = type.toLowerCase()
  for (const [n, i] of Object.entries(DEVICE_ICONS)) { if (k.includes(n)) return i }
  return DEVICE_ICONS.default
}
const STATUS_BADGE = {
  online:'bg-green-900/60 text-green-400',
  offline:'bg-zinc-800 text-zinc-400',
  on:'bg-green-900/60 text-green-400',
  off:'bg-zinc-800 text-zinc-400',
}
function statusBadge(s) { return STATUS_BADGE[s] ?? 'bg-zinc-800 text-zinc-400' }
function timeAgo(ts) {
  const mins = Math.floor((Date.now() - new Date(ts)) / 60000)
  if (mins < 1) return 'just now'
  if (mins < 60) return `${mins}m ago`
  const hrs = Math.floor(mins / 60)
  return hrs < 24 ? `${hrs}h ago` : `${Math.floor(hrs / 24)}d ago`
}
</script>
