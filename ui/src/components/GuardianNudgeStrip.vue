<template>
  <div
    v-if="guardianPanel.showNudgeStrip"
    class="rounded-lg border border-amber-800/60 bg-amber-950/40 px-3 py-2.5 flex flex-col gap-2"
    data-test="guardian-nudge-strip"
  >
    <p class="text-xs text-amber-100/90 leading-snug flex items-start gap-2">
      <span class="text-amber-400 shrink-0" aria-hidden="true">⚠</span>
      <span>{{ guardianPanel.activeNudge.message }}</span>
    </p>
    <div class="flex items-center gap-2">
      <button
        type="button"
        class="text-xs font-medium px-2.5 py-1 rounded bg-amber-900/60 border border-amber-700/80 text-amber-100 hover:bg-amber-900/80"
        data-test="guardian-nudge-review"
        @click="onReview"
      >
        Review
      </button>
      <button
        type="button"
        class="text-xs px-2.5 py-1 rounded text-amber-200/80 hover:text-amber-100 hover:bg-amber-950/60"
        data-test="guardian-nudge-dismiss"
        @click="guardianPanel.dismissNudge()"
      >
        Dismiss
      </button>
    </div>
  </div>
</template>

<script setup>
import { useGuardianPanelStore } from '../stores/guardianPanel'
import { buildNudgeReviewPayload } from '../lib/guardianNudge.js'

const emit = defineEmits(['review'])

const guardianPanel = useGuardianPanelStore()

function onReview() {
  const payload = buildNudgeReviewPayload(guardianPanel.activeNudge)
  if (!payload) return
  guardianPanel.clearNudgeAfterReview()
  emit('review', payload)
}
</script>
