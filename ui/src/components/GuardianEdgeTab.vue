<template>
  <Teleport to="body">
    <button
      v-if="visible && !guardianPanel.open"
      type="button"
      class="guardian-edge-tab fixed z-30 flex items-center gap-2 rounded-l-xl border border-r-0 border-green-700/60 bg-zinc-950/95 text-green-400 shadow-lg shadow-black/40 backdrop-blur-sm transition-[transform,box-shadow,background-color] duration-200 ease-out hover:bg-green-950/90 hover:shadow-green-900/30 focus:outline-none focus-visible:ring-2 focus-visible:ring-green-500 focus-visible:ring-offset-2 focus-visible:ring-offset-zinc-950"
      :class="compact ? 'guardian-edge-tab--compact right-0 top-[4.75rem] px-2 py-2' : 'guardian-edge-tab--center right-0 top-1/2 px-2.5 py-3'"
      title="Farm Guardian"
      aria-label="Open Farm Guardian"
      data-test="guardian-edge-tab"
      @click="guardianPanel.toggle()"
    >
      <span class="text-lg leading-none guardian-edge-tab-icon" aria-hidden="true">✨</span>
      <span
        class="text-xs font-semibold tracking-wide uppercase hidden sm:inline"
        style="writing-mode: vertical-rl; text-orientation: mixed;"
      >
        Guardian
      </span>
    </button>
  </Teleport>
</template>

<script setup>
import { computed, onMounted } from 'vue'
import { useGuardianPanelStore } from '../stores/guardianPanel'
import { useCapabilitiesStore } from '../stores/capabilities'

defineProps({
  /** When true, sit below the top bar instead of vertically centered (mobile). */
  compact: { type: Boolean, default: false },
})

const guardianPanel = useGuardianPanelStore()
const capabilities = useCapabilitiesStore()

const visible = computed(() => capabilities.loaded && !capabilities.isLite)

onMounted(async () => {
  if (!capabilities.loaded) await capabilities.fetch()
})
</script>

<style scoped>
.guardian-edge-tab--center {
  transform: translateY(-50%);
}

.guardian-edge-tab--center:hover {
  transform: translate(-4px, -50%);
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
