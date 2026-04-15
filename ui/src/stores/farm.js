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
    alerts: [],
    unreadAlertCount: 0,
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
    async loadAll(farmId) {
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

    async loadTasks(farmId) {
      const r = await api.get(`/farms/${farmId}/tasks`)
      this.tasks = Array.isArray(r.data) ? r.data : []
      return this.tasks
    },

    async createTask(farmId, data) {
      const r = await api.post(`/farms/${farmId}/tasks`, data)
      return r.data
    },

    async loadSchedules(farmId) {
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

    async loadAutomationRuns(farmId) {
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

    async loadSensorReadings(sensorId, { since, until, limit } = {}) {
      const params = new URLSearchParams()
      if (since) params.set('since', since)
      if (until) params.set('until', until)
      if (limit != null) params.set('limit', String(limit))
      const qs = params.toString()
      const r = await api.get(`/sensors/${sensorId}/readings${qs ? '?' + qs : ''}`)
      return Array.isArray(r.data) ? r.data : []
    },

    async loadSensorStats(sensorId, { since, until } = {}) {
      const params = new URLSearchParams()
      if (since) params.set('since', since)
      if (until) params.set('until', until)
      const qs = params.toString()
      const r = await api.get(`/sensors/${sensorId}/readings/stats${qs ? '?' + qs : ''}`)
      return r.data
    },

    async loadActuatorEvents(actuatorId, { since, limit } = {}) {
      const params = new URLSearchParams()
      if (since) params.set('since', since)
      if (limit) params.set('limit', String(limit))
      const qs = params.toString()
      const r = await api.get(`/actuators/${actuatorId}/events${qs ? '?' + qs : ''}`)
      return Array.isArray(r.data) ? r.data : []
    },

    async loadActuatorEventsBySchedule(scheduleId, { since, limit } = {}) {
      const params = new URLSearchParams()
      if (since) params.set('since', since)
      if (limit) params.set('limit', String(limit))
      const qs = params.toString()
      const r = await api.get(`/schedules/${scheduleId}/actuator-events${qs ? '?' + qs : ''}`)
      return Array.isArray(r.data) ? r.data : []
    },

    async loadFertigationPrograms(farmId) {
      const r = await api.get(`/farms/${farmId}/fertigation/programs`)
      return Array.isArray(r.data) ? r.data : []
    },

    async loadFertigationEvents(farmId, { cropCycleId } = {}) {
      const params = {}
      if (cropCycleId != null && cropCycleId !== '') params.crop_cycle_id = cropCycleId
      const r = await api.get(`/farms/${farmId}/fertigation/events`, { params })
      return Array.isArray(r.data) ? r.data : []
    },

    async loadReservoirs(farmId) {
      const r = await api.get(`/farms/${farmId}/fertigation/reservoirs`)
      return Array.isArray(r.data) ? r.data : []
    },

    async loadEcTargets(farmId) {
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

    async updateReservoir(id, data) {
      const r = await api.patch(`/fertigation/reservoirs/${id}`, data)
      return r.data
    },

    async deleteReservoir(id) {
      await api.delete(`/fertigation/reservoirs/${id}`)
    },

    async updateProgram(id, data) {
      const r = await api.patch(`/fertigation/programs/${id}`, data)
      return r.data
    },

    async deleteProgram(id) {
      await api.delete(`/fertigation/programs/${id}`)
    },

    // Zone CRUD
    async createZone(farmId, data) {
      const r = await api.post(`/farms/${farmId}/zones`, data)
      this.zones.push(r.data)
      return r.data
    },

    async updateZone(id, data) {
      const r = await api.put(`/zones/${id}`, data)
      const idx = this.zones.findIndex(z => z.id === id)
      if (idx >= 0) this.zones[idx] = r.data
      return r.data
    },

    async deleteZone(id) {
      await api.delete(`/zones/${id}`)
      this.zones = this.zones.filter(z => z.id !== id)
    },

    // Natural farming CRUD
    async loadNfInputs(farmId) {
      const r = await api.get(`/farms/${farmId}/naturalfarming/inputs`)
      return Array.isArray(r.data) ? r.data : []
    },

    async loadNfBatches(farmId) {
      const r = await api.get(`/farms/${farmId}/naturalfarming/batches`)
      return Array.isArray(r.data) ? r.data : []
    },

    async createNfInput(farmId, data) {
      const r = await api.post(`/farms/${farmId}/naturalfarming/inputs`, data)
      return r.data
    },

    async updateNfInput(id, data) {
      const r = await api.put(`/naturalfarming/inputs/${id}`, data)
      return r.data
    },

    async deleteNfInput(id) {
      await api.delete(`/naturalfarming/inputs/${id}`)
    },

    async createNfBatch(farmId, data) {
      const r = await api.post(`/farms/${farmId}/naturalfarming/batches`, data)
      return r.data
    },

    async updateNfBatch(id, data) {
      const r = await api.put(`/naturalfarming/batches/${id}`, data)
      return r.data
    },

    async deleteNfBatch(id) {
      await api.delete(`/naturalfarming/batches/${id}`)
    },

    async loadCropCycles(farmId) {
      const r = await api.get(`/farms/${farmId}/crop-cycles`)
      return Array.isArray(r.data) ? r.data : []
    },

    async createCropCycle(farmId, data) {
      const r = await api.post(`/farms/${farmId}/crop-cycles`, data)
      return r.data
    },

    async updateCropCycle(id, data) {
      const r = await api.put(`/crop-cycles/${id}`, data)
      return r.data
    },

    async updateCropCycleStage(id, stage) {
      const r = await api.patch(`/crop-cycles/${id}/stage`, { current_stage: stage })
      return r.data
    },

    async deleteCropCycle(id) {
      await api.delete(`/crop-cycles/${id}`)
    },

    async loadRecipes(farmId) {
      const r = await api.get(`/farms/${farmId}/naturalfarming/recipes`)
      return Array.isArray(r.data) ? r.data : []
    },

    async createRecipe(farmId, data) {
      const r = await api.post(`/farms/${farmId}/naturalfarming/recipes`, data)
      return r.data
    },

    async updateRecipe(id, data) {
      const r = await api.put(`/naturalfarming/recipes/${id}`, data)
      return r.data
    },

    async deleteRecipe(id) {
      await api.delete(`/naturalfarming/recipes/${id}`)
    },

    async loadRecipeComponents(recipeId) {
      const r = await api.get(`/naturalfarming/recipes/${recipeId}/components`)
      return Array.isArray(r.data) ? r.data : []
    },

    async addRecipeComponent(recipeId, data) {
      await api.post(`/naturalfarming/recipes/${recipeId}/components`, data)
    },

    async removeRecipeComponent(recipeId, inputDefinitionId) {
      await api.delete(`/naturalfarming/recipes/${recipeId}/components/${inputDefinitionId}`)
    },

    async loadCostSummary(farmId) {
      const r = await api.get(`/farms/${farmId}/costs/summary`)
      return r.data
    },

    async loadCosts(farmId, { limit = 50, offset = 0 } = {}) {
      const r = await api.get(`/farms/${farmId}/costs?limit=${limit}&offset=${offset}`)
      return Array.isArray(r.data) ? r.data : []
    },

    async createCost(farmId, data) {
      const r = await api.post(`/farms/${farmId}/costs`, data)
      return r.data
    },

    async updateCost(id, data) {
      const r = await api.put(`/costs/${id}`, data)
      return r.data
    },

    async deleteCost(id) {
      await api.delete(`/costs/${id}`)
    },

    // Alerts
    async loadAlerts(farmId, { limit = 50, offset = 0 } = {}) {
      const r = await api.get(`/farms/${farmId}/alerts?limit=${limit}&offset=${offset}`)
      this.alerts = Array.isArray(r.data) ? r.data : []
      return this.alerts
    },

    async countUnreadAlerts(farmId) {
      const r = await api.get(`/farms/${farmId}/alerts/unread-count`)
      this.unreadAlertCount = r.data?.unread_count ?? 0
      return this.unreadAlertCount
    },

    async markAlertRead(id) {
      const r = await api.patch(`/alerts/${id}/read`)
      const idx = this.alerts.findIndex(a => a.id === id)
      if (idx >= 0) this.alerts[idx] = r.data
      return r.data
    },

    async markAlertAcknowledged(id) {
      const r = await api.patch(`/alerts/${id}/acknowledge`)
      const idx = this.alerts.findIndex(a => a.id === id)
      if (idx >= 0) this.alerts[idx] = r.data
      return r.data
    },

    // Farm members
    async loadFarmMembers(farmId) {
      const r = await api.get(`/farms/${farmId}/members`)
      return Array.isArray(r.data) ? r.data : []
    },

    async addFarmMember(farmId, data) {
      const r = await api.post(`/farms/${farmId}/members`, data)
      return r.data
    },

    async updateFarmMemberRole(farmId, userId, roleInFarm) {
      const r = await api.patch(`/farms/${farmId}/members/${userId}/role`, { role_in_farm: roleInFarm })
      return r.data
    },

    async removeFarmMember(farmId, userId) {
      await api.delete(`/farms/${farmId}/members/${userId}`)
    },

    // Profile
    async getProfile() {
      const r = await api.get('/profile')
      return r.data
    },

    async updateProfile(data) {
      const r = await api.put('/profile', data)
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
