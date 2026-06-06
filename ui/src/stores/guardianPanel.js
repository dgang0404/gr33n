import { defineStore } from 'pinia'
import { routeContextRefFromRoute } from '../lib/guardianRouteRef.js'

/**
 * Phase 29 WS1 — global Farm Guardian slide-out panel state.
 *
 * Keeps drawer open/close, prefilled prompts (WS6 entry points), contextual
 * refs, and the active chat session_id across routes and drawer toggles.
 */
export const useGuardianPanelStore = defineStore('guardianPanel', {
  state: () => ({
    open: false,
    drawerTab: 'chat', // 'chat' | 'pending' — Phase 30 WS1
    prefilledMessage: '',
    contextRef: null, // { type: 'alert'|'crop_cycle'|'zone', id, ... } — WS6
    routeRef: null, // { type: 'route', path, name } — Phase 32 WS1
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

    /** Sync current Vue route for grounded chat honesty (Phase 32 WS1). */
    setRouteFromRouter(route) {
      this.routeRef = routeContextRefFromRoute(route)
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
