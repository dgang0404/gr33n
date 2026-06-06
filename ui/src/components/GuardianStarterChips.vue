<template>
  <div
    v-if="capabilities.aiEnabled && starters.length"
    class="flex flex-wrap gap-2"
    role="group"
    aria-label="Guardian conversation starters"
    data-test="guardian-starter-chips"
  >
    <button
      v-for="s in starters"
      :key="s.id"
      type="button"
      class="text-xs px-3 py-2 min-h-[44px] sm:min-h-0 rounded-full border border-green-600/85 bg-green-950/55 text-green-200 hover:bg-green-900/65 hover:text-green-100 transition-colors"
      :class="focusRingClass"
      :data-test="`guardian-starter-${s.id}`"
      :aria-label="`Ask Guardian: ${s.label}`"
      @click="pickStarter(s)"
    >
      {{ s.label }}
    </button>
  </div>
</template>

<script setup>
import { useGuardianPanelStore } from '../stores/guardianPanel'
import { useCapabilitiesStore } from '../stores/capabilities'
import { FARMER_FOCUS_RING } from '../lib/farmerA11y.js'

const focusRingClass = FARMER_FOCUS_RING

const props = defineProps({
  starters: { type: Array, default: () => [] },
})

const guardianPanel = useGuardianPanelStore()
const capabilities = useCapabilitiesStore()

function pickStarter(s) {
  guardianPanel.openDrawer({
    prefilledMessage: s.message,
    contextRef: s.contextRef ?? null,
    setupMode: !!s.setupMode,
  })
}
</script>
