/** Scroll container for Help Library (WorkspaceShell content pane). */
export function helpLibraryScrollRoot() {
  return document.querySelector('.workspace-shell__content')
}

/**
 * Scroll the Help Library content pane to a section anchor.
 * @param {string} sectionId guide | knowledge | symptoms | catalog
 * @param {{ smooth?: boolean }} [opts]
 */
export function scrollToHelpLibrarySection(sectionId, { smooth = true } = {}) {
  const el = document.getElementById(`help-section-${sectionId}`)
  const root = helpLibraryScrollRoot()
  if (!el || !root) return false
  const delta = el.getBoundingClientRect().top - root.getBoundingClientRect().top
  const top = Math.max(0, root.scrollTop + delta - 8)
  if (typeof root.scrollTo === 'function') {
    root.scrollTo({ top, behavior: smooth ? 'smooth' : 'instant' })
  } else {
    root.scrollTop = top
  }
  return true
}
