<template>
  <div v-if="visible" class="rounded-lg bg-gray-800/40 mb-1 relative" data-test="guardian-nav-launch">
    <button
      type="button"
      class="w-full flex items-center rounded-lg text-sm text-green-400 hover:text-green-300 hover:bg-gray-800 transition-colors"
      :class="collapsed ? 'justify-center px-0 py-2' : 'gap-3 px-3 py-2'"
      title="Open Farm Guardian (slide-out)"
      aria-label="Open Farm Guardian"
      data-test="guardian-nav-open-drawer"
      @click="openDrawer"
    >
      <span class="relative text-lg shrink-0" aria-hidden="true">
        ✨
        <span
          v-if="readiness.showBadge"
          class="absolute -bottom-0.5 -right-0.5 w-2 h-2 rounded-full ring-1 ring-gray-900"
          :class="readiness.badgeClass"
          data-test="guardian-readiness-dot"
          aria-hidden="true"
        />
      </span>
      <span v-if="!collapsed" class="flex-1 min-w-0 text-left font-semibold">Ask gr33n</span>
      <!-- Pending count badge lives on TopBar only (Phase 181) — readiness dot stays here -->
    </button>
  </div>
</template>

<script setup>
import { computed, onMounted, watch } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useCapabilitiesStore } from '../stores/capabilities'
import { useFarmContextStore } from '../stores/farmContext'
import { useGuardianPanelStore } from '../stores/guardianPanel'
import { useGuardianProposalsStore } from '../stores/guardianProposals'
import { useGuardianReadinessStore } from '../stores/guardianReadiness'

defineProps({
  collapsed: { type: Boolean, default: false },
})

const auth = useAuthStore()
const capabilities = useCapabilitiesStore()
const farmContext = useFarmContextStore()
const guardianPanel = useGuardianPanelStore()
const proposalsStore = useGuardianProposalsStore()
const readiness = useGuardianReadinessStore()

const visible = computed(() => capabilities.loaded && !capabilities.isLite)

function openDrawer() {
  if (proposalsStore.pendingCount > 0) {
    guardianPanel.openPendingTab()
  } else {
    guardianPanel.openDrawer({ tab: 'chat' })
  }
}

async function bootReadiness() {
  if (!auth.token || !capabilities.aiEnabled || !farmContext.farmId) return
  await readiness.fetchHealth(farmContext.farmId, 'farm_counsel')
  if (readiness.awakening?.state === 'sleeping') {
    await readiness.warmup(farmContext.farmId, 'farm_counsel')
  }
}

onMounted(async () => {
  if (!capabilities.loaded) await capabilities.fetch()
  await bootReadiness()
})

watch(() => farmContext.farmId, () => {
  void bootReadiness()
})
</script>
