import { defineStore } from 'pinia'
import api from '../api'
import { refinePrefillForProposal } from '../lib/guardianRefine.js'
import { routeContextRefFromRoute } from '../lib/guardianRouteRef.js'
import { useGuardianReadinessStore } from './guardianReadiness'

const NAV_HISTORY_MAX = 3

/**
 * Phase 29 WS1 — global Farm Guardian slide-out panel state.
 *
 * Keeps drawer open/close, prefilled prompts (WS6 entry points), contextual
 * refs, active chat session_id, and navigation history across routes.
 *
 * Phase 52: navHistory tracks the last NAV_HISTORY_MAX routes visited so the
 * Guardian receives a breadcrumb trail — it knows both where the user IS and
 * where they came from, eliminating the need for starters to embed "I'm on page X".
 */
export const useGuardianPanelStore = defineStore('guardianPanel', {
  state: () => ({
    open: false,
    drawerTab: 'chat', // 'chat' | 'pending' — Phase 30 WS1
    prefilledMessage: '',
    contextRef: null, // { type: 'alert'|'crop_cycle'|'zone', id, ... } — WS6
    routeRef: null, // { type: 'route', path, name } — Phase 32 WS1
    navHistory: [], // [{ type:'route', path, name }, ...] — previous pages (most recent first, excl. current)
    activeSessionId: '',
    setupMode: false, // Phase 44 WS4 — setup-mode persona for grounded chat
    preferFarmCounsel: false, // Phase 170 — open drawer in Farm counsel mode
    autoSendOnOpen: false, // Phase 170 — send prefilled prompt on drawer open
    refineTick: 0, // increments when Refine is tapped — chat panel loads session + prefill
    viewConversationTick: 0, // increments when View conversation is tapped — load session, no prefill
    activeNudge: null, // Phase 61 — { category, message, severity, action_route, nudge_id }
    snoozedNudgeCategories: [], // session-only dismiss/snooze
    nudgeLoading: false,
  }),

  getters: {
    /** Amber dot on edge tab / TopBar when a nudge is pending and panel is closed. */
    showNudgeDot(state) {
      if (!state.activeNudge || state.open) return false
      return !state.snoozedNudgeCategories.includes(state.activeNudge.category)
    },

    /** Phase 137 — critical nudge pre-warms Farm counsel; dot pulses while stirring. */
    criticalNudgePending(state) {
      return state.activeNudge?.category === 'critical_alert' && !state.open
        && !state.snoozedNudgeCategories.includes('critical_alert')
    },

    /** Nudge strip inside the Guardian panel (above starters). */
    showNudgeStrip(state) {
      if (!state.activeNudge) return false
      return !state.snoozedNudgeCategories.includes(state.activeNudge.category)
    },
  },

  actions: {
    toggle() {
      this.open = !this.open
    },

    openDrawer(opts = {}) {
      this.open = true
      if (opts.tab === 'pending') this.drawerTab = 'pending'
      else if (opts.tab === 'chat') this.drawerTab = 'chat'
      if (opts.prefilledMessage != null) this.prefilledMessage = opts.prefilledMessage
      if (opts.contextRef != null) this.contextRef = opts.contextRef
      if (opts.activeSessionId != null) this.activeSessionId = opts.activeSessionId
      if (opts.setupMode != null) this.setupMode = !!opts.setupMode
      this.preferFarmCounsel = !!opts.farmCounsel
      this.autoSendOnOpen = !!opts.autoSend
    },

    openPendingTab() {
      this.open = true
      this.drawerTab = 'pending'
    },

    setDrawerTab(tab) {
      if (tab === 'pending' || tab === 'chat') this.drawerTab = tab
    },

    close() {
      this.open = false
      this.setupMode = false
      this.preferFarmCounsel = false
      this.autoSendOnOpen = false
    },

    clearPrefill() {
      this.prefilledMessage = ''
      this.contextRef = null
      this.preferFarmCounsel = false
      this.autoSendOnOpen = false
    },

    /** Open chat with the proposal's session and a correction prefill (Phase 34 WS5). */
    requestRefine(proposal) {
      if (!proposal) return
      this.prefilledMessage = refinePrefillForProposal(proposal)
      if (proposal.session_id) this.activeSessionId = proposal.session_id
      this.drawerTab = 'chat'
      this.refineTick += 1
    },

    /** Open chat with the proposal's session transcript only (Phase 194). */
    requestViewConversation(proposal) {
      if (!proposal?.session_id) return
      this.prefilledMessage = ''
      this.activeSessionId = proposal.session_id
      this.drawerTab = 'chat'
      this.viewConversationTick += 1
    },

    /**
     * Sync current Vue route for grounded chat context (Phase 32 WS1).
     * Phase 52: also push previous page into navHistory so Guardian sees
     * where the user came from (breadcrumb trail, last 3 pages).
     */
    setRouteFromRouter(route) {
      const next = routeContextRefFromRoute(route)
      // Push old routeRef into history only when it differs from the new route
      // and differs from the head of the history (avoid duplicates on hot-reloads).
      if (
        this.routeRef &&
        this.routeRef.path !== next?.path &&
        this.navHistory[0]?.path !== this.routeRef.path
      ) {
        this.navHistory = [this.routeRef, ...this.navHistory].slice(0, NAV_HISTORY_MAX)
      }
      this.routeRef = next
    },

    /** Entity Ask Guardian ref wins over passive route ref for this turn. */
    chatContextRef() {
      return this.contextRef ?? this.routeRef
    },

    setActiveSessionId(id) {
      this.activeSessionId = id || ''
    },

    async fetchNudge(farmId) {
      if (!farmId) {
        this.activeNudge = null
        return
      }
      this.nudgeLoading = true
      try {
        const r = await api.get(`/farms/${farmId}/guardian-nudge`, {
          validateStatus: (s) => s === 200 || s === 204,
        })
        const nudge = r.status === 200 ? r.data : null
        if (nudge?.category && !this.snoozedNudgeCategories.includes(nudge.category)) {
          this.activeNudge = nudge
          if (nudge.category === 'critical_alert') {
            const readiness = useGuardianReadinessStore()
            void readiness.ensureAwake(farmId, 'farm_counsel')
          }
        } else {
          this.activeNudge = null
        }
      } catch {
        this.activeNudge = null
      } finally {
        this.nudgeLoading = false
      }
    },

    dismissNudge() {
      const cat = this.activeNudge?.category
      if (cat && !this.snoozedNudgeCategories.includes(cat)) {
        this.snoozedNudgeCategories = [...this.snoozedNudgeCategories, cat]
      }
      this.activeNudge = null
    },

    clearNudgeAfterReview() {
      const cat = this.activeNudge?.category
      if (cat && !this.snoozedNudgeCategories.includes(cat)) {
        this.snoozedNudgeCategories = [...this.snoozedNudgeCategories, cat]
      }
      this.activeNudge = null
    },
  },
})
