<template>
  <Teleport to="body">
    <button
      v-if="showTab"
      type="button"
      class="guardian-edge-tab fixed z-[35] flex flex-col items-center justify-center gap-1 rounded-l-xl border border-r-0 border-green-700/60 bg-zinc-950/95 text-green-400 shadow-lg shadow-black/40 backdrop-blur-sm transition-[transform,box-shadow,background-color] duration-200 ease-out hover:bg-green-950/90 hover:shadow-green-900/30 focus:outline-none focus-visible:ring-2 focus-visible:ring-green-500 focus-visible:ring-offset-2 focus-visible:ring-offset-zinc-950"
      :class="compact ? 'guardian-edge-tab--compact' : 'guardian-edge-tab--desktop'"
      title="Open Farm Guardian"
      aria-label="Open Farm Guardian"
      data-test="guardian-edge-tab"
      @click="guardianPanel.toggle()"
    >
      <span class="text-lg leading-none guardian-edge-tab-icon" aria-hidden="true">✨</span>
      <span
        v-if="guardianChat.streaming"
        class="h-2 w-2 rounded-full bg-green-500 animate-pulse"
        data-test="guardian-edge-thinking"
        aria-hidden="true"
      />
    </button>
  </Teleport>
</template>

<script setup>
import { computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useGuardianChatStore } from '../stores/guardianChat'
import { useGuardianPanelStore } from '../stores/guardianPanel'
import { useCapabilitiesStore } from '../stores/capabilities'

defineProps({
  /** When true, sit below the top bar instead of vertically centered (mobile). */
  compact: { type: Boolean, default: false },
})

const route = useRoute()
const guardianPanel = useGuardianPanelStore()
const guardianChat = useGuardianChatStore()
const capabilities = useCapabilitiesStore()

const aiAvailable = computed(() => capabilities.loaded && !capabilities.isLite)

/** Full-page /chat already is Guardian — hide duplicate edge chrome. */
const onChatPage = computed(() => route.path === '/chat' || route.path.startsWith('/chat/'))

const showTab = computed(() => aiAvailable.value && !guardianPanel.open && !onChatPage.value)

onMounted(async () => {
  if (!capabilities.loaded) await capabilities.fetch()
})
</script>

<style scoped>
.guardian-edge-tab {
  left: auto;
  right: 0;
}

.guardian-edge-tab--desktop {
  top: 50%;
  transform: translateY(-50%);
  padding: 0.75rem 0.5rem;
  min-width: 2.5rem;
}

.guardian-edge-tab--desktop:hover {
  transform: translate(-4px, -50%);
}

.guardian-edge-tab--compact {
  top: 4.75rem;
  padding: 0.5rem;
  min-width: 2.25rem;
}

.guardian-edge-tab--compact:hover {
  transform: translateX(-4px);
}

.guardian-edge-tab:hover .guardian-edge-tab-icon {
  animation: guardian-wiggle 0.45s ease-in-out;
}

@keyframes guardian-wiggle {
  0%, 100% { transform: rotate(0deg) scale(1); }
  20% { transform: rotate(-8deg) scale(1.08); }
  40% { transform: rotate(8deg) scale(1.08); }
  60% { transform: rotate(-4deg) scale(1.04); }
  80% { transform: rotate(4deg) scale(1.04); }
}
</style>
