import { onUnmounted, watch } from 'vue'

const FOCUSABLE =
  'a[href], button:not([disabled]), textarea:not([disabled]), input:not([disabled]), select:not([disabled]), [tabindex]:not([tabindex="-1"])'

function focusableElements(container) {
  if (!container) return []
  return [...container.querySelectorAll(FOCUSABLE)].filter(
    (el) => !el.closest('[aria-hidden="true"]'),
  )
}

/**
 * Phase 158 — trap Tab focus inside an open dialog and restore focus on close.
 * @param {import('vue').Ref<boolean>} openRef
 * @param {import('vue').Ref<HTMLElement|null>} containerRef
 * @param {{ initialFocusSelector?: string, onEscape?: () => void }} [options]
 */
export function useDialogFocusTrap(openRef, containerRef, options = {}) {
  let previouslyFocused = null

  function onKeyDown(e) {
    if (!openRef.value || !containerRef.value) return

    if (e.key === 'Escape' && options.onEscape) {
      e.preventDefault()
      options.onEscape()
      return
    }

    if (e.key !== 'Tab') return

    const nodes = focusableElements(containerRef.value)
    if (!nodes.length) return

    const first = nodes[0]
    const last = nodes[nodes.length - 1]

    if (e.shiftKey && document.activeElement === first) {
      e.preventDefault()
      last.focus()
    } else if (!e.shiftKey && document.activeElement === last) {
      e.preventDefault()
      first.focus()
    }
  }

  watch(openRef, (open) => {
    if (open) {
      previouslyFocused = document.activeElement
      requestAnimationFrame(() => {
        const container = containerRef.value
        if (!container) return
        const initial = options.initialFocusSelector
          ? container.querySelector(options.initialFocusSelector)
          : focusableElements(container)[0]
        initial?.focus()
      })
      document.addEventListener('keydown', onKeyDown)
    } else {
      document.removeEventListener('keydown', onKeyDown)
      if (previouslyFocused && typeof previouslyFocused.focus === 'function') {
        previouslyFocused.focus()
      }
      previouslyFocused = null
    }
  })

  onUnmounted(() => {
    document.removeEventListener('keydown', onKeyDown)
  })
}
