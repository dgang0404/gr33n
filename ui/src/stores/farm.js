import { defineStore } from 'pinia'
import api from '../api'
import {
  dataUrlToFile,
  fileToDataUrl,
  isRetryableTaskQueueError,
  loadOfflineQueue,
  makeCreateCostItem,
  makeCreateTaskItem,
  makeUpdateTaskStatusItem,
  pendingCount,
  saveOfflineQueue,
} from '../offline/taskQueue'

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
    taskWriteQueue: loadOfflineQueue(),
    taskQueueBusy: false,
    taskSyncStatus: {
      lastAttemptAt: '',
      lastResult: 'idle',
      lastMessage: '',
    },
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
    taskQueuePendingCount: (state) => (farmId) => pendingCount(state.taskWriteQueue, farmId),
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
      this.tasks = this.withTaskQueueOverlay(Array.isArray(r.data) ? r.data : [], farmId)
      return this.tasks
    },

    async createTask(farmId, data) {
      const tryOnline = typeof navigator === 'undefined' || navigator.onLine
      if (tryOnline) {
        try {
          const r = await api.post(`/farms/${farmId}/tasks`, data)
          return r.data
        } catch (err) {
          if (!isRetryableTaskQueueError(err)) throw err
        }
      }
      const item = makeCreateTaskItem(farmId, data)
      this.taskWriteQueue.push(item)
      this.persistTaskWriteQueue()
      const optimistic = this.optimisticTaskFromCreateItem(item)
      this.tasks = [...this.tasks, optimistic]
      return optimistic
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

    async createSchedule(farmId, payload) {
      const r = await api.post(`/farms/${farmId}/schedules`, payload)
      this.schedules = [...this.schedules, r.data]
      return r.data
    },

    async updateSchedule(scheduleId, payload) {
      const r = await api.put(`/schedules/${scheduleId}`, payload)
      const idx = this.schedules.findIndex(s => s.id === scheduleId)
      if (idx >= 0) this.schedules[idx] = r.data
      return r.data
    },

    async deleteSchedule(scheduleId) {
      await api.delete(`/schedules/${scheduleId}`)
      this.schedules = this.schedules.filter(s => s.id !== scheduleId)
    },

    async loadAutomationRuns(farmId) {
      const r = await api.get(`/farms/${farmId}/automation/runs`)
      this.automationRuns = Array.isArray(r.data) ? r.data : []
      return this.automationRuns
    },

    // ── Automation rules (Phase 20) ──────────────────────────────────────
    async loadAutomationRules(farmId) {
      const r = await api.get(`/farms/${farmId}/automation/rules`)
      return Array.isArray(r.data) ? r.data : []
    },

    async createAutomationRule(farmId, payload) {
      const r = await api.post(`/farms/${farmId}/automation/rules`, payload)
      return r.data
    },

    async getAutomationRule(ruleId) {
      const r = await api.get(`/automation/rules/${ruleId}`)
      return r.data
    },

    async updateAutomationRule(ruleId, payload) {
      const r = await api.put(`/automation/rules/${ruleId}`, payload)
      return r.data
    },

    async updateAutomationRuleActive(ruleId, isActive) {
      const r = await api.patch(`/automation/rules/${ruleId}/active`, { is_active: isActive })
      return r.data
    },

    async deleteAutomationRule(ruleId) {
      await api.delete(`/automation/rules/${ruleId}`)
    },

    async loadRuleActions(ruleId) {
      const r = await api.get(`/automation/rules/${ruleId}/actions`)
      return Array.isArray(r.data) ? r.data : []
    },

    async createRuleAction(ruleId, payload) {
      const r = await api.post(`/automation/rules/${ruleId}/actions`, payload)
      return r.data
    },

    async updateRuleAction(actionId, payload) {
      const r = await api.put(`/automation/actions/${actionId}`, payload)
      return r.data
    },

    async deleteRuleAction(actionId) {
      await api.delete(`/automation/actions/${actionId}`)
    },

    async updateTaskStatus(taskId, status) {
      const task = this.tasks.find((t) => String(t.id) === String(taskId))
      const farmId = task?.farm_id
      if (!task || !farmId) {
        await api.patch(`/tasks/${taskId}/status`, { status })
        return
      }
      const tryOnline = typeof navigator === 'undefined' || navigator.onLine
      if (tryOnline && typeof task.id === 'number') {
        try {
          await api.patch(`/tasks/${task.id}/status`, { status })
          task.status = status
          task._offline = null
          return
        } catch (err) {
          if (!isRetryableTaskQueueError(err)) throw err
        }
      }

      const taskRef = typeof task.id === 'number' ? task.id : task.id
      const item = makeUpdateTaskStatusItem(farmId, taskRef, status)
      this.taskWriteQueue.push(item)
      this.persistTaskWriteQueue()
      task.status = status
      task._offline = {
        ...(task._offline || {}),
        queued: true,
        pendingSync: true,
        stale: false,
        queueItemId: item.id,
      }
    },

    async updateTask(taskId, payload) {
      const r = await api.put(`/tasks/${taskId}`, payload)
      const idx = this.tasks.findIndex(t => t.id === taskId)
      if (idx >= 0) this.tasks[idx] = r.data
      return r.data
    },

    async deleteTask(taskId) {
      await api.delete(`/tasks/${taskId}`)
      this.tasks = this.tasks.filter(t => t.id !== taskId)
    },

    async flushTaskWriteQueue({ farmId, force = false } = {}) {
      if (this.taskQueueBusy) return
      if (!force && typeof navigator !== 'undefined' && !navigator.onLine) return
      this.taskQueueBusy = true
      this.taskSyncStatus.lastAttemptAt = new Date().toISOString()
      this.taskSyncStatus.lastResult = 'running'
      this.taskSyncStatus.lastMessage = ''
      const localToServer = new Map()
      let hadFailures = false
      try {
        for (const item of this.taskWriteQueue) {
          if (farmId != null && item.farmId !== farmId) continue
          if (item.state === 'synced') continue
          if (item.type === 'create_task') {
            try {
              const r = await api.post(`/farms/${item.farmId}/tasks`, item.payload)
              localToServer.set(item.clientTaskId, r.data.id)
              this.replaceLocalTask(item.clientTaskId, r.data)
              item.state = 'synced'
              item.lastError = ''
            } catch (err) {
              item.attempts += 1
              item.updatedAt = new Date().toISOString()
              if (isRetryableTaskQueueError(err)) {
                item.state = 'pending'
                item.lastError = err.message || 'network retry'
              } else {
                item.state = 'failed'
                item.lastError = err.response?.data?.error || err.message || 'sync failed'
                hadFailures = true
                this.markTaskConflict(item.clientTaskId, item.lastError)
              }
            }
            continue
          }

          if (item.type === 'update_task_status') {
            let targetID = item.payload.taskId
            if (!targetID && item.payload.clientTaskId) {
              targetID = localToServer.get(item.payload.clientTaskId) ||
                this.tasks.find((t) => t._offline?.clientTaskId === item.payload.clientTaskId)?.id
            }
            if (typeof targetID !== 'number') {
              continue
            }
            try {
              await api.patch(`/tasks/${targetID}/status`, { status: item.payload.status })
              const t = this.tasks.find((x) => x.id === targetID)
              if (t) {
                t.status = item.payload.status
                t._offline = null
              }
              item.state = 'synced'
              item.lastError = ''
            } catch (err) {
              item.attempts += 1
              item.updatedAt = new Date().toISOString()
              const t = this.tasks.find((x) => x.id === targetID)
              if (isRetryableTaskQueueError(err)) {
                item.state = 'pending'
                item.lastError = err.message || 'network retry'
              } else {
                item.state = 'failed'
                item.lastError = err.response?.data?.error || err.message || 'sync failed'
                hadFailures = true
                if (t) {
                  t._offline = {
                    ...(t._offline || {}),
                    queued: true,
                    pendingSync: false,
                    stale: true,
                    conflict: item.lastError,
                    queueItemId: item.id,
                  }
                }
              }
            }
            continue
          }

          if (item.type === 'create_cost') {
            try {
              const r = await api.post(`/farms/${item.farmId}/costs`, item.payload, {
                headers: { 'Idempotency-Key': item.idempotencyKey },
              })
              const row = r.data
              if (item.receiptDataUrl) {
                const file = dataUrlToFile(item.receiptDataUrl, item.receiptFileName || 'receipt')
                const fd = new FormData()
                fd.append('file', file)
                fd.append('cost_transaction_id', String(row.id))
                await api.post(`/farms/${item.farmId}/cost-receipts`, fd)
              }
              item.state = 'synced'
              item.lastError = ''
            } catch (err) {
              item.attempts += 1
              item.updatedAt = new Date().toISOString()
              if (isRetryableTaskQueueError(err)) {
                item.state = 'pending'
                item.lastError = err.message || 'network retry'
              } else {
                item.state = 'failed'
                item.lastError = err.response?.data?.error || err.message || 'sync failed'
                hadFailures = true
              }
            }
            continue
          }
        }
      } finally {
        this.taskWriteQueue = this.taskWriteQueue.filter((i) => i.state !== 'synced')
        this.persistTaskWriteQueue()
        if (hadFailures) {
          this.taskSyncStatus.lastResult = 'partial_error'
          this.taskSyncStatus.lastMessage = 'Some queued writes need review'
        } else {
          this.taskSyncStatus.lastResult = 'ok'
          this.taskSyncStatus.lastMessage = this.taskWriteQueue.length
            ? `${this.taskWriteQueue.length} write(s) still queued`
            : 'All queued writes synced'
        }
        this.taskQueueBusy = false
      }
    },

    clearTaskQueueItem(queueItemId) {
      this.discardTaskQueueItem(queueItemId)
    },

    retryTaskQueueItem(queueItemId) {
      const item = this.taskWriteQueue.find((i) => i.id === queueItemId)
      if (!item) return false
      item.state = 'pending'
      item.lastError = ''
      item.updatedAt = new Date().toISOString()
      const task = this.findTaskByQueueItemId(queueItemId)
      if (task) {
        task._offline = {
          ...(task._offline || {}),
          queued: true,
          pendingSync: true,
          stale: false,
          conflict: '',
          queueItemId,
        }
      }
      this.persistTaskWriteQueue()
      return true
    },

    discardTaskQueueItem(queueItemId) {
      const item = this.taskWriteQueue.find((i) => i.id === queueItemId)
      if (!item) return false
      this.taskWriteQueue = this.taskWriteQueue.filter((i) => i.id !== queueItemId)
      if (item.type === 'create_task') {
        this.tasks = this.tasks.filter(
          (t) => t.id !== item.clientTaskId && t._offline?.clientTaskId !== item.clientTaskId,
        )
      }
      if (item.type === 'create_cost') {
        /* Costs view merges from queue; dropping the item is enough */
      }
      if (item.type === 'update_task_status') {
        const task = this.findTaskByQueueItemId(queueItemId)
        if (task?._offline) {
          task._offline = null
        }
      }
      this.persistTaskWriteQueue()
      return true
    },

    persistTaskWriteQueue() {
      saveOfflineQueue(this.taskWriteQueue)
    },

    withCostQueueOverlay(serverCosts, farmId) {
      const list = [...(serverCosts || [])]
      for (const item of this.taskWriteQueue) {
        if (item.farmId !== farmId || item.type !== 'create_cost') continue
        if (item.state === 'synced') continue
        if (list.some((t) => t._offline?.clientCostId === item.clientCostId)) continue
        list.unshift(this.optimisticCostFromQueueItem(item))
      }
      return list
    },

    optimisticCostFromQueueItem(item) {
      const p = item.payload || {}
      return {
        id: item.clientCostId,
        farm_id: item.farmId,
        transaction_date: p.transaction_date,
        category: p.category,
        subcategory: p.subcategory ?? null,
        amount: p.amount,
        currency: p.currency,
        description: p.description ?? null,
        document_type: p.document_type ?? null,
        document_reference: p.document_reference ?? null,
        counterparty: p.counterparty ?? null,
        is_income: !!p.is_income,
        receipt_file_id: null,
        created_at: item.createdAt,
        updated_at: item.updatedAt,
        _offline: {
          queued: true,
          pendingSync: item.state !== 'failed',
          stale: item.state === 'failed',
          conflict: item.state === 'failed' ? item.lastError : '',
          queueItemId: item.id,
          clientCostId: item.clientCostId,
          receiptPending: !!item.receiptDataUrl,
        },
      }
    },

    withTaskQueueOverlay(serverTasks, farmId) {
      const list = [...serverTasks]
      for (const item of this.taskWriteQueue) {
        if (item.farmId !== farmId) continue
        if (item.type === 'create_task') {
          if (!list.some((t) => t._offline?.clientTaskId === item.clientTaskId)) {
            list.push(this.optimisticTaskFromCreateItem(item))
          }
          continue
        }
        if (item.type === 'update_task_status') {
          const task = list.find((t) => String(t.id) === String(item.payload.taskId))
          if (task) {
            task.status = item.payload.status
            task._offline = {
              queued: true,
              pendingSync: item.state !== 'failed',
              stale: item.state === 'failed',
              conflict: item.state === 'failed' ? item.lastError : '',
              queueItemId: item.id,
            }
          }
        }
      }
      return list
    },

    optimisticTaskFromCreateItem(item) {
      return {
        id: item.clientTaskId,
        farm_id: item.farmId,
        zone_id: item.payload.zone_id ?? null,
        schedule_id: item.payload.schedule_id ?? null,
        title: item.payload.title,
        description: item.payload.description ?? null,
        task_type: item.payload.task_type ?? null,
        status: 'todo',
        priority: item.payload.priority ?? 1,
        due_date: item.payload.due_date ?? null,
        created_at: item.createdAt,
        updated_at: item.updatedAt,
        _offline: {
          queued: true,
          pendingSync: item.state !== 'failed',
          stale: item.state === 'failed',
          conflict: item.state === 'failed' ? item.lastError : '',
          queueItemId: item.id,
          clientTaskId: item.clientTaskId,
        },
      }
    },

    replaceLocalTask(clientTaskId, serverTask) {
      const idx = this.tasks.findIndex((t) => t.id === clientTaskId || t._offline?.clientTaskId === clientTaskId)
      if (idx >= 0) {
        this.tasks[idx] = serverTask
      } else {
        this.tasks.push(serverTask)
      }
    },

    markTaskConflict(clientTaskId, message) {
      const t = this.tasks.find((x) => x.id === clientTaskId || x._offline?.clientTaskId === clientTaskId)
      if (!t) return
      t._offline = {
        ...(t._offline || {}),
        queued: true,
        pendingSync: false,
        stale: true,
        conflict: message || 'sync conflict',
      }
    },

    findTaskByQueueItemId(queueItemId) {
      return this.tasks.find((t) => t._offline?.queueItemId === queueItemId)
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

    async loadMixingEvents(farmId) {
      const r = await api.get(`/farms/${farmId}/fertigation/mixing-events`)
      return Array.isArray(r.data) ? r.data : []
    },

    async loadMixingEventComponents(farmId, mixingEventId) {
      const r = await api.get(`/farms/${farmId}/fertigation/mixing-events/${mixingEventId}/components`)
      return Array.isArray(r.data) ? r.data : []
    },

    async createMixingEvent(farmId, payload) {
      const r = await api.post(`/farms/${farmId}/fertigation/mixing-events`, payload)
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

    // Plants (crop tracking)
    async loadPlants(farmId) {
      const r = await api.get(`/farms/${farmId}/plants`)
      return Array.isArray(r.data) ? r.data : []
    },

    async getPlant(id) {
      const r = await api.get(`/plants/${id}`)
      return r.data
    },

    async createPlant(farmId, data) {
      const r = await api.post(`/farms/${farmId}/plants`, data)
      return r.data
    },

    async updatePlant(id, data) {
      const r = await api.put(`/plants/${id}`, data)
      return r.data
    },

    async deletePlant(id) {
      await api.delete(`/plants/${id}`)
    },

    // Animal husbandry (Phase 20.8)
    async loadAnimalGroups(farmId) {
      const r = await api.get(`/farms/${farmId}/animal-groups`)
      return Array.isArray(r.data) ? r.data : []
    },

    async getAnimalGroup(id) {
      const r = await api.get(`/animal-groups/${id}`)
      return r.data
    },

    async createAnimalGroup(farmId, data) {
      const r = await api.post(`/farms/${farmId}/animal-groups`, data)
      return r.data
    },

    async updateAnimalGroup(id, data) {
      const r = await api.put(`/animal-groups/${id}`, data)
      return r.data
    },

    async archiveAnimalGroup(id, reason) {
      const r = await api.patch(`/animal-groups/${id}/archive`, { archived_reason: reason || null })
      return r.data
    },

    async deleteAnimalGroup(id) {
      await api.delete(`/animal-groups/${id}`)
    },

    async loadLifecycleEvents(groupId) {
      const r = await api.get(`/animal-groups/${groupId}/lifecycle-events`)
      return Array.isArray(r.data) ? r.data : []
    },

    async createLifecycleEvent(groupId, data) {
      const r = await api.post(`/animal-groups/${groupId}/lifecycle-events`, data)
      return r.data
    },

    async deleteLifecycleEvent(id) {
      await api.delete(`/lifecycle-events/${id}`)
    },

    // Aquaponics loops (Phase 20.8)
    async loadAquaponicsLoops(farmId) {
      const r = await api.get(`/farms/${farmId}/aquaponics-loops`)
      return Array.isArray(r.data) ? r.data : []
    },

    async createAquaponicsLoop(farmId, data) {
      const r = await api.post(`/farms/${farmId}/aquaponics-loops`, data)
      return r.data
    },

    async updateAquaponicsLoop(id, data) {
      const r = await api.put(`/aquaponics-loops/${id}`, data)
      return r.data
    },

    async deleteAquaponicsLoop(id) {
      await api.delete(`/aquaponics-loops/${id}`)
    },

    // Commons Catalog
    async loadCatalog({ q = '', limit = 50, offset = 0 } = {}) {
      const params = new URLSearchParams({ limit: String(limit), offset: String(offset) })
      if (q) params.set('q', q)
      const r = await api.get(`/commons/catalog?${params}`)
      return Array.isArray(r.data) ? r.data : []
    },

    async getCatalogEntry(slug) {
      const r = await api.get(`/commons/catalog/${encodeURIComponent(slug)}`)
      return r.data
    },

    async loadCatalogImports(farmId) {
      const r = await api.get(`/farms/${farmId}/commons/catalog-imports`)
      return Array.isArray(r.data) ? r.data : []
    },

    async importCatalogEntry(farmId, slug, note) {
      const body = { slug }
      if (note) body.note = note
      const r = await api.post(`/farms/${farmId}/commons/catalog-imports`, body)
      return r.data
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

    async loadCoaMappings(farmId) {
      const r = await api.get(`/farms/${farmId}/finance/coa-mappings`)
      return Array.isArray(r.data) ? r.data : []
    },

    async saveCoaMappings(farmId, mappings) {
      const r = await api.put(`/farms/${farmId}/finance/coa-mappings`, { mappings })
      return Array.isArray(r.data) ? r.data : []
    },

    async resetCoaMappingCategory(farmId, category) {
      const r = await api.delete(`/farms/${farmId}/finance/coa-mappings/${encodeURIComponent(category)}`)
      return Array.isArray(r.data) ? r.data : []
    },

    async resetCoaMappingsAll(farmId) {
      const r = await api.delete(`/farms/${farmId}/finance/coa-mappings`)
      return Array.isArray(r.data) ? r.data : []
    },

    async loadCosts(farmId, { limit = 50, offset = 0 } = {}) {
      const r = await api.get(`/farms/${farmId}/costs?limit=${limit}&offset=${offset}`)
      const list = Array.isArray(r.data) ? r.data : []
      return this.withCostQueueOverlay(list, farmId)
    },

    /**
     * @param {object} data cost create body
     * @param {{ receiptFile?: File }=} options when offline or flaky network, optional receipt is queued as data URL
     */
    async createCost(farmId, data, options = {}) {
      const { receiptFile } = options
      const idempotencyKey =
        globalThis.crypto && typeof globalThis.crypto.randomUUID === 'function'
          ? globalThis.crypto.randomUUID()
          : `${Date.now()}-${Math.random()}`
      let receiptDataUrl = ''
      let receiptFileName = ''
      if (receiptFile) {
        receiptDataUrl = await fileToDataUrl(receiptFile)
        receiptFileName = receiptFile.name || 'receipt'
      }
      const tryOnline = typeof navigator === 'undefined' || navigator.onLine
      if (tryOnline) {
        try {
          const r = await api.post(`/farms/${farmId}/costs`, data, {
            headers: { 'Idempotency-Key': idempotencyKey },
          })
          const row = r.data
          if (receiptFile) {
            await this.uploadCostReceipt(farmId, receiptFile, row.id)
          }
          return row
        } catch (err) {
          if (!isRetryableTaskQueueError(err)) throw err
        }
      }
      const item = makeCreateCostItem(farmId, data, {
        idempotencyKey,
        receiptDataUrl,
        receiptFileName,
      })
      this.taskWriteQueue.push(item)
      this.persistTaskWriteQueue()
      return this.optimisticCostFromQueueItem(item)
    },

    async updateCost(id, data) {
      const r = await api.put(`/costs/${id}`, data)
      return r.data
    },

    async deleteCost(id) {
      await api.delete(`/costs/${id}`)
    },

    async uploadCostReceipt(farmId, file, costTransactionId) {
      const fd = new FormData()
      fd.append('file', file)
      if (costTransactionId != null) {
        fd.append('cost_transaction_id', String(costTransactionId))
      }
      const r = await api.post(`/farms/${farmId}/cost-receipts`, fd)
      return r.data
    },

    /**
     * @param {number} farmId
     * @param {boolean | { insert_commons_opt_in: boolean, insert_commons_require_approval?: boolean }} opts
     */
    async setInsertCommonsOptIn(farmId, opts) {
      const payload =
        typeof opts === 'boolean'
          ? { insert_commons_opt_in: opts }
          : {
              insert_commons_opt_in: opts.insert_commons_opt_in,
              ...(opts.insert_commons_require_approval !== undefined
                ? { insert_commons_require_approval: opts.insert_commons_require_approval }
                : {}),
            }
      const r = await api.patch(`/farms/${farmId}/insert-commons/opt-in`, payload)
      if (this.farm && this.farm.id === farmId) {
        this.farm = r.data
      }
      return r.data
    },

    /** @param {number} farmId */
    async previewInsertCommons(farmId) {
      const r = await api.get(`/farms/${farmId}/insert-commons/preview`)
      return r.data
    },

    async listInsertCommonsBundles(farmId, { status = '', limit = 25, offset = 0 } = {}) {
      const params = new URLSearchParams({
        limit: String(limit),
        offset: String(offset),
      })
      if (status) params.set('status', status)
      const r = await api.get(`/farms/${farmId}/insert-commons/bundles?${params}`)
      return Array.isArray(r.data) ? r.data : []
    },

    async approveInsertCommonsBundle(farmId, bundleId, body = {}) {
      const r = await api.post(
        `/farms/${farmId}/insert-commons/bundles/${bundleId}/approve`,
        body,
        { timeout: 35000 },
      )
      return r.data
    },

    async rejectInsertCommonsBundle(farmId, bundleId, { note }) {
      const r = await api.post(`/farms/${farmId}/insert-commons/bundles/${bundleId}/reject`, { note })
      return r.data
    },

    async retryInsertCommonsBundleDeliver(farmId, bundleId) {
      const r = await api.post(
        `/farms/${farmId}/insert-commons/bundles/${bundleId}/deliver`,
        {},
        { timeout: 35000 },
      )
      return r.data
    },

    async downloadInsertCommonsBundleExport(farmId, bundleId, format = 'ingest') {
      const r = await api.get(`/farms/${farmId}/insert-commons/bundles/${bundleId}/export`, {
        params: { format },
        responseType: 'blob',
        timeout: 60000,
      })
      const cd = r.headers['content-disposition']
      let filename = `insert-commons-bundle-${bundleId}.json`
      if (cd) {
        const m = /filename="([^"]+)"/.exec(cd)
        if (m) filename = m[1]
      }
      const url = URL.createObjectURL(r.data)
      const a = document.createElement('a')
      a.href = url
      a.download = filename
      a.click()
      URL.revokeObjectURL(url)
    },

    async insertCommonsSync(farmId) {
      const idem =
        (globalThis.crypto && typeof globalThis.crypto.randomUUID === 'function')
          ? globalThis.crypto.randomUUID()
          : `${Date.now()}-${Math.random()}`
      const r = await api.post(`/farms/${farmId}/insert-commons/sync`, {}, {
        headers: { 'Idempotency-Key': idem },
      })
      return r.data
    },

    async listInsertCommonsSyncEvents(farmId, { limit = 10, offset = 0 } = {}) {
      const r = await api.get(`/farms/${farmId}/insert-commons/sync-events?limit=${limit}&offset=${offset}`)
      return Array.isArray(r.data) ? r.data : []
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

    async createTaskFromAlert(alertId, overrides = {}) {
      const r = await api.post(`/alerts/${alertId}/create-task`, overrides || {})
      const task = r.data
      // Optimistically keep the local task list in sync so the Alerts page
      // can render the "→ Task #N" badge without a full reload.
      if (task && task.id != null) {
        const exists = this.tasks.some((t) => t.id === task.id)
        if (!exists) this.tasks = [...this.tasks, task]
      }
      return task
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
