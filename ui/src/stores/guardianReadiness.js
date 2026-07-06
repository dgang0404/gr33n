import { defineStore } from 'pinia'
import api from '../api'

const POLL_MS = 2000
const MAX_STIR_MS = 30000

/** Phase 129 WS6 — Guardian awakening readiness (polls GET /v1/chat/health). */
export const useGuardianReadinessStore = defineStore('guardianReadiness', {
  state: () => ({
    loaded: false,
    loading: false,
    error: '',
    awakening: null,
    fieldAssistant: null,
    proceduresAvailable: false,
    farmId: null,
    mode: 'farm_counsel',
    lastCheckedAt: null,
    pollTimer: null,
    warmupStarted: false,
    stirTimedOut: false,
    _stirTimer: null,
  }),

  getters: {
    state(state) {
      return state.awakening?.state || 'unknown'
    },
    badgeClass(state) {
      const s = state.awakening?.state
      if (s === 'ready') return 'bg-green-500'
      if (s === 'stirring') return 'bg-amber-500 animate-pulse'
      if (s === 'sleeping') return 'bg-zinc-500'
      if (s === 'unavailable') return 'bg-red-800/80'
      return 'bg-zinc-600'
    },
    showBadge(state) {
      return state.loaded && state.awakening && state.awakening.state !== 'ready'
    },
    isStirring(state) {
      return state.awakening?.state === 'stirring' || !!state.awakening?.warmup_in_progress
    },
    isReady(state) {
      return state.awakening?.state === 'ready'
    },
    hasStirTimedOut(state) {
      return !!state.stirTimedOut
    },
    farmCounselBlocked(state) {
      const s = state.awakening?.state
      if (state.mode !== 'farm_counsel') return false
      if (s === 'busy') return true
      return (s === 'stirring' || !!state.awakening?.warmup_in_progress) && !state.stirTimedOut
    },
    /** Phase 137 — LLM down but guided procedures still work (Phase 37). */
    showOfflineFieldBanner(state) {
      return !!state.proceduresAvailable && state.fieldAssistant?.llm_reachable === false
    },
  },

  actions: {
    async fetchHealth(farmId, mode = 'farm_counsel') {
      this.loading = true
      this.error = ''
      this.farmId = farmId ?? null
      this.mode = mode
      try {
        const params = {}
        if (farmId) params.farm_id = farmId
        if (mode) params.mode = mode
        const { data } = await api.get('/v1/chat/health', { params })
        this.awakening = data?.awakening ?? null
        this.fieldAssistant = data?.field_assistant ?? null
        this.proceduresAvailable = !!data?.procedures_available
        this.lastCheckedAt = Date.now()
        this.loaded = true
      } catch (e) {
        this.error = e.response?.data?.error || e.message || 'Health check failed'
      } finally {
        this.loading = false
      }
    },

    async warmup(farmId, mode = 'farm_counsel', opts = {}) {
      try {
        const body = { mode }
        if (farmId) body.farm_id = Number(farmId)
        if (opts.includeVision) body.include_vision = true
        await api.post('/guardian/warmup', body)
        this.warmupStarted = true
        await this.fetchHealth(farmId, mode)
        this.startPolling(farmId, mode)
      } catch (e) {
        this.error = e.response?.data?.error || e.message || 'Warmup failed'
      }
    },

    async ensureAwake(farmId, mode = 'farm_counsel') {
      await this.fetchHealth(farmId, mode)
      if (this.awakening?.state === 'ready' || this.awakening?.state === 'unavailable') {
        return
      }
      if (!this.warmupStarted && (this.awakening?.state === 'sleeping' || !this.awakening)) {
        await this.warmup(farmId, mode)
        return
      }
      if (this.isStirring) {
        this.startPolling(farmId, mode)
      }
    },

    startPolling(farmId, mode = 'farm_counsel') {
      this.stopPolling()
      this._clearStirTimer()
      this.stirTimedOut = false
      if (this.isStirring) {
        this._stirTimer = setTimeout(() => {
          if (this.isStirring) this.stirTimedOut = true
        }, MAX_STIR_MS)
      }
      this.pollTimer = setInterval(() => {
        void this.fetchHealth(farmId, mode).then(() => {
          if (this.awakening?.state === 'ready' || this.awakening?.state === 'unavailable') {
            this.stopPolling()
          }
        })
      }, POLL_MS)
    },

    _clearStirTimer() {
      if (this._stirTimer) {
        clearTimeout(this._stirTimer)
        this._stirTimer = null
      }
    },

    stopPolling() {
      if (this.pollTimer) {
        clearInterval(this.pollTimer)
        this.pollTimer = null
      }
      this._clearStirTimer()
    },

    resetSession() {
      this.stopPolling()
      this.warmupStarted = false
      this.loaded = false
      this.stirTimedOut = false
      this.awakening = null
      this.fieldAssistant = null
      this.proceduresAvailable = false
      this.error = ''
    },
  },
})
