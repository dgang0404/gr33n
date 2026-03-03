import { defineStore } from 'pinia'
import api from '../api'

export const useFarmStore = defineStore('farm', {
  state: () => ({
    farm:     null,
    zones:    [],
    sensors:  [],
    devices:  [],
    readings: {},   // sensor_id -> latest reading
    loading:  false,
    error:    null,
  }),

  getters: {
    sensorStatus: (state) => (id) => {
      const r = state.readings[id]
      if (!r) return 'unknown'
      if (!r.is_valid) return 'danger'
      return r.status ?? 'ok'
    },
    activeDevices: (state) => state.devices.filter(d => d.status === 'online'),
    devicesByZone: (state) => (zoneId) => state.devices.filter(d => d.zone_id === zoneId),
    sensorsByZone: (state) => (zoneId) => state.sensors.filter(s => s.zone_id === zoneId),
  },

  actions: {
    async loadAll(farmId = 1) {
      this.loading = true
      try {
        const [farm, zones, sensors, devices] = await Promise.all([
          api.get(`/farms/${farmId}`),
          api.get(`/farms/${farmId}/zones`),
          api.get(`/farms/${farmId}/sensors`),
          api.get(`/farms/${farmId}/devices`),
        ])
        this.farm    = farm.data
        this.zones   = Array.isArray(zones.data)   ? zones.data   : []
        this.sensors = Array.isArray(sensors.data) ? sensors.data : []
        this.devices = Array.isArray(devices.data) ? devices.data : []
      } catch (e) {
        this.error = e.message
      } finally {
        this.loading = false
      }
    },

    async refreshReadings() {
      for (const s of this.sensors) {
        try {
          const r = await api.get(`/sensors/${s.id}/readings/latest`)
          this.readings[s.id] = r.data
        } catch { /* sensor may have no readings yet */ }
      }
    },

    async toggleDevice(deviceId, currentStatus) {
      const next = currentStatus === 'online' ? 'offline' : 'online'
      await api.patch(`/devices/${deviceId}/status`, { status: next })
      const d = this.devices.find(d => d.id === deviceId)
      if (d) d.status = next
    },
  },
})
