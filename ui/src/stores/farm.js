import { defineStore } from 'pinia'
import api from '../api'

export const useFarmStore = defineStore('farm', {
  state: () => ({
    farm: null,
    zones: [],
    sensors: [],
    devices: [],
    actuators: [],
    schedules: [],
    automationRuns: [],
    tasks: [],
    readings: {},
    loading: false,
    error: null,
  }),

  getters: {
    sensorStatus: (state) => (id) => {
      const r = state.readings[id]
      if (!r) return 'unknown'
      if (!r.is_valid) return 'danger'
      return r.status ?? 'ok'
    },
    activeDevices:  (state) => state.devices.filter(d => d.status === 'online'),
    devicesByZone:  (state) => (zoneId) => state.devices.filter(d => d.zone_id === zoneId),
    sensorsByZone:  (state) => (zoneId) => state.sensors.filter(s => s.zone_id === zoneId),
    actuatorsByZone: (state) => (zoneId) => state.actuators.filter(a => a.zone_id === zoneId),
  },

  actions: {
    async loadAll(farmId = 1) {
      this.loading = true
      try {
        const [farm, zones, sensors, devices, actuators] = await Promise.all([
          api.get(`/farms/${farmId}`),
          api.get(`/farms/${farmId}/zones`),
          api.get(`/farms/${farmId}/sensors`),
          api.get(`/farms/${farmId}/devices`),
          api.get(`/farms/${farmId}/actuators`),
        ])
        this.farm    = farm.data
        this.zones   = Array.isArray(zones.data)   ? zones.data   : []
        this.sensors = Array.isArray(sensors.data) ? sensors.data : []
        this.devices = Array.isArray(devices.data) ? devices.data : []
        this.actuators = Array.isArray(actuators.data) ? actuators.data : []
      } catch (e) {
        this.error = e.message
      } finally {
        this.loading = false
      }
    },

    async loadTasks(farmId = 1) {
      const r = await api.get(`/farms/${farmId}/tasks`)
      this.tasks = Array.isArray(r.data) ? r.data : []
      return this.tasks
    },

    async loadSchedules(farmId = 1) {
      const r = await api.get(`/farms/${farmId}/schedules`)
      this.schedules = Array.isArray(r.data) ? r.data : []
      return this.schedules
    },

    async updateScheduleActive(scheduleId, isActive) {
      const r = await api.patch(`/schedules/${scheduleId}/active`, { is_active: isActive })
      const next = r.data
      const idx = this.schedules.findIndex(s => s.id === scheduleId)
      if (idx >= 0) this.schedules[idx] = next
      return next
    },

    async loadAutomationRuns(farmId = 1) {
      const r = await api.get(`/farms/${farmId}/automation/runs`)
      this.automationRuns = Array.isArray(r.data) ? r.data : []
      return this.automationRuns
    },

    async updateTaskStatus(taskId, status) {
      await api.patch(`/tasks/${taskId}/status`, { status })
      const t = this.tasks.find(t => t.id === taskId)
      if (t) t.status = status
    },

    async refreshReadings() {
      for (const s of this.sensors) {
        try {
          const r = await api.get(`/sensors/${s.id}/readings/latest`)
          this.readings[s.id] = r.data
        } catch { /* no readings yet */ }
      }
    },

    async loadActuatorEvents(actuatorId, { since, limit } = {}) {
      const params = new URLSearchParams()
      if (since) params.set('since', since)
      if (limit) params.set('limit', String(limit))
      const qs = params.toString()
      const r = await api.get(`/actuators/${actuatorId}/events${qs ? '?' + qs : ''}`)
      return Array.isArray(r.data) ? r.data : []
    },

    async loadFertigationPrograms(farmId = 1) {
      const r = await api.get(`/farms/${farmId}/fertigation/programs`)
      return Array.isArray(r.data) ? r.data : []
    },

    async loadFertigationEvents(farmId = 1) {
      const r = await api.get(`/farms/${farmId}/fertigation/events`)
      return Array.isArray(r.data) ? r.data : []
    },

    async loadReservoirs(farmId = 1) {
      const r = await api.get(`/farms/${farmId}/fertigation/reservoirs`)
      return Array.isArray(r.data) ? r.data : []
    },

    async loadEcTargets(farmId = 1) {
      const r = await api.get(`/farms/${farmId}/fertigation/ec-targets`)
      return Array.isArray(r.data) ? r.data : []
    },

    async createReservoir(farmId, data) {
      const r = await api.post(`/farms/${farmId}/fertigation/reservoirs`, data)
      return r.data
    },

    async createEcTarget(farmId, data) {
      const r = await api.post(`/farms/${farmId}/fertigation/ec-targets`, data)
      return r.data
    },

    async createProgram(farmId, data) {
      const r = await api.post(`/farms/${farmId}/fertigation/programs`, data)
      return r.data
    },

    async createFertigationEvent(farmId, data) {
      const r = await api.post(`/farms/${farmId}/fertigation/events`, data)
      return r.data
    },

    async toggleDevice(deviceId, currentStatus) {
      const next = currentStatus === 'online' ? 'offline' : 'online'
      await api.patch(`/devices/${deviceId}/status`, { status: next })
      const d = this.devices.find(d => d.id === deviceId)
      if (d) d.status = next
    },

    async toggleActuator(actuatorId, currentStateText) {
      const nextText = currentStateText === 'online' ? 'offline' : 'online'
      const r = await api.patch(`/actuators/${actuatorId}/state`, {
        state_text: nextText,
        state_numeric: nextText === 'online' ? 1 : 0,
      })
      const idx = this.actuators.findIndex(a => a.id === actuatorId)
      if (idx >= 0) this.actuators[idx] = r.data
      return r.data
    },
  },
})
