<template>
  <div v-if="visible" class="rounded-lg bg-gray-800/40 mb-1" data-test="guardian-nav-launch">
    <button
      type="button"
      class="w-full flex items-center rounded-lg text-sm text-green-400 hover:text-green-300 hover:bg-gray-800 transition-colors"
      :class="collapsed ? 'justify-center px-0 py-2' : 'gap-3 px-3 py-2'"
      title="Open Farm Guardian (slide-out)"
      aria-label="Open Farm Guardian"
      data-test="guardian-nav-open-drawer"
      @click="openDrawer"
    >
      <span class="text-lg shrink-0" aria-hidden="true">✨</span>
      <span v-if="!collapsed" class="flex-1 min-w-0 text-left font-semibold">Ask gr33n</span>
      <span
        v-if="!collapsed && proposalsStore.pendingCount > 0"
        class="min-w-[1.125rem] h-[1.125rem] px-1 rounded-full bg-amber-600 text-[10px] font-bold text-amber-950 flex items-center justify-center shrink-0"
        data-test="guardian-nav-pending-badge"
      >
        {{ proposalsStore.pendingCount > 9 ? '9+' : proposalsStore.pendingCount }}
      </span>
    </button>
  </div>
</template>

<script setup>
import { computed, onMounted } from 'vue'
import { useCapabilitiesStore } from '../stores/capabilities'
import { useGuardianPanelStore } from '../stores/guardianPanel'
import { useGuardianProposalsStore } from '../stores/guardianProposals'

defineProps({
  collapsed: { type: Boolean, default: false },
})

const capabilities = useCapabilitiesStore()
const guardianPanel = useGuardianPanelStore()
const proposalsStore = useGuardianProposalsStore()

const visible = computed(() => capabilities.loaded && !capabilities.isLite)

function openDrawer() {
  guardianPanel.openDrawer({ tab: 'chat' })
}

onMounted(async () => {
  if (!capabilities.loaded) await capabilities.fetch()
})
</script>
