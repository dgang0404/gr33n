<template>
  <div v-if="capabilities.aiEnabled && starters.length" class="flex flex-wrap gap-2" data-test="guardian-starter-chips">
    <button
      v-for="s in starters"
      :key="s.id"
      type="button"
      class="text-xs px-2.5 py-1 rounded-full border border-green-800/70 bg-green-950/40 text-green-300 hover:bg-green-900/50 hover:text-green-200 transition-colors"
      :data-test="`guardian-starter-${s.id}`"
      @click="pickStarter(s)"
    >
      {{ s.label }}
    </button>
  </div>
</template>

<script setup>
import { useGuardianPanelStore } from '../stores/guardianPanel'
import { useCapabilitiesStore } from '../stores/capabilities'

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
