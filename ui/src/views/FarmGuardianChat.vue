<template>
  <div class="p-4 sm:p-6 max-w-6xl mx-auto space-y-6">
    <header>
      <h1 class="text-2xl font-bold text-green-400 mb-2 flex items-center gap-2">
        Farm Guardian
        <HelpTip position="bottom">
          On-farm assistant grounded in this farm's data when a farm is selected.
          Replies stream in token-by-token via Server-Sent Events. Conversations
          persist server-side; load any prior session from the sidebar.
          See <code class="text-gr33n-400">docs/plans/phase_27_farm_guardian_ai_layer.md</code>.
        </HelpTip>
      </h1>
      <p class="text-sm text-zinc-500">
        Multi-turn. Tick <em>Use farm context</em> to ground answers in the
        selected farm's indexed chunks. Up to {{ maxHistoryTurns }} prior turns
        are replayed into each new question. Use the sidebar
        <strong>Guardian</strong> icon or the top bar sparkle to open the slide-out on any page.
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

    <GuardianChatPanel v-else layout="full" />
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import HelpTip from '../components/HelpTip.vue'
import GuardianChatPanel from '../components/GuardianChatPanel.vue'
import { useCapabilitiesStore } from '../stores/capabilities'

const capabilities = useCapabilitiesStore()
const maxHistoryTurns = 20

onMounted(async () => {
  if (!capabilities.loaded) await capabilities.fetch()
})
</script>
