<template>
  <div class="space-y-6">

    <!-- Farm header -->
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-xl font-bold text-white">{{ store.farm?.name ?? 'Loading...' }}</h2>
        <p class="text-sm text-gray-500">{{ store.zones.length }} zones · {{ store.sensors.length }} sensors · {{ store.devices.length }} devices</p>
      </div>
      <button @click="farmContext.farmId && store.loadAll(farmContext.farmId)" class="text-xs text-gr33n-400 hover:text-gr33n-300 transition-colors">
        ↻ Refresh
      </button>
    </div>

    <!-- Sensor tiles -->
    <section>
      <h3 class="text-xs font-semibold text-gray-500 uppercase tracking-widest mb-3">Live Sensors</h3>
      <div v-if="store.sensors.length" class="grid grid-cols-2 md:grid-cols-4 xl:grid-cols-7 gap-3">
        <SensorTile v-for="s in store.sensors" :key="s.id"
          :sensor="s" :reading="store.readings[s.id]" />
      </div>
      <div v-else class="text-sm text-gray-600">No sensors found for this farm.</div>
    </section>

    <!-- Zone cards -->
    <section>
      <h3 class="text-xs font-semibold text-gray-500 uppercase tracking-widest mb-3">Zones</h3>
      <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
        <div v-for="zone in store.zones" :key="zone.id" class="card space-y-3">
          <div class="flex items-center justify-between">
            <span class="font-semibold text-white">{{ zone.name }}</span>
            <span class="text-xs text-gray-500">{{ zone.zone_type }}</span>
          </div>
          <!-- Zone sensors -->
          <div class="grid grid-cols-2 gap-2">
            <SensorTile v-for="s in store.sensorsByZone(zone.id)" :key="s.id"
              :sensor="s" :reading="store.readings[s.id]" />
          </div>
          <!-- Zone actuators -->
          <div class="space-y-2 pt-1 border-t border-gray-800">
            <ActuatorCard v-for="d in store.devicesByZone(zone.id)" :key="d.id" :device="d" />
          </div>
          <div v-if="!store.devicesByZone(zone.id).length && !store.sensorsByZone(zone.id).length"
            class="text-xs text-gray-600">No devices assigned to this zone yet.</div>
        </div>
        <div v-if="!store.zones.length" class="text-sm text-gray-600">No zones found.</div>
      </div>
    </section>

    <!-- Quick actuator panel -->
    <section>
      <h3 class="text-xs font-semibold text-gray-500 uppercase tracking-widest mb-3">All Actuators</h3>
      <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-3">
        <ActuatorCard v-for="d in store.devices" :key="d.id" :device="d" />
        <div v-if="!store.devices.length" class="text-sm text-gray-600">No devices found.</div>
      </div>
    </section>

  </div>
</template>

<script setup>
import { useFarmStore } from '../stores/farm'
import { useFarmContextStore } from '../stores/farmContext'
import SensorTile   from '../components/SensorTile.vue'
import ActuatorCard from '../components/ActuatorCard.vue'
const store = useFarmStore()
const farmContext = useFarmContextStore()
</script>
