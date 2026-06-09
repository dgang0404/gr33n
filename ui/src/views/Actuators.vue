<template>
  <div :class="embedded ? '' : 'p-6'">
    <div v-if="!embedded" class="flex items-center justify-between mb-2">
      <h1 class="text-xl font-semibold text-white">Controls</h1>
      <span class="text-xs text-zinc-500">{{ store.actuators.length }} actuators</span>
    </div>
    <p v-if="!embedded" class="text-zinc-500 text-sm mb-6 max-w-2xl">
      Manual switches for pumps, lights, and fans. For zone-scoped edits, open a
      <router-link v-nav-hint="'/zones'" to="/zones" class="text-green-600 hover:text-green-400">zone</router-link>
      and use the Water / Light / Climate tabs.
    </p>

    <div v-if="store.loading" class="text-zinc-400 text-sm">Loading controls…</div>
    <div v-else-if="!store.actuators.length" class="text-zinc-500 text-sm">
      No actuators found. Register hardware in Settings or apply a starter pack.
    </div>

    <template v-else>
      <div
        v-for="group in displayGroups"
        :key="group.zoneId ?? 'unassigned'"
        :class="groupByZone ? 'mb-6' : ''"
        data-test="fleet-zone-group"
      >
        <h2 v-if="groupByZone" class="text-sm font-semibold text-zinc-300 mb-2">{{ group.zoneName }}</h2>
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
          <div
            v-for="actuator in group.items"
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
                  <p class="text-zinc-600 text-[10px]">{{ needLabel(actuator.actuator_type) }}</p>
                  <div class="mt-1 flex flex-wrap gap-1 items-center">
                    <HardwareWiringBadge :entity="actuator" show-empty />
                    <span
                      v-if="pinConflict(actuator)"
                      class="text-[10px] px-1.5 py-0.5 rounded bg-amber-900/40 text-amber-300 border border-amber-800/40"
                      data-test="fleet-pin-conflict"
                    >Pin conflict</span>
                  </div>
                </div>
              </div>
              <button
                @click="toggle(actuator)"
                :disabled="toggling[actuator.id]"
                class="relative shrink-0 w-11 h-6 rounded-full transition-colors focus:outline-none focus:ring-2 focus:ring-green-600 disabled:opacity-40"
                :class="actuator.current_state_text === 'online' ? 'bg-green-600' : 'bg-zinc-700'"
              >
                <span
                  class="absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full shadow transition-transform duration-200"
                  :class="actuator.current_state_text === 'online' ? 'translate-x-5' : 'translate-x-0'"
                />
              </button>
            </div>

            <div v-if="!groupByZone" class="flex items-center justify-between text-xs">
              <router-link
                v-if="actuator.zone_id"
                v-nav-hint="`/zones/${actuator.zone_id}`"
                :to="`/zones/${actuator.zone_id}`"
                class="text-green-600 hover:text-green-400 truncate"
                data-test="actuator-zone-link"
              >
                {{ zoneName(actuator.zone_id) }}
              </router-link>
              <span v-else class="text-zinc-400 truncate">Unassigned</span>
              <span
                v-if="(actuator.current_state_text || 'offline') === 'online'"
                class="shrink-0 ml-2 px-2 py-0.5 rounded-full font-medium bg-green-900/60 text-green-400"
              >
                online
              </span>
              <span
                v-else
                v-nav-hint="'/pi-setup'"
                class="shrink-0 ml-2 px-2 py-0.5 rounded-full font-medium bg-zinc-800 text-zinc-400 cursor-default"
                title="Device offline — see Pi + HAT setup in sidebar"
                data-test="actuator-offline-hint"
              >
                {{ (actuator.current_state_text || 'offline').replaceAll('_', ' ') }}
              </span>
            </div>

            <ActuatorPulseControl :actuator="actuator" />

            <button
              type="button"
              class="text-[10px] text-left text-zinc-500 hover:text-zinc-300 flex items-center gap-1 -mb-1"
              data-test="actuator-wiring-toggle"
              @click="toggleWiring(actuator.id)"
            >
              <span>{{ wiringOpen[actuator.id] ? '▾' : '▸' }}</span>
              <span>{{ wiringOpen[actuator.id] ? 'Hide wiring' : 'Edit wiring' }}</span>
            </button>

            <ActuatorWiringPanel
              v-if="wiringOpen[actuator.id]"
              :actuator="actuator"
              :devices="store.devices"
              data-test="actuator-wiring-panel"
              @updated="onActuatorUpdated"
            />
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup>
import { computed, ref, onMounted } from 'vue'
import { useFarmStore } from '../stores/farm'
import { useFarmContextStore } from '../stores/farmContext'
import { actuatorPlantNeed, NEED_META } from '../lib/plantNeeds.js'
import ActuatorPulseControl from '../components/ActuatorPulseControl.vue'
import HardwareWiringBadge from '../components/HardwareWiringBadge.vue'
import ActuatorWiringPanel from '../components/ActuatorWiringPanel.vue'
import { actuatorPinConflict, groupEntitiesByZone } from '../lib/fleetGrouping.js'

const props = defineProps({
  embedded: { type: Boolean, default: false },
  groupByZone: { type: Boolean, default: false },
})

const store = useFarmStore()
const farmContext = useFarmContextStore()
const toggling = ref({})
const wiringOpen = ref({})

const displayGroups = computed(() =>
  props.groupByZone
    ? groupEntitiesByZone(store.actuators, store.zones)
    : [{ zoneId: null, zoneName: '', items: store.actuators }],
)

onMounted(() => { if (!store.actuators.length && farmContext.farmId) store.loadAll(farmContext.farmId) })

function pinConflict(actuator) {
  return actuatorPinConflict(actuator, store.sensors, store.actuators)
}

function toggleWiring(id) {
  wiringOpen.value[id] = !wiringOpen.value[id]
}

function onActuatorUpdated(updated) {
  const idx = store.actuators.findIndex(a => a.id === updated.id)
  if (idx >= 0) store.actuators[idx] = updated
}

async function toggle(actuator) {
  toggling.value[actuator.id] = true
  try { await store.toggleActuator(actuator.id, actuator.current_state_text || 'offline') }
  finally { toggling.value[actuator.id] = false }
}
function zoneName(id) {
  if (!id) return 'Unassigned'
  return store.zones.find(z => z.id === id)?.name ?? `Zone ${id}`
}
function needLabel(actuatorType) {
  const need = actuatorPlantNeed(actuatorType)
  return NEED_META[need]?.label ?? ''
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
</script>
