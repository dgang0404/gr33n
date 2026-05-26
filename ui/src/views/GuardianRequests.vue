<template>
  <div class="p-4 sm:p-6 max-w-2xl mx-auto space-y-4" data-test="guardian-requests-page">
    <header>
      <h1 class="text-2xl font-bold text-green-400 mb-1">Guardian change requests</h1>
      <p class="text-sm text-zinc-500">
        Pending proposals waiting for your Confirm. Same queue as the Guardian drawer inbox.
      </p>
    </header>

    <section
      v-if="capabilities.isLite"
      class="rounded-xl border border-amber-900/60 bg-amber-950/40 px-4 py-3 text-sm text-amber-200"
    >
      Farm Guardian is not available on this installation (Lite mode).
    </section>

    <GuardianRequestsInbox v-else :show-full-page-link="false" />
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import GuardianRequestsInbox from '../components/GuardianRequestsInbox.vue'
import { useCapabilitiesStore } from '../stores/capabilities'

const capabilities = useCapabilitiesStore()

onMounted(async () => {
  if (!capabilities.loaded) await capabilities.fetch()
})
</script>
