<template>
  <div class="space-y-2" data-test="guardian-context-mode-cards">
    <p
      v-if="capabilities.isLite"
      class="text-xs text-amber-200/90 rounded border border-amber-900/50 bg-amber-950/30 px-3 py-2"
      data-test="guardian-mode-lite"
    >
      Lite mode — Pi and dashboard only. Set <code class="text-amber-100/80">AI_ENABLED=true</code> on the API and restart to enable Guardian chat.
    </p>
    <template v-else>
      <p class="text-[10px] uppercase tracking-widest text-zinc-500">How should Guardian answer?</p>
      <p
        v-if="ollamaUnavailable"
        class="text-[11px] text-red-200/90 rounded border border-red-900/50 bg-red-950/30 px-3 py-2"
        data-test="guardian-mode-ollama-down"
      >
        Ollama is not reachable — start Ollama (<code class="text-red-100/80">systemctl start ollama</code> or open the Ollama app), then retry awakening in Settings.
      </p>
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
        <button
          type="button"
          class="text-left rounded-lg border px-3 py-2.5 transition-colors"
          :class="quickSelected ? 'border-green-700 bg-green-950/40 ring-1 ring-green-800/60' : 'border-zinc-700 bg-zinc-950 hover:border-zinc-500'"
          data-test="guardian-mode-quick"
          @click="selectQuick"
        >
          <span class="block text-sm font-medium text-zinc-100">Quick chat</span>
          <span class="block text-[11px] text-zinc-400 mt-0.5">Fast — general horticulture, no farm data</span>
          <span class="inline-flex mt-2 text-[10px] px-1.5 py-0.5 rounded bg-zinc-800 text-zinc-300">Chat only</span>
        </button>
        <button
          type="button"
          class="text-left rounded-lg border px-3 py-2.5 transition-colors"
          :class="farmSelected ? 'border-green-700 bg-green-950/40 ring-1 ring-green-800/60' : 'border-zinc-700 bg-zinc-950 hover:border-zinc-500'"
          :disabled="farmDisabled"
          data-test="guardian-mode-farm-counsel"
          @click="selectFarmCounsel"
        >
          <span class="block text-sm font-medium text-zinc-100">Farm counsel</span>
          <span class="block text-[11px] text-zinc-400 mt-0.5">Reads your farm — may take many minutes on CPU</span>
          <span class="flex flex-wrap gap-1 mt-2">
            <span class="text-[10px] px-1.5 py-0.5 rounded bg-zinc-800 text-zinc-300" title="Live DB snapshot">Snapshot · live</span>
            <span class="text-[10px] px-1.5 py-0.5 rounded bg-zinc-800 text-zinc-300" title="Read-only tools">Read tools · live</span>
            <span class="text-[10px] px-1.5 py-0.5 rounded bg-green-950/50 text-green-300/90" title="Field guides & docs">RAG · guides/notes</span>
            <span class="text-[10px] px-1.5 py-0.5 rounded bg-zinc-800 text-zinc-300">Chat</span>
          </span>
          <span v-if="!farmContext.farmId" class="block text-amber-300/80 text-[11px] mt-1">Select a farm in the sidebar first.</span>
          <span v-else-if="noGroundedModels" class="block text-amber-300/80 text-[11px] mt-1">
            No grounded models installed — pull phi3:mini in Settings, then refresh.
          </span>
          <span v-else-if="ragWarning" class="block text-amber-300/80 text-[11px] mt-1">Field memories not ingested — run bootstrap.</span>
          <span
            v-if="farmSelected && capabilities.visionChatEnabled"
            class="block text-zinc-500 text-[10px] mt-1.5 leading-snug"
            data-test="guardian-mode-vision-note"
          >
            Zone photos use a separate vision model — first photo question may take extra time on CPU.
          </span>
        </button>
      </div>
      <p class="text-[10px] text-zinc-500">
        First chat after a cold start may take several minutes on CPU — use Quick chat for faster general answers.
      </p>
      <p class="text-[10px] text-zinc-600 leading-snug" data-test="guardian-mode-session-memory-note">
        Session memory uses keyword tags only — not semantic recall.
      </p>
    </template>
  </div>
</template>

<script setup>
import { computed, onMounted, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useCapabilitiesStore } from '../stores/capabilities'
import { useFarmContextStore } from '../stores/farmContext'
import { useGuardianReadinessStore } from '../stores/guardianReadiness'
import { useGuardianModels } from '../composables/useGuardianModels'
import { filterGroundedCapableModels } from '../lib/guardianModelGrounded'

const useFarmContext = defineModel('useFarmContext', { type: Boolean, default: false })

const capabilities = useCapabilitiesStore()
const farmContext = useFarmContextStore()
const readiness = useGuardianReadinessStore()
const { awakening } = storeToRefs(readiness)
const { models, loadModels } = useGuardianModels()

const quickSelected = computed(() => !useFarmContext.value)
const farmSelected = computed(() => useFarmContext.value)
const groundedCapable = computed(() => filterGroundedCapableModels(models.value))
const noGroundedModels = computed(() => models.value.length > 0 && !groundedCapable.value.length)
const farmDisabled = computed(() =>
  !farmContext.farmId || noGroundedModels.value || ollamaUnavailable.value,
)
const ollamaUnavailable = computed(() => awakening.value?.state === 'unavailable')
const ragWarning = computed(() =>
  useFarmContext.value && awakening.value && awakening.value.rag_corpus_ok === false,
)

function selectQuick() {
  useFarmContext.value = false
  void readiness.ensureAwake(null, 'quick')
}

function selectFarmCounsel() {
  if (farmDisabled.value) return
  useFarmContext.value = true
  void readiness.ensureAwake(farmContext.farmId, 'farm_counsel')
}

watch(
  () => useFarmContext.value,
  (on) => {
    if (on && farmContext.farmId && !farmDisabled.value) {
      void readiness.ensureAwake(farmContext.farmId, 'farm_counsel')
    } else if (!on) {
      void readiness.ensureAwake(null, 'quick')
    }
  },
)

watch(noGroundedModels, (blocked) => {
  if (blocked && useFarmContext.value) useFarmContext.value = false
})

onMounted(() => {
  void loadModels()
})
</script>
