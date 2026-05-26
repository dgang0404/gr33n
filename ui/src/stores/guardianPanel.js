import { defineStore } from 'pinia'

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
    activeSessionId: '',
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
    },

    clearPrefill() {
      this.prefilledMessage = ''
      this.contextRef = null
    },

    setActiveSessionId(id) {
      this.activeSessionId = id || ''
    },
  },
})
