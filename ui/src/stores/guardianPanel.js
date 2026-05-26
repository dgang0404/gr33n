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
      if (opts.prefilledMessage != null) this.prefilledMessage = opts.prefilledMessage
      if (opts.contextRef != null) this.contextRef = opts.contextRef
      if (opts.activeSessionId != null) this.activeSessionId = opts.activeSessionId
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
