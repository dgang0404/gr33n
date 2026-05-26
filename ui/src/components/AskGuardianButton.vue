<template>
  <button
    v-if="capabilities.aiEnabled"
    type="button"
    :class="buttonClass"
    :title="title"
    data-test="ask-guardian-button"
    @click.stop="ask"
  >
    ✨ Ask Guardian
  </button>
</template>

<script setup>
import { computed } from 'vue'
import { useGuardianPanelStore } from '../stores/guardianPanel'
import { useCapabilitiesStore } from '../stores/capabilities'

const props = defineProps({
  prefilledMessage: { type: String, required: true },
  contextRef: { type: Object, default: null },
  variant: { type: String, default: 'secondary' }, // secondary | primary
  size: { type: String, default: 'xs' },
  title: { type: String, default: 'Open Farm Guardian with context from this page' },
})

const guardianPanel = useGuardianPanelStore()
const capabilities = useCapabilitiesStore()

const buttonClass = computed(() => {
  const size = props.size === 'sm'
    ? 'text-xs px-3 py-1.5'
    : 'text-[11px] px-2 py-1'
  const palette = props.variant === 'primary'
    ? 'text-green-300 border-green-700 bg-green-950/50 hover:bg-green-900/60 hover:text-green-200'
    : 'text-gr33n-300 border-gr33n-800/70 bg-gr33n-950/40 hover:bg-gr33n-900/50 hover:text-gr33n-200'
  return `${size} font-medium border rounded-lg transition-colors ${palette}`
})

function ask() {
  guardianPanel.openDrawer({
    prefilledMessage: props.prefilledMessage,
    contextRef: props.contextRef,
  })
}
</script>
