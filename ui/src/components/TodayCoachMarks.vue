<template>
  <div
    v-if="visible && currentStep"
    class="fixed z-40 pointer-events-none"
    :class="anchorClass"
    data-test="today-coach-marks"
    role="status"
    :aria-live="reduceMotion ? 'polite' : 'off'"
  >
    <div
      class="pointer-events-auto rounded-xl border border-green-800/60 bg-zinc-900/95 shadow-xl backdrop-blur-sm px-4 py-3 space-y-2 max-w-sm"
      :class="panelMotionClass"
    >
      <div class="flex items-start justify-between gap-2">
        <div class="min-w-0">
          <p class="text-[10px] uppercase tracking-widest text-green-400/90">
            Tip {{ stepIndex + 1 }} of {{ steps.length }}
          </p>
          <h4 class="text-sm font-semibold text-white mt-0.5">{{ currentStep.title }}</h4>
        </div>
        <button
          type="button"
          class="shrink-0 text-zinc-500 hover:text-zinc-300 text-lg leading-none min-h-[44px] min-w-[44px] flex items-center justify-center"
          aria-label="Dismiss tips"
          data-test="today-coach-dismiss"
          @click="dismiss"
        >
          ×
        </button>
      </div>
      <p class="text-xs text-zinc-300 leading-relaxed">{{ currentStep.body }}</p>
      <div class="flex items-center justify-end gap-2 pt-1">
        <button
          v-if="!isLast"
          type="button"
          class="min-h-[44px] px-3 text-xs text-zinc-400 hover:text-zinc-200"
          data-test="today-coach-skip"
          @click="dismiss"
        >
          Skip
        </button>
        <button
          type="button"
          class="min-h-[44px] px-4 py-2 rounded-lg text-xs font-medium bg-green-900/60 text-green-300 border border-green-800 hover:bg-green-900/80"
          data-test="today-coach-next"
          @click="advance"
        >
          {{ isLast ? 'Got it' : 'Next' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import {
  buildTodayCoachSteps,
  isNarrowTodayViewport,
  isTodayCoachDone,
  markTodayCoachDone,
  todayCoachTransitionClass,
} from '../lib/farmTodayCoachMarks.js'

const props = defineProps({
  enabled: { type: Boolean, default: false },
  hasAttention: { type: Boolean, default: false },
})

const stepIndex = ref(0)
const viewportWidth = ref(typeof window !== 'undefined' ? window.innerWidth : 1280)
const reduceMotion = ref(false)
const dismissed = ref(isTodayCoachDone())

const steps = computed(() => buildTodayCoachSteps({
  hasAttention: props.hasAttention,
  narrowViewport: isNarrowTodayViewport(viewportWidth.value),
}))

const currentStep = computed(() => steps.value[stepIndex.value] || null)
const isLast = computed(() => stepIndex.value >= steps.value.length - 1)
const visible = computed(() => props.enabled && !dismissed.value && steps.value.length > 0)

const panelMotionClass = computed(() => todayCoachTransitionClass(reduceMotion.value))

const anchorClass = computed(() => {
  const target = currentStep.value?.target
  if (target === 'farm-today-attention') {
    return 'left-4 right-4 md:left-6 md:right-auto bottom-6 md:max-w-sm'
  }
  if (target === 'farm-site-strip') {
    return 'left-4 right-4 md:left-6 md:right-auto top-36 md:top-40 md:max-w-sm'
  }
  return 'left-4 right-4 md:left-1/2 md:-translate-x-1/2 bottom-6 md:max-w-md'
})

function dismiss() {
  dismissed.value = true
  markTodayCoachDone()
  clearHighlight()
}

function advance() {
  if (isLast.value) {
    dismiss()
    return
  }
  stepIndex.value += 1
}

function clearHighlight() {
  document.querySelectorAll('[data-today-coach-highlight]').forEach((el) => {
    el.removeAttribute('data-today-coach-highlight')
  })
}

function applyHighlight() {
  clearHighlight()
  const target = currentStep.value?.target
  if (!target) return
  const el = document.querySelector(`[data-test="${target}"]`)
  if (el) el.setAttribute('data-today-coach-highlight', 'true')
}

function onResize() {
  viewportWidth.value = window.innerWidth
}

onMounted(() => {
  reduceMotion.value = window.matchMedia('(prefers-reduced-motion: reduce)').matches
  window.addEventListener('resize', onResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', onResize)
  clearHighlight()
})

watch([currentStep, visible], () => {
  if (visible.value) applyHighlight()
  else clearHighlight()
}, { immediate: true })
</script>

<style scoped>
:global([data-today-coach-highlight="true"]) {
  outline: 2px solid rgb(74 222 128 / 0.55);
  outline-offset: 3px;
  border-radius: 0.75rem;
}
</style>
