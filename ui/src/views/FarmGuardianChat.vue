<template>
  <div class="p-4 sm:p-6 max-w-6xl mx-auto space-y-4 flex flex-col min-h-[calc(100vh-8rem)]">
    <header>
      <h1 class="text-2xl font-bold text-green-400 mb-2 flex items-center gap-2">
        Farm Guardian
        <HelpTip position="bottom">
          On-farm assistant grounded in this farm's data when a farm is selected.
          Replies stream in token-by-token via Server-Sent Events. Conversations
          persist server-side; load any prior session from the sidebar.
        </HelpTip>
      </h1>
      <p class="text-sm text-zinc-500">
        Chat and pending change requests live here. On other pages, use the
        <strong>robot tab</strong> on the right edge for a quick slide-out.
      </p>
    </header>

    <section
      v-if="capabilities.isLite"
      data-test="chat-lite-banner"
      class="rounded-xl border border-amber-900/60 bg-amber-950/40 px-4 py-3 text-sm text-amber-200"
    >
      Farm Guardian is not available on this installation.
      Your farm is running in <strong>Lite mode</strong> — all operational features
      remain fully active. Set <code class="text-gr33n-400">AI_ENABLED=true</code>
      on the API and restart to enable chat.
    </section>

    <div v-else class="flex-1 min-h-0 flex flex-col min-w-0 bg-zinc-950/40 border border-zinc-800 rounded-xl overflow-hidden">
      <GuardianTabNav
        v-model="activeTab"
        :pending-count="proposalsStore.pendingCount"
        class="px-3"
      />
      <div class="flex-1 min-h-0 overflow-y-auto p-4">
        <GuardianChatPanel v-show="activeTab === 'chat'" layout="full" />
        <GuardianRequestsInbox
          v-show="activeTab === 'pending'"
          :active="activeTab === 'pending'"
        />
      </div>
      <footer class="px-4 py-2 border-t border-zinc-800 text-[10px] text-zinc-600 shrink-0">
        Guardian proposes changes; you confirm. Confirmed actions appear in the farm audit log.
      </footer>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import HelpTip from '../components/HelpTip.vue'
import GuardianChatPanel from '../components/GuardianChatPanel.vue'
import GuardianRequestsInbox from '../components/GuardianRequestsInbox.vue'
import GuardianTabNav from '../components/GuardianTabNav.vue'
import { useCapabilitiesStore } from '../stores/capabilities'
import { useFarmContextStore } from '../stores/farmContext'
import { useGuardianPanelStore } from '../stores/guardianPanel'
import { useGuardianProposalsStore } from '../stores/guardianProposals'

const route = useRoute()
const router = useRouter()
const capabilities = useCapabilitiesStore()
const farmContext = useFarmContextStore()
const proposalsStore = useGuardianProposalsStore()
const guardianPanel = useGuardianPanelStore()

const activeTab = computed({
  get() {
    return route.query.tab === 'pending' ? 'pending' : 'chat'
  },
  set(tab) {
    const query = tab === 'pending' ? { tab: 'pending' } : {}
    router.replace({ path: '/chat', query })
  },
})

onMounted(async () => {
  if (!capabilities.loaded) await capabilities.fetch()
  if (route.query.setup === '1') {
    guardianPanel.setupMode = true
  }
  if (farmContext.farmId) {
    await proposalsStore.refreshPendingCount(farmContext.farmId)
  }
})

watch(
  () => farmContext.farmId,
  (id) => {
    if (id) proposalsStore.refreshPendingCount(id)
    else proposalsStore.pendingCount = 0
  },
  { immediate: true },
)
</script>
