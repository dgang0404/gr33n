import { defineStore } from 'pinia'
import { routeContextRefFromRoute } from '../lib/guardianRouteRef.js'

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
  }),

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
    },

    clearPrefill() {
      this.prefilledMessage = ''
      this.contextRef = null
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
  },
})
