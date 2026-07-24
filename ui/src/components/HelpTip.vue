<template>
  <span class="inline-flex items-center">
    <button
      ref="triggerRef"
      type="button"
      @mouseenter="open"
      @mouseleave="scheduleClose"
      @click.stop="toggle"
      @focus="open"
      @blur="scheduleClose"
      class="inline-flex items-center justify-center w-4 h-4 rounded-full bg-zinc-800 border border-zinc-700 text-zinc-500 hover:text-zinc-300 hover:border-zinc-600 text-[10px] leading-none transition-colors cursor-help ml-1 shrink-0"
      aria-label="Help"
    >?</button>
    <Teleport to="body">
      <Transition
        enter-active-class="transition duration-150 ease-out"
        enter-from-class="opacity-0 scale-95"
        enter-to-class="opacity-100 scale-100"
        leave-active-class="transition duration-100 ease-in"
        leave-from-class="opacity-100 scale-100"
        leave-to-class="opacity-0 scale-95"
      >
        <div
          v-if="show"
          ref="popoverRef"
          class="fixed z-[9999] w-72 max-w-[calc(100vw-1.5rem)] px-3 py-2 text-xs text-zinc-300 bg-zinc-800 border border-zinc-700 rounded-lg shadow-xl pointer-events-auto"
          :style="popoverStyle"
          @mouseenter="cancelClose"
          @mouseleave="scheduleClose"
        >
          <slot>{{ text }}</slot>
        </div>
      </Transition>
    </Teleport>
  </span>
</template>

<script setup>
import { ref, computed, watch, onBeforeUnmount, nextTick } from 'vue'

const props = defineProps({
  text: { type: String, default: '' },
  position: { type: String, default: 'top', validator: (v) => ['top', 'bottom', 'left', 'right'].includes(v) },
})

const show = ref(false)
const triggerRef = ref(null)
const popoverRef = ref(null)
const coords = ref({ top: 0, left: 0 })
let closeTimer = null

const popoverStyle = computed(() => ({
  top: `${coords.value.top}px`,
  left: `${coords.value.left}px`,
}))

function updatePosition() {
  const trigger = triggerRef.value
  const popover = popoverRef.value
  if (!trigger || !popover) return

  const rect = trigger.getBoundingClientRect()
  const pop = popover.getBoundingClientRect()
  const gap = 8
  const margin = 8
  let top = rect.bottom + gap
  let left = rect.left + rect.width / 2 - pop.width / 2

  if (props.position === 'top') {
    top = rect.top - pop.height - gap
  } else if (props.position === 'left') {
    top = rect.top + rect.height / 2 - pop.height / 2
    left = rect.left - pop.width - gap
  } else if (props.position === 'right') {
    top = rect.top + rect.height / 2 - pop.height / 2
    left = rect.right + gap
  }

  left = Math.max(margin, Math.min(left, window.innerWidth - pop.width - margin))
  top = Math.max(margin, Math.min(top, window.innerHeight - pop.height - margin))
  coords.value = { top, left }
}

function bindViewportListeners() {
  window.addEventListener('scroll', updatePosition, true)
  window.addEventListener('resize', updatePosition)
}

function unbindViewportListeners() {
  window.removeEventListener('scroll', updatePosition, true)
  window.removeEventListener('resize', updatePosition)
}

async function open() {
  cancelClose()
  show.value = true
  await nextTick()
  updatePosition()
  bindViewportListeners()
}

function close() {
  show.value = false
  unbindViewportListeners()
}

function scheduleClose() {
  cancelClose()
  closeTimer = setTimeout(close, 120)
}

function cancelClose() {
  if (closeTimer) {
    clearTimeout(closeTimer)
    closeTimer = null
  }
}

function toggle() {
  if (show.value) close()
  else open()
}

watch(show, async (visible) => {
  if (visible) {
    await nextTick()
    updatePosition()
  }
})

onBeforeUnmount(() => {
  cancelClose()
  unbindViewportListeners()
})
</script>
