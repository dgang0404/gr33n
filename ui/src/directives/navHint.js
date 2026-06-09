import { useNavHighlightStore } from '../stores/navHighlight'
import { canonicalSidebarPath } from '../lib/workspaces.js'

/**
 * Phase 49 WS3 — `v-nav-hint` directive.
 *
 * Attach to any in-page link or button that navigates somewhere. On hover or
 * keyboard focus it tells the sidebar (via the navHighlight store) which route
 * the element leads to, so the matching sidebar item wiggles. Clears on leave.
 *
 * Accepts a route-location: a string path (`'/feeding'`) or a router object
 * (`{ path: '/feeding', query: {...} }`). Legacy paths map to workspace routes.
 *
 * @see docs/plans/phase_49_sidebar_nav_polish.plan.md
 */

/**
 * @param {string | { path?: string } | null | undefined} value
 * @returns {string | null}
 */
export function resolveHintPath(value) {
  if (!value) return null
  let raw = null
  if (typeof value === 'string') raw = value.split('?')[0] || null
  else if (typeof value === 'object' && typeof value.path === 'string') {
    raw = value.path.split('?')[0] || null
  }
  return raw ? canonicalSidebarPath(raw) : null
}

export const navHint = {
  mounted(el, binding) {
    el.__navHintPath = resolveHintPath(binding.value)
    el.__navHintEnter = () => {
      if (el.__navHintPath) useNavHighlightStore().set(el.__navHintPath)
    }
    el.__navHintLeave = () => useNavHighlightStore().clear()
    el.addEventListener('mouseenter', el.__navHintEnter)
    el.addEventListener('mouseleave', el.__navHintLeave)
    el.addEventListener('focus', el.__navHintEnter)
    el.addEventListener('blur', el.__navHintLeave)
  },
  updated(el, binding) {
    el.__navHintPath = resolveHintPath(binding.value)
  },
  beforeUnmount(el) {
    if (el.__navHintEnter) {
      el.removeEventListener('mouseenter', el.__navHintEnter)
      el.removeEventListener('focus', el.__navHintEnter)
    }
    if (el.__navHintLeave) {
      el.removeEventListener('mouseleave', el.__navHintLeave)
      el.removeEventListener('blur', el.__navHintLeave)
    }
    useNavHighlightStore().clear()
  },
}
