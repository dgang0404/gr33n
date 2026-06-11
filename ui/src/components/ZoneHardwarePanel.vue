<template>
  <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4" data-test="zone-hardware-panel">
    <div class="flex items-start justify-between gap-3 mb-3 flex-wrap">
      <div>
        <h2 class="text-sm font-semibold text-white">Sensors &amp; controls</h2>
        <p class="text-zinc-500 text-xs mt-1">
          Every sensor and actuator in this zone — moisture, climate, lights, pumps, and GPIO wiring.
        </p>
      </div>
      <router-link
        v-nav-hint="'/zones'"
        :to="{ path: '/zones', query: { tab: 'fleet', fleet: 'sensors' } }"
        class="text-xs text-zinc-500 hover:text-green-400 shrink-0"
      >
        Farm fleet →
      </router-link>
    </div>

    <EmptyStateHint
      v-if="!sensors.length && !actuators.length"
      reason="no_telemetry"
      message="No sensors or actuators in this zone yet. Add hardware from Hardware & devices or the Pi setup guide."
      compact
    />

    <div v-if="sensors.length" class="mb-4">
      <h3 class="text-xs font-semibold text-zinc-400 uppercase tracking-wide mb-2">
        Sensors ({{ sensors.length }})
      </h3>
      <div class="space-y-2">
        <div
          v-for="s in sensors"
          :key="s.id"
          class="bg-zinc-950 border border-zinc-800 rounded-lg p-3"
          :data-test="`zone-hardware-sensor-${s.id}`"
        >
          <div class="flex items-start justify-between gap-2 flex-wrap">
            <div class="min-w-0 flex-1">
              <div class="flex items-center gap-2 flex-wrap">
                <p class="text-sm text-white font-medium truncate">{{ s.name || s.sensor_type }}</p>
                <span class="text-[10px] px-1.5 py-0.5 rounded bg-zinc-800 text-zinc-400 capitalize">
                  {{ s.sensor_type }}
                </span>
                <span class="text-[10px] px-1.5 py-0.5 rounded bg-zinc-800/80 text-zinc-500">
                  {{ needLabel(sensorPlantNeed(s.sensor_type)) }}
                </span>
              </div>
              <HardwareWiringBadge :entity="s" show-empty class="mt-1.5" :hint-path="zoneHintPath" />
            </div>
          </div>
          <SensorTile :sensor="s" :reading="store.readings[s.id]" class="mt-2" />
          <button
            type="button"
            class="mt-2 text-[10px] text-zinc-500 hover:text-zinc-300"
            :data-test="`zone-hardware-sensor-wiring-${s.id}`"
            @click="toggleSensorWiring(s.id)"
          >
            {{ sensorWiringOpen[s.id] ? '▾ Hide wiring' : '▸ Edit wiring' }}
          </button>
          <HardwareWiringPanel
            v-if="sensorWiringOpen[s.id]"
            :sensor-id="s.id"
            :sensor="s"
            :devices="store.devices"
            :sensors="sensors"
            :actuators="actuators"
            class="mt-2 border-0 bg-transparent p-0"
            @updated="$emit('hardware-updated')"
          />
        </div>
      </div>
    </div>

    <div v-if="actuators.length">
      <h3 class="text-xs font-semibold text-zinc-400 uppercase tracking-wide mb-2">
        Controls ({{ actuators.length }})
      </h3>
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
        <div
          v-for="a in actuators"
          :key="a.id"
          class="bg-zinc-950 border rounded-lg p-3"
          :class="a.current_state_text === 'online' ? 'border-green-800/70' : 'border-zinc-800'"
          :data-test="`zone-hardware-actuator-${a.id}`"
        >
          <div class="flex items-center justify-between gap-2">
            <div class="min-w-0">
              <p class="text-white text-sm font-medium truncate">{{ a.name }}</p>
              <div class="flex items-center gap-2 flex-wrap mt-0.5">
                <p class="text-zinc-500 text-xs capitalize">{{ a.actuator_type }}</p>
                <span class="text-[10px] px-1.5 py-0.5 rounded bg-zinc-800/80 text-zinc-500">
                  {{ needLabel(actuatorPlantNeed(a.actuator_type)) }}
                </span>
              </div>
              <HardwareWiringBadge :entity="a" show-empty class="mt-1" :hint-path="zoneHintPath" />
            </div>
            <button
              type="button"
              class="relative shrink-0 w-11 h-6 rounded-full transition-colors disabled:opacity-40"
              :class="a.current_state_text === 'online' ? 'bg-green-600' : 'bg-zinc-700'"
              :disabled="toggling[a.id]"
              @click="$emit('toggle-actuator', a)"
            >
              <span
                class="absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full shadow transition-transform duration-200"
                :class="a.current_state_text === 'online' ? 'translate-x-5' : 'translate-x-0'"
              />
            </button>
          </div>
          <ActuatorPulseControl :actuator="a" />
          <button
            type="button"
            class="mt-2 text-[10px] text-zinc-500 hover:text-zinc-300"
            :data-test="`zone-hardware-actuator-wiring-${a.id}`"
            @click="toggleActuatorWiring(a.id)"
          >
            {{ actuatorWiringOpen[a.id] ? '▾ Hide wiring' : '▸ Edit wiring' }}
          </button>
          <ActuatorWiringPanel
            v-if="actuatorWiringOpen[a.id]"
            :actuator="a"
            :devices="store.devices"
            class="mt-2"
            @updated="$emit('hardware-updated')"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, reactive } from 'vue'
import {
  NEED_META,
  PLANT_NEEDS,
  sensorPlantNeed,
  actuatorPlantNeed,
} from '../lib/plantNeeds.js'
import { useFarmStore } from '../stores/farm.js'
import SensorTile from './SensorTile.vue'
import HardwareWiringBadge from './HardwareWiringBadge.vue'
import HardwareWiringPanel from './HardwareWiringPanel.vue'
import ActuatorWiringPanel from './ActuatorWiringPanel.vue'
import ActuatorPulseControl from './ActuatorPulseControl.vue'
import EmptyStateHint from './EmptyStateHint.vue'

const props = defineProps({
  zoneId: { type: Number, required: true },
  sensors: { type: Array, default: () => [] },
  actuators: { type: Array, default: () => [] },
  toggling: { type: Object, default: () => ({}) },
})

defineEmits(['toggle-actuator', 'hardware-updated'])

const store = useFarmStore()
const sensorWiringOpen = reactive({})
const actuatorWiringOpen = reactive({})

const zoneHintPath = computed(() => `/zones/${props.zoneId}`)

function needLabel(need) {
  const meta = NEED_META[need] || NEED_META[PLANT_NEEDS.air]
  return `${meta.icon} ${meta.shortLabel}`
}

function toggleSensorWiring(id) {
  sensorWiringOpen[id] = !sensorWiringOpen[id]
}

function toggleActuatorWiring(id) {
  actuatorWiringOpen[id] = !actuatorWiringOpen[id]
}
</script>
