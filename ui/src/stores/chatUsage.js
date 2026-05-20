import { defineStore } from 'pinia'
import api from '../api'

/**
 * Phase 28 WS5 — Farm Guardian token usage store.
 *
 * Backs the "Guardian usage" card on Settings.vue. Calls
 * GET /v1/chat/usage and surfaces the per-user dimension (always) +
 * optional per-farm dimension when a farmId is supplied.
 *
 * The Settings card hides itself when ai_enabled === false OR neither
 * cap is configured (max_tokens === 0 on both dimensions). Capabilities
 * gating is the operator's primary signal — this store is the data
 * loader and renders gracefully when AI is disabled or uncapped.
 *
 * The "fetched" flag flips true once a request resolves (success OR
 * error) so the card can show "Loading…" exactly once.
 */
export const useChatUsageStore = defineStore('chatUsage', {
  state: () => ({
    fetched: false,
    loading: false,
    error: null,
    windowHours: 1,
    aiEnabled: true,
    user: {
      used_tokens: 0,
      max_tokens: 0,
      remaining_tokens: 0,
      pct_used: 0,
      warning_threshold_pct: null,
    },
    farm: null, // null when no farmId was requested
  }),

  getters: {
    // True when either dimension has a configured cap (max > 0).
    hasAnyCap: (state) =>
      (state.user?.max_tokens || 0) > 0 || (state.farm?.max_tokens || 0) > 0,

    // True when either dimension's pct_used >= 0.80 — the Settings
    // card uses this to switch the bar to amber so the operator
    // notices before they 429.
    nearLimit: (state) => {
      const u = (state.user?.pct_used || 0) >= 0.8
      const f = (state.farm?.pct_used || 0) >= 0.8
      return u || f
    },

    // True when AT_LIMIT — usage rounded up >= 1.0. The bar goes red.
    atLimit: (state) =>
      (state.user?.pct_used || 0) >= 1 || (state.farm?.pct_used || 0) >= 1,
  },

  actions: {
    /**
     * Load the usage snapshot. Pass `{ farmId: 7 }` to include the
     * per-farm dimension. Passing nothing or 0 returns user-only.
     *
     * Failures (non-2xx) populate `error` and set `fetched=true` so
     * the card can render a quiet "couldn't load" hint and the UI
     * doesn't retry in a loop.
     */
    async load({ farmId } = {}) {
      this.loading = true
      this.error = null
      try {
        const params = {}
        if (farmId && Number.isFinite(Number(farmId)) && Number(farmId) > 0) {
          params.farm_id = farmId
        }
        const r = await api.get('/v1/chat/usage', { params })
        this.windowHours = r.data?.window_hours ?? 1
        this.aiEnabled = r.data?.ai_enabled !== false
        this.user = r.data?.user || this.user
        this.farm = r.data?.farm || null
      } catch (e) {
        this.error = e?.response?.data?.error || e.message || 'usage fetch failed'
        // Treat a 503 specially — that's "AI disabled at the server"
        // which the capabilities store already surfaces. Leave the
        // user dimension at its zero/default state so the card hides
        // itself via `hasAnyCap`.
        if (e?.response?.status === 503) {
          this.aiEnabled = false
        }
      } finally {
        this.loading = false
        this.fetched = true
      }
      return this.user
    },
  },
})
