import { defineStore } from 'pinia'
import api from '../api'

/**
 * Phase 27 WS6 — server-side capability flags surfaced for the UI.
 *
 * `aiEnabled` mirrors the API's AI_ENABLED toggle. When AI is off, the UI hides
 * or disables LLM-backed actions (Ask LLM, Farm Guardian chat) rather than
 * letting calls return 503 mid-flow. `loaded` flips once fetch() has resolved,
 * so views can wait one tick before rendering the Lite/Full label.
 *
 * See docs/plans/archive/phase_27_farm_guardian_ai_layer.md (WS2 + WS6).
 */
export const useCapabilitiesStore = defineStore('capabilities', {
  state: () => ({
    aiEnabled: true,
    visionChatEnabled: false,
    sttLocalEnabled: false,
    weatherForecastAvailable: false,
    weatherProvider: 'off',
    weatherProviderLabel: '',
    loaded: false,
    fetchError: null,
  }),

  getters: {
    isLite: (state) => state.loaded && state.aiEnabled === false,
  },

  actions: {
    async fetch() {
      try {
        const r = await api.get('/capabilities')
        this.aiEnabled = r.data?.ai_enabled !== false
        this.visionChatEnabled = r.data?.vision_chat_enabled === true
        this.sttLocalEnabled = r.data?.stt_local_enabled === true
        this.weatherForecastAvailable = r.data?.weather_forecast_available === true
        this.weatherProvider = r.data?.weather_provider || 'off'
        this.weatherProviderLabel = r.data?.weather_provider_label || ''
        this.fetchError = null
      } catch (e) {
        // Older API builds without /capabilities → treat as AI on (back-compat).
        this.aiEnabled = true
        this.visionChatEnabled = false
        this.sttLocalEnabled = false
        this.weatherForecastAvailable = false
        this.weatherProvider = 'off'
        this.weatherProviderLabel = ''
        this.fetchError = e.message || 'capabilities fetch failed'
      } finally {
        this.loaded = true
      }
      return this.aiEnabled
    },
  },
})
