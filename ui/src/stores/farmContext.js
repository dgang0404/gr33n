import { defineStore } from 'pinia'
import api from '../api'
import { useFarmStore } from './farm'
import { parseFarmCoordinates } from '../lib/siteWeather.js'

/** Prefer GeoJSON Point when meta has lat/lon (PostGIS EWKB is not UI-parseable). */
function normalizeFarmCoordinates(farm) {
  if (!farm) return farm
  const { latitude, longitude } = parseFarmCoordinates(farm)
  if (!Number.isFinite(latitude) || !Number.isFinite(longitude)) return farm
  const gis = farm.location_gis
  if (typeof gis === 'object' && gis?.type === 'Point' && Array.isArray(gis.coordinates)) {
    return farm
  }
  return {
    ...farm,
    location_gis: { type: 'Point', coordinates: [longitude, latitude] },
  }
}

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
      this.farms = (Array.isArray(r.data) ? r.data : []).map(normalizeFarmCoordinates)
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

    /** Farm admin: apply a named starter template to an existing farm (skipped if already applied). */
    async applyBootstrapTemplate(farmId, template) {
      const r = await api.post(`/farms/${farmId}/bootstrap-template`, { template })
      const farmStore = useFarmStore()
      await farmStore.loadAll(farmId)
      return r.data
    },

    async updateFarm(id, data) {
      const r = await api.put(`/farms/${id}`, data)
      const idx = this.farms.findIndex(f => f.id === id)
      if (idx >= 0) this.farms[idx] = r.data
      return r.data
    },

    /** Phase 66 — set farm site coordinates for offline solar math. */
    async patchSite(id, { latitude, longitude, elevation_m }) {
      const r = await api.patch(`/farms/${id}/site`, { latitude, longitude, elevation_m })
      const lat = Number(latitude)
      const lon = Number(longitude)
      const prev = this.farms.find(f => f.id === id) || {}
      const prevMeta = prev.meta_data && typeof prev.meta_data === 'object' ? { ...prev.meta_data } : {}
      const meta = {
        ...(r.data?.meta_data && typeof r.data.meta_data === 'object' ? r.data.meta_data : prevMeta),
        latitude: lat,
        longitude: lon,
      }
      if (elevation_m == null) delete meta.elevation_m
      else meta.elevation_m = elevation_m
      // API may return PostGIS EWKB for location_gis — keep a GeoJSON Point the UI can parse.
      const farm = {
        ...r.data,
        location_gis: {
          type: 'Point',
          coordinates: [lon, lat],
        },
        meta_data: meta,
      }
      const idx = this.farms.findIndex(f => f.id === id)
      if (idx >= 0) this.farms[idx] = farm
      return farm
    },

    /** Phase 178 — opt in/out of online weather forecast + temp display unit for this farm. */
    async patchWeatherSettings(id, { weather_forecast_enabled, temperature_unit }) {
      const body = {}
      if (weather_forecast_enabled != null) body.weather_forecast_enabled = weather_forecast_enabled
      if (temperature_unit != null) body.temperature_unit = temperature_unit
      const r = await api.patch(`/farms/${id}/weather/settings`, body)
      const idx = this.farms.findIndex(f => f.id === id)
      if (idx >= 0) {
        const prev = this.farms[idx]
        const meta = {
          ...(prev.meta_data && typeof prev.meta_data === 'object' ? prev.meta_data : {}),
          ...(r.data?.meta_data && typeof r.data.meta_data === 'object' ? r.data.meta_data : {}),
        }
        if (weather_forecast_enabled != null) meta.weather_forecast_enabled = weather_forecast_enabled
        if (temperature_unit != null) meta.temperature_unit = temperature_unit
        this.farms[idx] = { ...prev, ...r.data, meta_data: meta }
      }
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
