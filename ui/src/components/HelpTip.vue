<template>
  <span class="relative inline-flex items-center">
    <button
      type="button"
      @mouseenter="show = true"
      @mouseleave="show = false"
      @click.stop="show = !show"
      class="inline-flex items-center justify-center w-4 h-4 rounded-full bg-zinc-800 border border-zinc-700 text-zinc-500 hover:text-zinc-300 hover:border-zinc-600 text-[10px] leading-none transition-colors cursor-help ml-1"
      aria-label="Help"
    >?</button>
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
        class="absolute z-50 w-64 px-3 py-2 text-xs text-zinc-300 bg-zinc-800 border border-zinc-700 rounded-lg shadow-lg"
        :class="posClass"
      >
        <slot>{{ text }}</slot>
      </div>
    </Transition>
  </span>
</template>

<script setup>
import { ref, computed } from 'vue'

const props = defineProps({
  text: { type: String, default: '' },
  position: { type: String, default: 'top', validator: v => ['top','bottom','left','right'].includes(v) },
})

const show = ref(false)

const posClass = computed(() => ({
  top: 'bottom-full left-1/2 -translate-x-1/2 mb-2',
  bottom: 'top-full left-1/2 -translate-x-1/2 mt-2',
  left: 'right-full top-1/2 -translate-y-1/2 mr-2',
  right: 'left-full top-1/2 -translate-y-1/2 ml-2',
}[props.position]))
</script>
