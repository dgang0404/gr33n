import { defineStore } from 'pinia'
import api from '../api'
import { useFarmStore } from './farm'

export const useFarmContextStore = defineStore('farmContext', {
  state: () => ({
    farmId: Number(localStorage.getItem('gr33n_farm_id')) || null,
    farms: [],
    farmSelectionNotice: null,
  }),

  getters: {
    selectedFarm: (state) => state.farms.find(f => f.id === state.farmId) ?? null,
  },

  actions: {
    async fetchFarms() {
      const r = await api.get('/farms')
      this.farms = Array.isArray(r.data) ? r.data : []
      // Recover from stale persisted farm IDs (e.g. after DB reset/reseed).
      // If selected farm is missing, switch to first available farm.
      if (this.farmId && !this.farms.some(f => f.id === this.farmId)) {
        const previous = this.farmId
        const next = this.farms[0]?.id ?? null
        this.farmId = next
        if (next) localStorage.setItem('gr33n_farm_id', String(next))
        else localStorage.removeItem('gr33n_farm_id')
        this.farmSelectionNotice = next
          ? `Selected farm ${previous} was not found. Switched to farm ${next}.`
          : `Selected farm ${previous} was not found. No farms are currently available.`
      }
      return this.farms
    },

    async selectFarm(id) {
      this.farmSelectionNotice = null
      this.farmId = id
      localStorage.setItem('gr33n_farm_id', String(id))
      const farmStore = useFarmStore()
      await farmStore.loadAll(id)
    },

    clearFarmSelectionNotice() {
      this.farmSelectionNotice = null
    },

    async createFarm(data) {
      const r = await api.post('/farms', data)
      await this.fetchFarms()
      const raw = r.data
      if (raw && typeof raw === 'object' && raw.farm != null) {
        return { farm: raw.farm, bootstrap: raw.bootstrap ?? null }
      }
      return { farm: raw, bootstrap: null }
    },

    async updateFarm(id, data) {
      const r = await api.put(`/farms/${id}`, data)
      const idx = this.farms.findIndex(f => f.id === id)
      if (idx >= 0) this.farms[idx] = r.data
      return r.data
    },

    async deleteFarm(id) {
      await api.delete(`/farms/${id}`)
      this.farms = this.farms.filter(f => f.id !== id)
      if (this.farmId === id) {
        const next = this.farms[0]?.id ?? null
        if (next) await this.selectFarm(next)
        else this.farmId = null
      }
    },
  },
})
