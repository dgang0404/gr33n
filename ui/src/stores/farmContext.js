import { defineStore } from 'pinia'
import api from '../api'
import { useFarmStore } from './farm'

export const useFarmContextStore = defineStore('farmContext', {
  state: () => ({
    farmId: Number(localStorage.getItem('gr33n_farm_id')) || null,
    farms: [],
  }),

  getters: {
    selectedFarm: (state) => state.farms.find(f => f.id === state.farmId) ?? null,
  },

  actions: {
    async fetchFarms() {
      const r = await api.get('/farms')
      this.farms = Array.isArray(r.data) ? r.data : []
      return this.farms
    },

    async selectFarm(id) {
      this.farmId = id
      localStorage.setItem('gr33n_farm_id', String(id))
      const farmStore = useFarmStore()
      await farmStore.loadAll(id)
    },

    async createFarm(data) {
      const r = await api.post('/farms', data)
      await this.fetchFarms()
      return r.data
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
