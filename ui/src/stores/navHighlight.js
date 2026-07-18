import { defineStore } from 'pinia'

/**
 * Phase 49 WS3 — cross-component sidebar highlight.
 *
 * In-page links (e.g. "Feed & water hub →", "Advanced feeding →") set the
 * destination path on hover/focus via the `v-nav-hint` directive; SideNav
 * wiggles the matching sidebar item so users learn where each link lives.
 *
 * @see docs/plans/archive/phase_49_sidebar_nav_polish.plan.md
 */
export const useNavHighlightStore = defineStore('navHighlight', {
  state: () => ({
    /** @type {string | null} normalized path (no query) of the hinted destination */
    route: null,
  }),
  actions: {
    set(route) {
      this.route = route || null
    },
    clear() {
      this.route = null
    },
  },
})
